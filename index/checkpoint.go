package index

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	af "github.com/spf13/afero"
	"github.com/themillenniumfalcon/smolDB/log"
)

// checkpointMeta contains metadata about a checkpoint
type checkpointMeta struct {
	Timestamp int64             `json:"ts"`
	Keys      map[string]string `json:"keys"`      // key -> content hash map
	WalOffset int64             `json:"walOffset"` // offset in WAL file where this checkpoint was taken
}

// CreateCheckpoint creates a new checkpoint file with current state
func (i *FileIndex) CreateCheckpoint() error {
	i.mu.RLock()
	defer i.mu.RUnlock()

	// Create checkpoint directory if it doesn't exist
	checkpointDir := filepath.Join(i.dir, "checkpoint")
	if err := i.FileSystem.MkdirAll(checkpointDir, 0755); err != nil {
		return fmt.Errorf("failed to create checkpoint directory: %v", err)
	}

	// Generate checkpoint filename with timestamp
	ts := time.Now().UnixNano()
	filename := filepath.Join(checkpointDir, fmt.Sprintf("%09d.snap", ts))

	// Create metadata with current state
	meta := &checkpointMeta{
		Timestamp: ts,
		Keys:      make(map[string]string),
		WalOffset: 0, // Will be set after copying files
	}

	// Create a new checkpoint file
	f, err := i.FileSystem.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create checkpoint file: %v", err)
	}
	defer f.Close()

	// Copy all current files to checkpoint
	for key, file := range i.index {
		content, err := file.ReadContent()
		if err != nil {
			log.Warn("checkpoint: failed to read key %s: %v", key, err)
			continue
		}
		meta.Keys[key] = content
	}

	// If WAL exists, get current offset
	if i.wal != nil && i.wal.file != nil {
		if info, err := i.wal.file.Stat(); err == nil {
			meta.WalOffset = info.Size()
		}
	}

	// Write metadata to checkpoint file
	encoder := json.NewEncoder(f)
	if err := encoder.Encode(meta); err != nil {
		return fmt.Errorf("failed to write checkpoint metadata: %v", err)
	}

	// After successful checkpoint creation, truncate WAL
	if i.wal != nil && meta.WalOffset > 0 {
		if err := i.wal.TruncateAt(meta.WalOffset); err != nil {
			log.Warn("checkpoint: failed to truncate WAL: %v", err)
			// Don't fail the checkpoint creation if truncation fails
		}
	}

	return nil
}

// RestoreFromCheckpoint restores the database state from the latest checkpoint
func (i *FileIndex) RestoreFromCheckpoint() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Find latest checkpoint file
	checkpointDir := filepath.Join(i.dir, "checkpoint")
	files, err := af.ReadDir(i.FileSystem, checkpointDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No checkpoints exist
		}
		return fmt.Errorf("failed to read checkpoint directory: %v", err)
	}

	var latestTs int64
	var latestFile string
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		if len(name) < 13 || name[len(name)-5:] != ".snap" { // "000000000.snap"
			continue
		}
		ts, err := strconv.ParseInt(name[:len(name)-5], 10, 64)
		if err != nil {
			continue
		}
		if ts > latestTs {
			latestTs = ts
			latestFile = filepath.Join(checkpointDir, name)
		}
	}

	if latestFile == "" {
		return nil // No valid checkpoints found
	}

	// Read checkpoint file
	f, err := i.FileSystem.Open(latestFile)
	if err != nil {
		return fmt.Errorf("failed to open checkpoint file: %v", err)
	}
	defer f.Close()

	var meta checkpointMeta
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&meta); err != nil {
		return fmt.Errorf("failed to decode checkpoint metadata: %v", err)
	}

	// Clear current index
	i.index = make(map[string]*File)

	// Restore files from checkpoint
	for key, content := range meta.Keys {
		file := &File{FileName: key}
		if err := file.ReplaceContent(content); err != nil {
			log.Warn("checkpoint: failed to restore key %s: %v", key, err)
			continue
		}
		i.index[key] = file
	}

	// Return the WAL offset where we need to resume replay
	if i.wal != nil {
		i.wal.lastCheckpointOffset = meta.WalOffset
	}

	return nil
}
