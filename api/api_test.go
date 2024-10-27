// provides tests for HTTP handlers and routing logic for the core API
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

// exampleJSON is a simple test fixture used across multiple tests
var exampleJSON = map[string]interface{}{
	"field": "value",
}

// sets up the test environment by initializing a new file index
// and handles proper test cleanup
func TestMain(m *testing.M) {
	index.I = index.NewFileIndex(".")
	exitVal := m.Run()
	os.Exit(exitVal)
}

// verifies the behavior of the Health endpoint
func TestHealth(t *testing.T) {
	router := httprouter.New()
	router.GET("/", Health)

	// Test Case 1: test health checkpoint
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

// verifies that the regenerate endpoint properly rebuilds the file index when called
// it tests that new files are detected after regeneration
func TestRegenerateIndex(t *testing.T) {
	router := httprouter.New()
	router.POST("/regenerate", RegenerateIndex)

	// Test Case 1: index gets regenerated
	t.Run("test regenerate modifies index", func(t *testing.T) {
		// use in-memory filesystem for testing
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

// verifies the behavior of the GetKeys endpoint under different scenarios
func TestGetKeys(t *testing.T) {
	router := httprouter.New()
	router.GET("/getKeys", GetKeys)

	// Test Case 1: when the index is empty
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

	// Test Case 2: when the index contains multiple files
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

// verifies the behavior of the GetKey endpoint
func TestGetKey(t *testing.T) {
	router := httprouter.New()
	router.GET("/:key", GetKey)

	// Test Case 1: when requesting a non-existent file (should return 404)
	t.Run("get non-existent file", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		req, _ := http.NewRequest("GET", "/nothinghere", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusNotFound)
	})

	// Test Case 2: when requesting an existing file (should return file contents)
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

// verifies the behavior of getting specific fields from JSON files
func TestGetKeyField(t *testing.T) {
	router := httprouter.New()
	router.GET("/:key/:field", GetKeyField)

	// Test Case 1: when the key doesn't exist
	t.Run("get field of non-existent key", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		req, _ := http.NewRequest("GET", "/nothinghere1/nothinghere2", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusNotFound)
	})

	// Test Case 2: when the field doesn't exist in the JSON
	t.Run("get non-existent field of key", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		makeNewJSON("test", exampleJSON)

		index.I.Regenerate()

		req, _ := http.NewRequest("GET", "/test/no-field", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusBadRequest)

	})

	// Test Case 3: when getting a simple value field
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

	// Test Case 4: when getting a nested object field
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

// verifies the behavior of updating keys in the database
func TestUpdateKey(t *testing.T) {
	router := httprouter.New()
	router.PUT("/:key", UpdateKey)

	// Test Case 1: when updating a non-existent key (should create it)
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

	// Test Case 2: when updating an existing key (should overwrite it)
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

	// Test Case 3: when updating with non-JSON content (should store as raw bytes)
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

// verifies the behavior of deleting keys
func TestDeleteKey(t *testing.T) {
	router := httprouter.New()
	router.DELETE("/:key", DeleteKey)

	// Test Case 1: when attempting to delete a non-existent key (should return 404)
	t.Run("delete non-existent key", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		req, _ := http.NewRequest("DELETE", "/nothinghere", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusNotFound)
	})

	// Test Case 2: when deleting an existing key (should remove it from index)
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

// verifies the behavior of patching specific fields in JSON files
func TestPatchKeyField(t *testing.T) {
	router := httprouter.New()
	router.PATCH("/:key/:field", PatchKeyField)

	// Test Case 1: when the key doesn't exist
	t.Run("patch field of non-existent key", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		byteReader := mapToIOReader(exampleJSON)

		req, _ := http.NewRequest("PATCH", "/nofile/nofield", byteReader)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusNotFound)
	})

	// Test Case 2: when patching a non-existent field (should add it)
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

	// Test Case 3: when patching with non-JSON content
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
