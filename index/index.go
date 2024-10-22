package index

import (
	"fmt"
	"os"
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

	return &File{FileName: key}, false
}

func (i *FileIndex) Regenerate(dir string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	start := time.Now()
	log.Infof("building index for directory %s...", dir)

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

func (f *File) ReplaceContent(str string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	os.Create(f.ResolvePath())
	file, err := os.OpenFile(f.ResolvePath(), os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, e := file.WriteString(str)
	if e != nil {
		log.Fatal(err)
	}
}
