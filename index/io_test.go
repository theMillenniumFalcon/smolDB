package index

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	af "github.com/spf13/afero"
)

func checkDeepEquals(t *testing.T, a interface{}, b interface{}) {
	t.Helper()
	if !cmp.Equal(a, b) {
		t.Errorf("got %+v, want %+v", a, b)
	}
}

func checkJSONEquals(t *testing.T, a interface{}, b interface{}) {
	t.Helper()
	if fmt.Sprintf("%+v", a) != fmt.Sprintf("%+v", b) {
		t.Errorf("got %+v, want %+v", a, b)
	}
}

func makeNewFile(fs af.Fs, name string, contents string) {
	af.WriteFile(fs, name, []byte(contents), 0644)
}

func makeNewJSON(fs af.Fs, name string, contents map[string]interface{}) *File {
	jsonData, _ := json.Marshal(contents)
	af.WriteFile(fs, name+".json", jsonData, 0644)
	return &File{FileName: name}
}

func mapToString(contents map[string]interface{}) string {
	jsonData, _ := json.Marshal(contents)
	return string(jsonData)
}

func assertNilErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("got error %s when shouldn't have", err.Error())
	}
}

func assertErr(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("didnt get error when wanted error")
	}
}

func TestCrawlDirectory(t *testing.T) {

	t.Run("crawl empty directory", func(t *testing.T) {
		fs = af.NewMemMapFs()
		checkDeepEquals(t, crawlDirectory(""), []string{})
	})

	t.Run("crawl directory with two files", func(t *testing.T) {
		fs = af.NewMemMapFs()
		makeNewFile(fs, "one.json", "one")
		makeNewFile(fs, "two.json", "two")
		checkDeepEquals(t, crawlDirectory(""), []string{"one", "two"})
	})

	t.Run("crawl directory with non json file", func(t *testing.T) {
		fs = af.NewMemMapFs()
		makeNewFile(fs, "one.json", "one")
		makeNewFile(fs, "test.txt", "test")
		checkDeepEquals(t, crawlDirectory(""), []string{"one"})
	})
}

func TestToMap(t *testing.T) {
	t.Run("simple flat json to map", func(t *testing.T) {
		fs = af.NewMemMapFs()

		expected := map[string]interface{}{
			"field1": "value1",
			"field2": "value2",
		}

		f := makeNewJSON(fs, "test", expected)

		got, err := f.ToMap()
		assertNilErr(t, err)
		checkDeepEquals(t, expected, got)
	})

	t.Run("json with array", func(t *testing.T) {
		fs = af.NewMemMapFs()

		expected := map[string]interface{}{
			"array": []string{
				"element1",
				"element2",
			},
		}

		f := makeNewJSON(fs, "test", expected)
		got, err := f.ToMap()
		assertNilErr(t, err)
		checkJSONEquals(t, expected, got)
	})

	t.Run("deep nested with map", func(t *testing.T) {
		fs = af.NewMemMapFs()

		expected := map[string]interface{}{
			"array": []interface{}{
				"a",
				map[string]interface{}{
					"test": "deep nest",
				},
			},
		}

		f := makeNewJSON(fs, "test", expected)
		got, err := f.ToMap()
		assertNilErr(t, err)
		checkJSONEquals(t, expected, got)
	})
}

func TestReplaceContent(t *testing.T) {
	t.Run("change all content", func(t *testing.T) {
		fs = af.NewMemMapFs()

		old := map[string]interface{}{
			"field1": "value1",
			"field2": "value2",
		}

		new := map[string]interface{}{
			"field": "value",
		}

		f := makeNewJSON(fs, "test", old)
		err := f.ReplaceContent(mapToString(new))

		assertNilErr(t, err)
		got, err := f.ToMap()
		assertNilErr(t, err)
		checkDeepEquals(t, got, new)
	})
}

func TestDelete(t *testing.T) {
	t.Run("delete non-existent file", func(t *testing.T) {
		fs = af.NewMemMapFs()
		f := &File{FileName: "doesnt-exist"}
		err := f.Delete()
		assertErr(t, err)
	})
}
