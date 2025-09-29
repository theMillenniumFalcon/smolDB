package index

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	af "github.com/spf13/afero"
	"github.com/themillenniumfalcon/smolDB/log"
)

// DurabilityLevel controls when we fsync WAL appends
type DurabilityLevel int

const (
	DurabilityNone DurabilityLevel = iota
	DurabilityCommit
	DurabilityGrouped
)

// SyncMode controls how we sync data to disk
type SyncMode int

const (
	SyncNone SyncMode = iota
	SyncFsync
	SyncDSync // best-effort; mapped to fsync for portability
)

// WAL operation kinds
const (
	opPut    = "PUT"
	opDelete = "DELETE"
	opCommit = "COMMIT"
)

// walEntry is the line-delimited JSON format we append to wal.log
type walEntry struct {
	V     int    `json:"v"`
	Op    string `json:"op"`
	Key   string `json:"key"`
	Field string `json:"field,omitempty"`
	Body  string `json:"body,omitempty"`
	Ts    int64  `json:"ts"`
	Csum  uint32 `json:"csum"`
}

// WAL encapsulates write-ahead logging
type WAL struct {
	fs         af.Fs
	file       af.File
	dir        string
	durability DurabilityLevel
	groupMs    int
	groupBatch int
	appendCnt  int
	syncMode   SyncMode
}

func newWAL(fs af.Fs, dir string, durability DurabilityLevel, groupMs int, groupBatch int, syncMode SyncMode) (*WAL, error) {
	walDir := filepath.Join(dir, ".smoldb")
	// ensure .smoldb directory exists
	if err := fs.MkdirAll(walDir, 0o755); err != nil {
		return nil, err
	}
	walPath := filepath.Join(walDir, "wal.log")
	// open append-only
	f, err := fs.OpenFile(walPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	return &WAL{fs: fs, file: f, dir: dir, durability: durability, groupMs: groupMs, groupBatch: groupBatch, syncMode: syncMode}, nil
}

// Append writes one entry to the WAL and optionally fsyncs
func (w *WAL) Append(entry walEntry) error {
	if w == nil || w.file == nil {
		return fmt.Errorf("wal not initialized")
	}
	entry.V = 1
	entry.Ts = time.Now().UnixNano()
	entry.Csum = simpleChecksum(entry)
	bytes, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if _, err = w.file.Write(append(bytes, '\n')); err != nil {
		return err
	}
	switch w.durability {
	case DurabilityCommit:
		w.doSync()
	case DurabilityGrouped:
		// batch-triggered fsync
		if w.groupBatch > 0 {
			w.appendCnt++
			if w.appendCnt%w.groupBatch == 0 {
				w.doSync()
				return nil
			}
		}
		// time-triggered fsync
		if w.groupMs > 0 {
			time.Sleep(time.Duration(w.groupMs) * time.Millisecond)
			w.doSync()
		}
	}
	return nil
}

// doSync performs the actual sync according to syncMode
func (w *WAL) doSync() {
	// write an explicit COMMIT marker to denote a durability boundary
	commit := walEntry{V: 1, Op: opCommit, Ts: time.Now().UnixNano(), Csum: simpleChecksum(walEntry{Op: opCommit})}
	bytes, err := json.Marshal(commit)
	if err == nil {
		_, _ = w.file.Write(append(bytes, '\n'))
	}

	switch w.syncMode {
	case SyncNone:
		// no fsync; commit marker still appended for auditing
		return
	case SyncFsync, SyncDSync:
		_ = w.file.Sync()
	}
}

// Close WAL file handle
func (w *WAL) Close() error {
	if w == nil || w.file == nil {
		return nil
	}
	return w.file.Close()
}

// Replay scans wal.log and re-applies any operations to reach a consistent state
func (w *WAL) Replay(idx *FileIndex) error {
	walPath := filepath.Join(w.dir, ".smoldb", "wal.log")
	f, err := w.fs.OpenFile(walPath, os.O_RDONLY, 0)
	if err != nil {
		// no WAL yet
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var e walEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			log.Warn("wal: skipping malformed line: %s", err.Error())
			break
		}
		file := &File{FileName: e.Key}
		switch e.Op {
		case opPut:
			if err := file.ReplaceContent(e.Body); err != nil {
				log.Warn("wal: put apply failed for key '%s': %s", e.Key, err.Error())
			}
			idx.index[file.FileName] = file
		case opDelete:
			if err := file.Delete(); err != nil {
				log.Warn("wal: delete apply failed for key '%s': %s", e.Key, err.Error())
			}
			delete(idx.index, file.FileName)
		}
	}
	return nil
}

// simpleChecksum computes a lightweight checksum over key/op/body
func simpleChecksum(e walEntry) uint32 {
	const offset32 uint32 = 2166136261
	const prime32 uint32 = 16777619
	sum := offset32
	for _, b := range []byte(e.Op) {
		sum ^= uint32(b)
		sum *= prime32
	}
	for _, b := range []byte(e.Key) {
		sum ^= uint32(b)
		sum *= prime32
	}
	for _, b := range []byte(e.Body) {
		sum ^= uint32(b)
		sum *= prime32
	}
	return sum
}
