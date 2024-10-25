package index

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFile_ResolvePath(t *testing.T) {
	t.Run("test path correct with empty dir", func(t *testing.T) {
		setup()

		file := createAndReturnFile(t, "resolve_empty")
		got := file.ResolvePath()
		want := "resolve_empty.json"

		checkDeepEquals(t, got, want)
	})

	t.Run("file path correct with directories", func(t *testing.T) {
		setup()

		I.dir = "db"
		file := createAndReturnFile(t, "resolve_test")

		got := file.ResolvePath()
		want := "db/resolve_test.json"

		checkDeepEquals(t, got, want)
	})
}

func TestFileIndex_Lookup(t *testing.T) {
	t.Run("lookup existing file", func(t *testing.T) {
		setup()

		key := "lookup"
		createAndReturnFile(t, key)

		file, ok := I.Lookup(key)
		if !ok {
			t.Errorf("should have found file: '%s'", file.FileName)
		}

		bytes, _ := file.getByteArray()
		checkDeepEquals(t, string(bytes), "test")
	})

	t.Run("lookup non-existent file", func(t *testing.T) {
		setup()

		file, ok := I.Lookup("doesnt_exist")
		if ok {
			t.Errorf("should not have found file: '%s'", file.FileName)
		}
	})
}

func TestFileIndex_Delete(t *testing.T) {
	t.Run("delete file that exists", func(t *testing.T) {
		setup()

		key := "delete_test1"
		file := createAndReturnFile(t, key)
		err := I.Delete(file)
		assertNilErr(t, err)

		checkKeyNotInIndex(t, key)
	})

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

func TestFileIndex_List(t *testing.T) {
	t.Run("list empty dir", func(t *testing.T) {
		setup()

		list := I.ListKeys()
		checkDeepEquals(t, len(list), 0)
	})

	t.Run("list dir with two files", func(t *testing.T) {
		setup()

		createAndReturnFile(t, "list1")
		createAndReturnFile(t, "list2")

		assert.True(t, sliceContains(I.ListKeys(), "list1"))
		assert.True(t, sliceContains(I.ListKeys(), "list2"))
	})
}

func TestFileIndex_Regenerate(t *testing.T) {
	t.Run("test if new files are added to current index", func(t *testing.T) {
		setup()

		makeNewFile("regenerate1.json", "test")
		makeNewFile("regenerate2.json", "test")

		checkDeepEquals(t, len(I.ListKeys()), 0)

		I.Regenerate()

		assert.True(t, sliceContains(I.ListKeys(), "regenerate1"))
		assert.True(t, sliceContains(I.ListKeys(), "regenerate2"))
	})

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

func TestFileIndex_Put(t *testing.T) {
	content := map[string]interface{}{
		"array": []interface{}{
			"a",
			map[string]interface{}{
				"test": "deep nest",
			},
		},
	}

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