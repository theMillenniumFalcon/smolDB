package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
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

func TestMain(m *testing.M) {
	index.I = index.NewFileIndex(".")
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestGetIndex(t *testing.T) {

	t.Run("get empty index", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		router := httprouter.New()
		router.GET("/getKeys", GetKeys)

		req, _ := http.NewRequest("GET", "/getKeys", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertHTTPBody(t, rr, map[string]interface{}{
			"files": nil,
		})
	})

	t.Run("get index with files", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		expected := map[string]interface{}{
			"field": "value",
		}

		_ = makeNewJSON("test1", expected)
		_ = makeNewJSON("test2", expected)

		index.I.Regenerate()

		router := httprouter.New()
		router.GET("/getKeys", GetKeys)

		req, _ := http.NewRequest("GET", "/getKeys", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertHTTPContains(t, rr, []string{"test1", "test2"})
	})
}

func TestGetKey(t *testing.T) {

	t.Run("get non-existent file", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		router := httprouter.New()
		router.GET("/get/:key", GetKey)

		req, _ := http.NewRequest("GET", "/nothinghere", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusNotFound)
	})

	t.Run("get file", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		expected := map[string]interface{}{
			"field": "value",
		}

		_ = makeNewJSON("test", expected)

		index.I.Regenerate()

		router := httprouter.New()
		router.GET("/get/:key", GetKey)

		req, _ := http.NewRequest("GET", "/get/test", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assertHTTPStatus(t, rr, http.StatusOK)
		assertHTTPBody(t, rr, map[string]interface{}{
			"field": "value",
		})
	})
}
