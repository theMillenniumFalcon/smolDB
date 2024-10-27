// provides tests for utilities for managing file indices and content manipulation
package index

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// tests the file path resolution functionality ensuring correct path
// construction with and without directory prefixes
func TestFile_ResolvePath(t *testing.T) {
	// Test Case 1: path resolution without directory
	t.Run("test path correct with empty dir", func(t *testing.T) {
		setup()

		file := createAndReturnFile(t, "resolve_empty")
		got := file.ResolvePath()
		want := "resolve_empty.json"

		checkDeepEquals(t, got, want)
	})

	// Test Case 2: path resolution with directory prefix
	t.Run("file path correct with directories", func(t *testing.T) {
		setup()

		I.dir = "db"
		file := createAndReturnFile(t, "resolve_test")

		got := file.ResolvePath()
		want := "db/resolve_test.json"

		checkDeepEquals(t, got, want)
	})
}

// tests the file lookup functionality verifying both successful
// retrievals and handling of missing files
func TestFileIndex_Lookup(t *testing.T) {
	// Test Case 1: successful file lookup
	t.Run("lookup existing file", func(t *testing.T) {
		setup()

		key := "lookup"
		createAndReturnFile(t, key)

		file, ok := I.Lookup(key)
		if !ok {
			t.Errorf("should have found file: '%s'", file.FileName)
		}

		bytes, _ := file.GetByteArray()
		checkDeepEquals(t, string(bytes), "test")
	})

	// Test Case 2: lookup of non-existent file
	t.Run("lookup non-existent file", func(t *testing.T) {
		setup()

		file, ok := I.Lookup("doesnt_exist")
		if ok {
			t.Errorf("should not have found file: '%s'", file.FileName)
		}
	})
}

// tests the file deletion functionality ensuring proper handling
// of both existing and non-existent files
func TestFileIndex_Delete(t *testing.T) {
	// Test Case 1: deletion of existing file
	t.Run("delete file that exists", func(t *testing.T) {
		setup()

		key := "delete_test1"
		file := createAndReturnFile(t, key)
		err := I.Delete(file)
		assertNilErr(t, err)

		checkKeyNotInIndex(t, key)
	})

	// Test Case 2: deletion of non-existent file
	t.Run("delete file that does not exist", func(t *testing.T) {
		setup()

		key := "doesnt_exist"
		file := &File{FileName: "doesnt-exist"}
		assertFileDoesNotExist(t, "doesnt-exist")

		err := I.Delete(file)
		assertErr(t, err)

		checkKeyNotInIndex(t, key)
	})
}

// tests the directory listing functionality verifying correct behavior
// for both empty and populated directories
func TestFileIndex_List(t *testing.T) {
	// Test Case 1: listing empty directory
	t.Run("list empty dir", func(t *testing.T) {
		setup()

		list := I.ListKeys()
		checkDeepEquals(t, len(list), 0)
	})

	// Test Case 2: listing directory with multiple files
	t.Run("list dir with two files", func(t *testing.T) {
		setup()

		createAndReturnFile(t, "list1")
		createAndReturnFile(t, "list2")

		assert.True(t, sliceContains(I.ListKeys(), "list1"))
		assert.True(t, sliceContains(I.ListKeys(), "list2"))
	})
}

// tests the index regeneration functionality ensuring proper updating
// of the index when files are added or modified
func TestFileIndex_Regenerate(t *testing.T) {
	// Test Case 1: index update with new files
	t.Run("test if new files are added to current index", func(t *testing.T) {
		setup()

		makeNewFile("regenerate1.json", "test")
		makeNewFile("regenerate2.json", "test")

		checkDeepEquals(t, len(I.ListKeys()), 0)

		I.Regenerate()

		assert.True(t, sliceContains(I.ListKeys(), "regenerate1"))
		assert.True(t, sliceContains(I.ListKeys(), "regenerate2"))
	})

	// Test Case 2: selective directory regeneration
	t.Run("test RegenerateNew correctly updates index with files in directory", func(t *testing.T) {
		setup()

		// in . not db
		makeNewFile("regenerate_new.json", "test")

		// in db
		makeNewFile("db/regenerate_new_db.json", "test")

		checkDeepEquals(t, len(I.ListKeys()), 0)

		I.RegenerateNew("db")

		checkDeepEquals(t, I.ListKeys(), []string{"regenerate_new_db"})
		checkDeepEquals(t, I.dir, "db")
	})
}

// tests the file creation and update functionality verifying proper handling
// of content for both new and existing files
func TestFileIndex_Put(t *testing.T) {
	content := map[string]interface{}{
		"array": []interface{}{
			"a",
			map[string]interface{}{
				"test": "deep nest",
			},
		},
	}

	// Test Case 1: creating new file with content
	t.Run("create empty file with given content", func(t *testing.T) {
		setup()

		key := "put_empty"
		file := &File{FileName: key}
		assertFileDoesNotExist(t, key)

		bytes, _ := json.Marshal(content)
		err := I.Put(file, bytes)
		assertNilErr(t, err)
		assertFileExists(t, key)

		checkContentEqual(t, key, content)
	})

	// Test Case 2: updating existing file with new content
	t.Run("replace existing file with given content", func(t *testing.T) {
		setup()

		newContent := map[string]interface{}{
			"field": "value",
		}

		key := "put_existing"
		file := makeNewJSON(key, content)
		assertFileExists(t, key)

		bytes, _ := json.Marshal(newContent)
		err := I.Put(file, bytes)
		assertNilErr(t, err)
		assertFileExists(t, key)

		checkContentEqual(t, key, newContent)
	})
}
