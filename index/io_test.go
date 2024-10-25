package index

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	I = NewFileIndex("")
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestCrawlDirectory(t *testing.T) {

	t.Run("crawl empty directory", func(t *testing.T) {
		setup()

		checkDeepEquals(t, crawlDirectory(""), []string{})
	})

	t.Run("crawl directory with two files", func(t *testing.T) {
		setup()

		makeNewFile("one.json", "one")
		makeNewFile("two.json", "two")
		checkDeepEquals(t, crawlDirectory(""), []string{"one", "two"})
	})

	t.Run("crawl directory with non json file", func(t *testing.T) {
		setup()

		makeNewFile("one.json", "one")
		makeNewFile("test.txt", "test")
		checkDeepEquals(t, crawlDirectory(""), []string{"one"})
	})
}

func TestToMap(t *testing.T) {

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

func TestReplaceContent(t *testing.T) {

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

func TestDelete(t *testing.T) {

	t.Run("delete non-existent file", func(t *testing.T) {
		setup()

		f := &File{FileName: "doesnt-exist"}
		assertFileDoesNotExist(t, "doesnt-exist")

		err := f.Delete()
		assertErr(t, err)
		assertFileDoesNotExist(t, "doesnt-exist")
	})

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
