package index

import "sync"

type File struct {
	FileName string
	mu       sync.RWMutex
}

type FileIndex struct {
	mu    sync.RWMutex
	dir   string
	index map[string]File
}

var I *FileIndex

func init() {
	I = &FileIndex{
		dir:   "",
		index: map[string]File{},
	}
}

func (i FileIndex) Lookup(key string) (File, bool) {
	i.mu.RLock()
	defer i.mu.Unlock()

	if file, ok := i.index[key]; ok {
		return file, true
	}

	return File{}, false
}

func (i FileIndex) Regenerate(dir string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.dir = dir
	i.index = i.buildIndexMap()
}

func (i FileIndex) buildIndexMap() map[string]File {
	newIndexMap := make(map[string]File)

	files := crawlDirectory(i.dir)
	for _, f := range files {
		newIndexMap[f] = File{FileName: f}
	}

	return newIndexMap
}
