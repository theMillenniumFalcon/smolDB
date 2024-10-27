// provides tests for for managing and manipulating a file-based index system
package index

import (
	"os"
	"testing"
)

// initializes the test environment and runs all tests
// sets up a new FileIndex instance and handles test execution cleanup
func TestMain(m *testing.M) {
	I = NewFileIndex("")
	exitVal := m.Run()
	os.Exit(exitVal)
}

// verifies the directory crawling functionality
// Tests the system's ability to identify and process JSON files while ignoring other file types
func TestCrawlDirectory(t *testing.T) {
	// Test Case 1: empty directory handling
	t.Run("crawl empty directory", func(t *testing.T) {
		setup()

		checkDeepEquals(t, crawlDirectory(""), []string{})
	})

	// Test Case 2: multiple JSON files
	t.Run("crawl directory with two files", func(t *testing.T) {
		setup()

		makeNewFile("one.json", "one")
		makeNewFile("two.json", "two")
		checkDeepEquals(t, crawlDirectory(""), []string{"one", "two"})
	})

	// Test Case 3: file type filtering
	t.Run("crawl directory with non json file", func(t *testing.T) {
		setup()

		makeNewFile("one.json", "one")
		makeNewFile("test.txt", "test")
		checkDeepEquals(t, crawlDirectory(""), []string{"one"})
	})
}

// verifies the JSON file to map conversion functionality
// tests various JSON structure scenarios including flat, nested, and array contents
func TestToMap(t *testing.T) {
	// Test Case 1: flat JSON structure
	t.Run("simple flat json to map", func(t *testing.T) {
		setup()
		I.FileSystem.Mkdir("db/", os.ModeAppend)

		expected := map[string]interface{}{
			"field1": "value1",
			"field2": "value2",
		}

		f := makeNewJSON("test", expected)

		got, err := f.ToMap()
		assertNilErr(t, err)
		checkDeepEquals(t, expected, got)
	})

	// Test Case 2: JSON with array elements
	t.Run("json with array", func(t *testing.T) {
		setup()

		expected := map[string]interface{}{
			"array": []string{
				"element1",
				"element2",
			},
		}

		f := makeNewJSON("test", expected)

		got, err := f.ToMap()
		assertNilErr(t, err)
		checkJSONEquals(t, got, expected)
	})

	// Test Case 3: nested JSON structures
	t.Run("deep nested with map", func(t *testing.T) {
		setup()

		expected := map[string]interface{}{
			"array": []interface{}{
				"a",
				map[string]interface{}{
					"test": "deep nest",
				},
			},
		}

		f := makeNewJSON("test", expected)

		got, err := f.ToMap()
		assertNilErr(t, err)
		checkJSONEquals(t, got, expected)
	})
}

// verifies the file content replacement functionality
// tests both updating existing files and creating new files with new content
func TestReplaceContent(t *testing.T) {
	// Test Case 1: update existing file content
	t.Run("update existing file", func(t *testing.T) {
		setup()

		old := map[string]interface{}{
			"field1": "value1",
			"field2": "value2",
		}

		new := map[string]interface{}{
			"field": "value",
		}

		f := makeNewJSON("test", old)
		assertFileExists(t, "test")

		err := f.ReplaceContent(mapToString(new))
		assertNilErr(t, err)

		got, err := f.ToMap()
		assertNilErr(t, err)
		checkJSONEquals(t, got, new)
	})

	// Test Case 2: create and populate new file
	t.Run("create new file", func(t *testing.T) {
		setup()

		new := map[string]interface{}{
			"field": "value",
		}

		f := &File{FileName: "test"}
		assertFileDoesNotExist(t, "test")

		err := f.ReplaceContent(mapToString(new))
		assertNilErr(t, err)
		assertFileExists(t, "test")

		got, err := f.ToMap()
		assertNilErr(t, err)
		checkJSONEquals(t, got, new)
	})
}

// verifies the file deletion functionality
// tests both successful deletion of existing files and handling of non-existent files
func TestDelete(t *testing.T) {
	// Test Case 1: deletion of non-existent file
	t.Run("delete non-existent file", func(t *testing.T) {
		setup()

		f := &File{FileName: "doesnt-exist"}
		assertFileDoesNotExist(t, "doesnt-exist")

		err := f.Delete()
		assertErr(t, err)
		assertFileDoesNotExist(t, "doesnt-exist")
	})

	// Test Case 2: deletion of existing file
	t.Run("delete existing file", func(t *testing.T) {
		setup()

		data := map[string]interface{}{
			"field": "value",
		}

		f := makeNewJSON("test", data)
		assertFileExists(t, "test")

		err := f.Delete()
		assertNilErr(t, err)
		assertFileDoesNotExist(t, "test")
	})
}
