package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/julienschmidt/httprouter"
	af "github.com/spf13/afero"
	"github.com/themillenniumfalcon/smolDB/index"
)

var exampleJSON = map[string]interface{}{
	"field": "value",
}

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
	router.GET("/:key", GetKey)

	t.Run("get non-existent file", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		req, _ := http.NewRequest("GET", "/nothinghere", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusNotFound)
	})

	t.Run("get file", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		makeNewJSON("test", exampleJSON)

		index.I.Regenerate()

		req, _ := http.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertHTTPBody(t, rr, exampleJSON)
	})
}

func TestGetKeyField(t *testing.T) {
	router := httprouter.New()
	router.GET("/:key/:field", GetKeyField)

	t.Run("get field of non-existent key", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		req, _ := http.NewRequest("GET", "/nothinghere1/nothinghere2", nil)
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

func TestUpdateKey(t *testing.T) {
	router := httprouter.New()
	router.PUT("/:key", UpdateKey)

	t.Run("update non-existent key", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())
		byteReader := mapToIOReader(exampleJSON)

		req, _ := http.NewRequest("PUT", "/something", byteReader)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertSliceContains(t, index.I.ListKeys(), "something")
		assertJSONFileContents(t, index.I, "something", exampleJSON)
	})

	t.Run("update existing key", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())
		shortTest := map[string]interface{}{
			"qwer": "asdf",
		}

		_ = makeNewJSON("something", shortTest)
		index.I.Regenerate()

		assertJSONFileContents(t, index.I, "something", shortTest)
		byteReader := mapToIOReader(exampleJSON)

		req, _ := http.NewRequest("PUT", "/something", byteReader)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertSliceContains(t, index.I.ListKeys(), "something")
		assertJSONFileContents(t, index.I, "something", exampleJSON)
	})

	t.Run("update key with non-json bytes", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		jsonBytes := []byte("non-json bytes")
		byteReader := bytes.NewReader(jsonBytes)

		req, _ := http.NewRequest("PUT", "/something", byteReader)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertSliceContains(t, index.I.ListKeys(), "something")
		assertRawFileContents(t, index.I, "something", jsonBytes)
	})
}

func TestDeleteKey(t *testing.T) {
	router := httprouter.New()
	router.DELETE("/:key", DeleteKey)

	t.Run("delete non-existent key", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		req, _ := http.NewRequest("DELETE", "/nothinghere", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusNotFound)
	})

	t.Run("delete existing key", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())
		_ = makeNewJSON("test", exampleJSON)
		index.I.Regenerate()

		req, _ := http.NewRequest("DELETE", "/test", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertEmptySlice(t, index.I.ListKeys())
	})
}

func TestPatchKeyField(t *testing.T) {
	router := httprouter.New()
	router.PATCH("/:key/:field", PatchKeyField)

	t.Run("patch field of non-existent key", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		byteReader := mapToIOReader(exampleJSON)

		req, _ := http.NewRequest("PATCH", "/nofile/nofield", byteReader)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusNotFound)
	})

	t.Run("patch non-existent field of existing key", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		_ = makeNewJSON("test", exampleJSON)
		index.I.Regenerate()

		byteReader := mapToIOReader(exampleJSON)

		req, _ := http.NewRequest("PATCH", "/test/nofield", byteReader)
		rr := httptest.NewRecorder()

		expected := map[string]interface{}{
			"field":   "value",
			"nofield": exampleJSON,
		}

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertJSONFileContents(t, index.I, "test", expected)
	})

	t.Run("patch field of existing key with non-json bytes", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		_ = makeNewJSON("test", exampleJSON)
		index.I.Regenerate()

		jsonBytes := []byte("non-json bytes")
		byteReader := bytes.NewReader(jsonBytes)

		req, _ := http.NewRequest("PATCH", "/test/field", byteReader)
		rr := httptest.NewRecorder()

		expected := map[string]interface{}{
			"field": "non-json bytes",
		}

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertJSONFileContents(t, index.I, "test", expected)
	})
}
