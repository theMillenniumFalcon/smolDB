package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/julienschmidt/httprouter"
	af "github.com/spf13/afero"
	"github.com/themillenniumfalcon/smolDB/index"
)

func assertHTTPStatus(t *testing.T, rr *httptest.ResponseRecorder, status int) {
	t.Helper()
	got := rr.Code
	if got != http.StatusOK {
		t.Errorf("returned wrong status code: got %+v, wanted %+v", got, status)
	}
}

func assertHTTPBody(t *testing.T, rr *httptest.ResponseRecorder, expected string) {
	t.Helper()
	if rr.Body.String() != expected {
		t.Errorf("returned unexpected body: got %+v want %+v", rr.Body.String(), expected)
	}
}

func makeNewJSON(name string, contents map[string]interface{}) *index.File {
	jsonData, _ := json.Marshal(contents)
	af.WriteFile(index.I.FileSystem, name+".json", jsonData, 0644)
	return &index.File{FileName: name}
}

func TestMain(m *testing.M) {
	index.I = index.NewFileIndex("")
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
		assertHTTPBody(t, rr, `{"files":null}`)
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
		assertHTTPBody(t, rr, `{"files":["test1","test2"]}`)
	})
}
