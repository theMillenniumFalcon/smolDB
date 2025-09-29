// provides the core file indexing and management functionality for smolDB,
// implements thread-safe operations for managing JSON files in the database
package index

import (
	"fmt"
	"sync"
	"time"

	af "github.com/spf13/afero"
	"github.com/themillenniumfalcon/smolDB/log"
)

// file represents a single JSON file in the database with thread-safe operations
type File struct {
	FileName string       // name of the file without extension
	mu       sync.RWMutex // mutex for thread-safe file operations
}

// the main index structure that manages all files in the database
type FileIndex struct {
	mu         sync.RWMutex     // mutex for thread-safe index operations
	dir        string           // base directory for database files
	index      map[string]*File // map of filename to File objects
	FileSystem af.Fs            // abstract filesystem interface for testing and flexibility
	wal        *WAL             // write-ahead log for durability
	durability DurabilityLevel  // durability level for fsync behavior
	groupBatch int              // fsync after this many appends when grouped
	syncMode   SyncMode         // sync mode for WAL
}

// global instance of FileIndex used throughout the application
var I *FileIndex

// creates a new FileIndex instance with the specified directory
// and initializes it with an empty index map and OS filesystem
func NewFileIndex(dir string) *FileIndex {
	return &FileIndex{
		dir:        dir,
		index:      map[string]*File{},
		FileSystem: af.NewOsFs(),
	}
}

// InitWAL initializes the WAL with the given durability level
func (i *FileIndex) InitWAL(level DurabilityLevel) error {
	i.durability = level
	w, err := newWAL(i.FileSystem, i.dir, level, 0, 0, SyncFsync)
	if err != nil {
		return err
	}
	i.wal = w
	return nil
}

// InitWALWithOptions initializes the WAL with durability and group commit interval/batch
func (i *FileIndex) InitWALWithOptions(level DurabilityLevel, groupCommitMs int, groupCommitBatch int) error {
	i.durability = level
	i.groupBatch = groupCommitBatch
	w, err := newWAL(i.FileSystem, i.dir, level, groupCommitMs, groupCommitBatch, i.syncMode)
	if err != nil {
		return err
	}
	i.wal = w
	return nil
}

// SetSyncMode sets sync mode for WAL
func (i *FileIndex) SetSyncMode(mode SyncMode) {
	i.syncMode = mode
	if i.wal != nil {
		i.wal.syncMode = mode
	}
}

// allows injection of a different filesystem implementation,
// primarily used for testing purposes
func (i *FileIndex) SetFileSystem(fs af.Fs) {
	i.FileSystem = fs
}

// retrieves a File object from the index by its key,
// returns the File and true if found, a new File and false if not found
// thread-safe through read lock
func (i *FileIndex) Lookup(key string) (*File, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	if file, ok := i.index[key]; ok {
		return file, true
	}

	return &File{FileName: key}, false
}

// put adds or updates a file in the index with the provided content
// thread-safe through write lock
func (i *FileIndex) Put(file *File, bytes []byte) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.index[file.FileName] = file
	// append to WAL before applying mutation
	if i.wal != nil {
		_ = i.wal.Append(walEntry{Op: opPut, Key: file.FileName, Body: string(bytes)})
	}
	err := file.ReplaceContent(string(bytes))
	return err
}

// rebuilds the entire index by scanning the database directory
// thread-safe through write lock
func (i *FileIndex) Regenerate() {
	i.mu.Lock()
	defer i.mu.Unlock()

	start := time.Now()
	log.Info("building index for directory %s...", I.dir)

	i.index = i.buildIndexMap()
	log.Success("built index of %d files in %d ms", len(i.index), time.Since(start).Milliseconds())
}

// changes the database directory and regenerates the index,
// used when switching to a different database directory
func (i *FileIndex) RegenerateNew(dir string) {
	i.dir = dir
	i.Regenerate()
}

// creates a new index map by scanning the database directory
func (i *FileIndex) buildIndexMap() map[string]*File {
	newIndexMap := make(map[string]*File)

	files := crawlDirectory(i.dir)
	for _, f := range files {
		newIndexMap[f] = &File{FileName: f}
	}

	return newIndexMap
}

// removes a file from both the filesystem and the index
// thread-safe through write lock
func (i *FileIndex) Delete(file *File) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	// append to WAL before applying mutation
	if i.wal != nil {
		_ = i.wal.Append(walEntry{Op: opDelete, Key: file.FileName})
	}
	err := file.Delete()
	if err == nil {
		delete(i.index, file.FileName)
	}

	return err
}

// WALAvailable reports whether WAL is initialized
func (i *FileIndex) WALAvailable() bool {
	return i != nil && i.wal != nil
}

// WALReplay replays the WAL to bring files and index to a consistent state
func (i *FileIndex) WALReplay() error {
	if i == nil || i.wal == nil {
		return nil
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.wal.Replay(i)
}

// returns a slice of all keys (filenames) in the index
// thread-safe through read lock
func (i *FileIndex) ListKeys() (res []string) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	for k := range i.index {
		res = append(res, k)
	}

	return res
}

// returns the full filesystem path for a file
// handles both root directory and subdirectory cases
func (f *File) ResolvePath() string {
	if I.dir == "" {
		return fmt.Sprintf("%s.json", f.FileName)
	}

	return fmt.Sprintf("%s/%s.json", I.dir, f.FileName)
}
