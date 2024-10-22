package index

import (
	"fmt"
	"sync"
	"time"

	"github.com/themillenniumfalcon/smolDB/log"
)

type File struct {
	FileName string
	mu       sync.RWMutex
}

type FileIndex struct {
	mu    sync.RWMutex
	Dir   string
	index map[string]*File
}

var I *FileIndex

func init() {
	I = &FileIndex{
		Dir:   "",
		index: map[string]*File{},
	}
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
	i.index[file.FileName] = file
	i.mu.Unlock()
	return file.replaceContent(string(bytes))
}

func (i *FileIndex) Regenerate(dir string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	start := time.Now()
	log.Info("building index for directory %s...", dir)

	i.Dir = dir
	i.index = i.buildIndexMap()
	log.Info("built index of %d files in %d ms", len(i.index), time.Since(start).Milliseconds())
}

func (i *FileIndex) buildIndexMap() map[string]*File {
	newIndexMap := make(map[string]*File)

	files := crawlDirectory(i.Dir)
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
	return fmt.Sprintf("%s/%s.json", I.Dir, f.FileName)
}
