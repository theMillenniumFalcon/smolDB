// provides testing utilities for API tests
package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	af "github.com/spf13/afero"
	"github.com/themillenniumfalcon/smolDB/index"
)

// checks if the HTTP response status code matches the expected status
// fails the test with a descriptive error message if they don't match
func assertHTTPStatus(t *testing.T, rr *httptest.ResponseRecorder, status int) {
	t.Helper()
	got := rr.Code
	if got != status {
		t.Errorf("returned wrong status code: got %+v, wanted %+v", got, status)
	}
}

// verifies that the HTTP response body contains all the expected strings
// fails the test if any expected string is not found in the response body
func assertHTTPContains(t *testing.T, rr *httptest.ResponseRecorder, expected []string) {
	t.Helper()
	for _, v := range expected {
		if !strings.Contains(rr.Body.String(), v) {
			t.Errorf("couldn't find %s in body %+v", v, rr.Body.String())
		}
	}
}

// checks if a string slice contains a specific string
// fails the test if the string is not found in the slice
func assertSliceContains(t *testing.T, list []string, s string) {
	found := false
	for _, v := range list {
		if v == s {
			found = true
		}
	}
	if !found {
		t.Errorf("slice %+v didn't contain %s", list, s)
	}
}

// verifies that a string slice is empty
// fails the test if the slice has any elements
func assertEmptySlice(t *testing.T, list []string) {
	if len(list) != 0 {
		t.Errorf("slice %+v wasnt empty", list)
	}
}

// verifies that the HTTP response body matches the expected JSON structure.
func assertHTTPBody(t *testing.T, rr *httptest.ResponseRecorder, expected map[string]interface{}) {
	t.Helper()
	resp := rr.Result()
	body, _ := io.ReadAll(resp.Body)

	var parsedJSON map[string]interface{}
	err := json.Unmarshal(body, &parsedJSON)

	if err != nil {
		t.Errorf("got an error parsing json when shouldn't have")
	}

	// test will fail if parsed JSON doesn't match the expected map structure
	if !reflect.DeepEqual(parsedJSON, expected) {
		t.Errorf("json mismatched, got: %+v, want: %+v", parsedJSON, expected)
	}
}

// creates a new JSON file in the test filesystem with the given name and contents
// returns a File struct representing the created file
func makeNewJSON(name string, contents map[string]interface{}) *index.File {
	jsonData, _ := json.Marshal(contents)
	// file is created with 0644 permissions and a .json extension is automatically added
	af.WriteFile(index.I.FileSystem, name+".json", jsonData, 0644)
	return &index.File{FileName: name}
}

// verifies that a file in the index contains the expected JSON content
func assertJSONFileContents(t *testing.T, ind *index.FileIndex, key string, wanted map[string]interface{}) {
	f, ok := ind.Lookup(key)
	if !ok {
		t.Errorf("couldn't find key %s in index", key)
	}

	// test will fail if file content cannot be parsed as JSON
	m, err := f.ToMap()
	if err != nil {
		t.Errorf("got error %+v parsing json when shouldn't have", err.Error())
	}

	// test will fail if parsed JSON doesn't match the expected structure
	if !cmp.Equal(m, wanted) {
		t.Errorf("file content %+v didn't match! wanted %+v", m, wanted)
	}
}

// verifies that a file in the index contains the expected raw byte content
func assertRawFileContents(t *testing.T, ind *index.FileIndex, key string, wanted []byte) {
	f, ok := ind.Lookup(key)
	if !ok {
		t.Errorf("couldn't find key %s in index", key)
	}

	// test will fail if file content doesn't match the expected bytes
	b, _ := f.GetByteArray()
	if !cmp.Equal(b, wanted) {
		t.Errorf("file content %+v didn't match! wanted %+v", string(b), string(wanted))
	}
}

// converts a map to an io.Reader containing its JSON representation
// useful for creating request bodies in tests
func mapToIOReader(m map[string]interface{}) io.Reader {
	jsonData, _ := json.Marshal(m)
	return bytes.NewReader(jsonData)
}
