package index

import (
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
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

	return &File{}, false
}

func (i *FileIndex) Regenerate(dir string) {
	start := time.Now()
	log.Infof("building index for directory %s...", dir)
	i.mu.Lock()
	defer i.mu.Unlock()

	i.Dir = dir
	i.index = i.buildIndexMap()
	log.Infof("built index in %d ms", time.Since(start).Milliseconds())
}

func (i *FileIndex) buildIndexMap() map[string]*File {
	newIndexMap := make(map[string]*File)

	files := crawlDirectory(i.Dir)
	for _, f := range files {
		newIndexMap[f] = &File{FileName: f}
	}

	return newIndexMap
}

func (i *FileIndex) ListKeys() string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	var res []string
	for k := range i.index {
		res = append(res, k)
	}

	return strings.Join(res, ", ")
}
