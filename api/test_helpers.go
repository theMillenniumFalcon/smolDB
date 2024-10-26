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

func assertHTTPStatus(t *testing.T, rr *httptest.ResponseRecorder, status int) {
	t.Helper()
	got := rr.Code
	if got != status {
		t.Errorf("returned wrong status code: got %+v, wanted %+v", got, status)
	}
}

func assertHTTPContains(t *testing.T, rr *httptest.ResponseRecorder, expected []string) {
	t.Helper()
	for _, v := range expected {
		if !strings.Contains(rr.Body.String(), v) {
			t.Errorf("couldn't find %s in body %+v", v, rr.Body.String())
		}
	}
}

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
func assertEmptySlice(t *testing.T, list []string) {
	if len(list) != 0 {
		t.Errorf("slice %+v wasnt empty", list)
	}
}

func assertHTTPBody(t *testing.T, rr *httptest.ResponseRecorder, expected map[string]interface{}) {
	t.Helper()
	resp := rr.Result()
	body, _ := io.ReadAll(resp.Body)

	var parsedJSON map[string]interface{}
	err := json.Unmarshal(body, &parsedJSON)

	if err != nil {
		t.Errorf("got an error parsing json when shouldn't have")
	}

	if !reflect.DeepEqual(parsedJSON, expected) {
		t.Errorf("json mismatched, got: %+v, want: %+v", parsedJSON, expected)
	}
}

func makeNewJSON(name string, contents map[string]interface{}) *index.File {
	jsonData, _ := json.Marshal(contents)
	af.WriteFile(index.I.FileSystem, name+".json", jsonData, 0644)
	return &index.File{FileName: name}
}

func assertJSONFileContents(t *testing.T, ind *index.FileIndex, key string, wanted map[string]interface{}) {
	f, ok := ind.Lookup(key)
	if !ok {
		t.Errorf("couldn't find key %s in index", key)
	}

	m, err := f.ToMap()
	if err != nil {
		t.Errorf("got error %+v parsing json when shouldn't have", err.Error())
	}

	if !cmp.Equal(m, wanted) {
		t.Errorf("file content %+v didn't match! wanted %+v", m, wanted)
	}
}

func assertRawFileContents(t *testing.T, ind *index.FileIndex, key string, wanted []byte) {
	f, ok := ind.Lookup(key)
	if !ok {
		t.Errorf("couldn't find key %s in index", key)
	}

	b, _ := f.GetByteArray()
	if !cmp.Equal(b, wanted) {
		t.Errorf("file content %+v didn't match! wanted %+v", string(b), string(wanted))
	}
}

func mapToIOReader(m map[string]interface{}) io.Reader {
	jsonData, _ := json.Marshal(m)
	return bytes.NewReader(jsonData)
}
