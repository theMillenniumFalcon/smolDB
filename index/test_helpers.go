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

func checkDeepEquals(t *testing.T, a interface{}, b interface{}) {
	t.Helper()
	if !cmp.Equal(a, b) {
		t.Errorf("got %+v, want %+v", a, b)
	}
}

func checkJSONEquals(t *testing.T, a interface{}, b interface{}) {
	t.Helper()
	if fmt.Sprintf("%+v", a) != fmt.Sprintf("%+v", b) {
		t.Errorf("got %+v, want %+v", a, b)
	}
}

func makeNewFile(name string, contents string) {
	af.WriteFile(I.FileSystem, name, []byte(contents), 0644)
}

func makeNewJSON(name string, contents map[string]interface{}) *File {
	jsonData, _ := json.Marshal(contents)
	af.WriteFile(I.FileSystem, name+".json", jsonData, 0644)
	return &File{FileName: name}
}

func mapToString(contents map[string]interface{}) string {
	jsonData, _ := json.Marshal(contents)
	return string(jsonData)
}

func assertNilErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("got error %s when shouldn't have", err.Error())
	}
}

func assertErr(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("didnt get error when wanted error")
	}
}

func assertFileExists(t *testing.T, filePath string) {
	t.Helper()
	if _, err := I.FileSystem.Stat(filePath + ".json"); os.IsNotExist(err) {
		t.Errorf("didnt find file at %s when should have", filePath)
	}
}

func assertFileDoesNotExist(t *testing.T, filePath string) {
	t.Helper()
	if _, err := I.FileSystem.Stat(filePath + ".json"); err == nil {
		t.Errorf("found file at %s when shouldn't have", filePath)
	}
}

func setup() {
	I = NewFileIndex("")
	I.SetFileSystem(af.NewMemMapFs())
}

func createAndReturnFile(t *testing.T, key string) *File {
	t.Helper()

	file := &File{FileName: key}
	err := I.Put(file, []byte("test"))
	if err != nil {
		t.Errorf("err creating file '%s': '%s'", key, err.Error())
	}

	return file
}

func checkKeyNotInIndex(t *testing.T, key string) {
	t.Helper()

	if _, ok := I.index[key]; ok {
		t.Errorf("should not have found key: '%s'", key)
	}
}

func sliceContains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func checkContentEqual(t *testing.T, key string, newContent map[string]interface{}) {
	got, ok := I.Lookup(key)
	assert.True(t, ok)

	gotBytes, err := got.GetByteArray()
	assertNilErr(t, err)
	checkJSONEquals(t, string(gotBytes), mapToString(newContent))
}
