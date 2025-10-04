// provides file system operations for a simple JSON-based database
// implements thread-safe operations to ensure thread safety
package index

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	af "github.com/spf13/afero"
	"github.com/themillenniumfalcon/smolDB/log"
)

// scans a directory and returns a list of JSON file names without their extension,
// filters for .json files only and returns their base names
func crawlDirectory(directory string) []string {
	files, err := af.ReadDir(I.FileSystem, directory)
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

// replaces the entire content of a file with the provided string,
// uses mutex locking to ensure thread safety
func (f *File) ReplaceContent(str string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// create (or truncate) the file
	_, err := I.FileSystem.Create(f.ResolvePath())
	if err != nil {
		return err
	}

	// open the file for writing
	file, err := I.FileSystem.OpenFile(f.ResolvePath(), os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}

	defer file.Close()

	// write the new content
	_, err = file.WriteString(str)
	if err != nil {
		return err
	}

	// Update metadata with new checksum
	meta := &MetaData{
		Checksum: calculateChecksum([]byte(str)),
		Modified: time.Now().UTC().Format(time.RFC3339),
	}

	// Read existing metadata if it exists
	existing, _ := f.readMetadata()
	if existing != nil {
		meta.Created = existing.Created
	} else {
		meta.Created = meta.Modified
	}

	err = f.writeMetadata(meta)
	if err != nil {
		return fmt.Errorf("failed to update metadata: %v", err)
	}

	return nil
}

// removes the file from the filesystem
// uses mutex locking to ensure thread safety
func (f *File) Delete() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	err := I.FileSystem.Remove(f.ResolvePath())
	if err != nil {
		return err
	}

	return nil
}

// reads the entire file content and returns it as a byte slice
// uses read lock to allow concurrent reads
func (f *File) GetByteArray() ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return af.ReadFile(I.FileSystem, f.ResolvePath())
}

// reads the file content and unmarshals it into a map
// expects the file content to be valid JSON
func (f *File) ToMap() (res map[string]interface{}, err error) {
	bytes, err := f.GetByteArray()
	if err != nil {
		return res, err
	}

	// parse the JSON content into a map
	err = json.Unmarshal(bytes, &res)
	return res, err
}
