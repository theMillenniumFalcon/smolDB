// provides testing utilities
package index

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	af "github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

// performs a deep equality comparison between two interfaces
// uses google/go-cmp for structural equality checking
func checkDeepEquals(t *testing.T, a interface{}, b interface{}) {
	t.Helper()
	if !cmp.Equal(a, b) {
		t.Errorf("got %+v, want %+v", a, b)
	}
}

// compares two interfaces by converting them to string representations
// useful for comparing JSON structures where order might not matter
func checkJSONEquals(t *testing.T, a interface{}, b interface{}) {
	t.Helper()
	if fmt.Sprintf("%+v", a) != fmt.Sprintf("%+v", b) {
		t.Errorf("got %+v, want %+v", a, b)
	}
}

// creates a new file in the virtual filesystem with given contents
// uses default permissions (0644) for file creation
func makeNewFile(name string, contents string) {
	af.WriteFile(I.FileSystem, name, []byte(contents), 0644)
}

// creates a new JSON file from a map and returns a File struct
// uses default permissions (0644) for file creation
func makeNewJSON(name string, contents map[string]interface{}) *File {
	jsonData, _ := json.Marshal(contents)
	af.WriteFile(I.FileSystem, name+".json", jsonData, 0644)
	return &File{FileName: name}
}

// converts a map to its JSON string representation
func mapToString(contents map[string]interface{}) string {
	jsonData, _ := json.Marshal(contents)
	return string(jsonData)
}

// verifies that no error occurred
// fails the test if an error is present
func assertNilErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("got error %s when shouldn't have", err.Error())
	}
}

// verifies that an error occurred
// fails the test if no error is present
func assertErr(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("didnt get error when wanted error")
	}
}

// checks if a file exists in the filesystem
func assertFileExists(t *testing.T, filePath string) {
	t.Helper()
	if _, err := I.FileSystem.Stat(filePath + ".json"); os.IsNotExist(err) {
		t.Errorf("didnt find file at %s when should have", filePath)
	}
}

// verifies that a file does not exist in the filesystem
func assertFileDoesNotExist(t *testing.T, filePath string) {
	t.Helper()
	if _, err := I.FileSystem.Stat(filePath + ".json"); err == nil {
		t.Errorf("found file at %s when shouldn't have", filePath)
	}
}

// initializes a new file index with an in-memory filesystem
// should be called at the start of each test that needs a fresh filesystem
func setup() {
	I = NewFileIndex("")
	I.SetFileSystem(af.NewMemMapFs())
}

// creates a new file with test content and returns the File struct
// fails the test if file creation fails
func createAndReturnFile(t *testing.T, key string) *File {
	t.Helper()

	file := &File{FileName: key}
	err := I.Put(file, []byte("test"))
	if err != nil {
		t.Errorf("err creating file '%s': '%s'", key, err.Error())
	}

	return file
}

// verifies that a key is not present in the index
// fails the test if the key is found
func checkKeyNotInIndex(t *testing.T, key string) {
	t.Helper()

	if _, ok := I.index[key]; ok {
		t.Errorf("should not have found key: '%s'", key)
	}
}

// checks if a string is present in a slice
// returns true if found, false otherwise
func sliceContains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// verifies that the content of a file matches the expected map
// uses JSON string comparison for equality checking
func checkContentEqual(t *testing.T, key string, newContent map[string]interface{}) {
	got, ok := I.Lookup(key)
	assert.True(t, ok)

	gotBytes, err := got.GetByteArray()
	assertNilErr(t, err)
	checkJSONEquals(t, string(gotBytes), mapToString(newContent))
}
