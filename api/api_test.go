package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/julienschmidt/httprouter"
	af "github.com/spf13/afero"
	"github.com/themillenniumfalcon/smolDB/index"
)

func TestMain(m *testing.M) {
	index.I = index.NewFileIndex(".")
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestHealth(t *testing.T) {
	router := httprouter.New()
	router.GET("/", Health)

	t.Run("check health endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assertHTTPStatus(t, rr, http.StatusOK)
		assertHTTPBody(t, rr, map[string]interface{}{
			"message": "smolDB is working fine!",
		})
	})
}

func TestGetKeys(t *testing.T) {
	router := httprouter.New()
	router.GET("/getKeys", GetKeys)

	t.Run("get empty index", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

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

		_ = makeNewJSON("test1", exampleJSON)
		_ = makeNewJSON("test2", exampleJSON)

		index.I.Regenerate()

		req, _ := http.NewRequest("GET", "/getKeys", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertHTTPContains(t, rr, []string{"test1", "test2"})
	})
}

func TestGetKey(t *testing.T) {
	router := httprouter.New()
	router.GET("/get/:key", GetKey)

	t.Run("get non-existent file", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		req, _ := http.NewRequest("GET", "/nothinghere1/nothinghere1", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusNotFound)
	})

	t.Run("get file", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		makeNewJSON("test", exampleJSON)

		index.I.Regenerate()

		req, _ := http.NewRequest("GET", "/get/test", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertHTTPBody(t, rr, map[string]interface{}{
			"field": "value",
		})
	})
}

func TestRegenerateIndex(t *testing.T) {
	router := httprouter.New()
	router.POST("/regenerate", RegenerateIndex)

	t.Run("test regenerate modifies index", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		index.I.Regenerate()

		makeNewJSON("test", exampleJSON)
		assertEmptySlice(t, index.I.ListKeys())

		req, _ := http.NewRequest("POST", "/regenerate", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)

		assertSliceContains(t, index.I.ListKeys(), "test")
	})
}

func TestGetKeyField(t *testing.T) {
	router := httprouter.New()
	router.GET("/:key/:field", GetKeyField)

	t.Run("get field of non-existent key", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		req, _ := http.NewRequest("GET", "/nothinghere1/nothinghere2/nothinghere3", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusNotFound)
	})

	t.Run("get non-existent field of key", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		makeNewJSON("test", exampleJSON)

		index.I.Regenerate()

		req, _ := http.NewRequest("GET", "/test/no-field", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusBadRequest)

	})

	t.Run("get field of key simple value", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		_ = makeNewJSON("test", exampleJSON)

		index.I.Regenerate()

		req, _ := http.NewRequest("GET", "/test/field", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertHTTPContains(t, rr, []string{"value"})
	})

	t.Run("get field of key", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		nested := map[string]interface{}{
			"more_fields": "yay",
			"nested_thing": map[string]interface{}{
				"f": "asdf",
			},
		}

		expected := map[string]interface{}{
			"field":       nested,
			"other_field": "yeet",
		}

		_ = makeNewJSON("test", expected)

		index.I.Regenerate()

		req, _ := http.NewRequest("GET", "/test/field", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertHTTPBody(t, rr, nested)
	})
}
