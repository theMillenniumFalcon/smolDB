package index

import (
	"fmt"
	"sync"
	"time"

	"github.com/themillenniumfalcon/smolDB/log"

	af "github.com/spf13/afero"
)

type File struct {
	FileName string
	mu       sync.RWMutex
}

type FileIndex struct {
	mu         sync.RWMutex
	dir        string
	index      map[string]*File
	FileSystem af.Fs
}

var I *FileIndex

func NewFileIndex(dir string) *FileIndex {
	return &FileIndex{
		dir:        dir,
		index:      map[string]*File{},
		FileSystem: af.NewOsFs(),
	}
}

func (i *FileIndex) SetFileSystem(fs af.Fs) {
	i.FileSystem = fs
}

func (i *FileIndex) Lookup(key string) (*File, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	if file, ok := i.index[key]; ok {
		return file, true
	}

	return &File{FileName: key}, false
}

func (i *FileIndex) Put(file *File, bytes []byte) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.index[file.FileName] = file
	err := file.ReplaceContent(string(bytes))
	return err
}

func (i *FileIndex) Regenerate() {
	i.mu.Lock()
	defer i.mu.Unlock()

	start := time.Now()
	log.Info("building index for directory %s...", I.dir)

	i.index = i.buildIndexMap()
	log.Info("built index of %d files in %d ms", len(i.index), time.Since(start).Milliseconds())
}

func (i *FileIndex) RegenerateNew(dir string) {
	i.dir = dir
	i.Regenerate()
}

func (i *FileIndex) buildIndexMap() map[string]*File {
	newIndexMap := make(map[string]*File)

	files := crawlDirectory(i.dir)
	for _, f := range files {
		newIndexMap[f] = &File{FileName: f}
	}

	return newIndexMap
}

func (i *FileIndex) Delete(file *File) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	err := file.Delete()
	if err == nil {
		delete(i.index, file.FileName)
	}

	return err
}

func (i *FileIndex) ListKeys() (res []string) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	for k := range i.index {
		res = append(res, k)
	}

	return res
}

func (f *File) ResolvePath() string {
	if I.dir == "" {
		return fmt.Sprintf("%s.json", f.FileName)
	}

	return fmt.Sprintf("%s/%s.json", I.dir, f.FileName)
}
