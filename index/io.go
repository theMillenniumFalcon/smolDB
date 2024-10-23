package index

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/themillenniumfalcon/smolDB/log"
)

func crawlDirectory(directory string) []string {
	files, err := os.ReadDir(directory)
	if err != nil {
		log.Fatal(err)
	}

	res := []string{}

	for _, file := range files {
		ext := filepath.Ext(file.Name())
		if ext == ".json" {
			name := strings.TrimSuffix(file.Name(), ".json")
			res = append(res, name)
		}
	}

	return res
}

func (f *File) ReplaceContent(str string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	_, err := os.Create(f.ResolvePath())
	if err != nil {
		return err
	}

	file, err := os.OpenFile(f.ResolvePath(), os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}

	defer file.Close()

	_, e := file.WriteString(str)
	if e != nil {
		return err
	}

	return nil
}

func (f *File) Delete() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	err := os.Remove(f.ResolvePath())
	if err != nil {
		return err
	}

	return nil
}

func (f *File) getByteArray() ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return os.ReadFile(f.ResolvePath())
}

func (f *File) ToMap() (res map[string]interface{}, err error) {
	bytes, err := f.getByteArray()
	if err != nil {
		return res, err
	}

	err = json.Unmarshal(bytes, &res)
	return res, err
}
