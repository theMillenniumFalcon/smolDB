// provides comprehensive tests for the reference resolution system
package index

import (
	"reflect"
	"strings"
	"testing"

	af "github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestResolveReferences(t *testing.T) {
	// test fixtures: sample data structures used across multiple tests

	// contains a direct reference to 'second'
	firstContentWithRef := map[string]interface{}{
		"test":      "testVal",
		"secondVal": "REF::second",
	}

	// contains a reference to 'third' - used for testing nested references
	secondContentWithRef := map[string]interface{}{
		"just": "strings",
		"ref":  "REF::third",
	}

	// base content without any references - used as final resolution target
	baseContent := map[string]interface{}{
		"key": "value",
	}

	// Test Case 1: basic string handling
	t.Run("string with no ref should be returned as is", func(t *testing.T) {
		I.SetFileSystem(af.NewMemMapFs())

		got := ResolveReferences("test", 1)
		want := "test"

		assert.Equal(t, got, want)
	})

	// Test Case 2: non-handled data types
	t.Run("datatypes other than string, slice, and map are returned as is", func(t *testing.T) {
		I.SetFileSystem(af.NewMemMapFs())

		got := ResolveReferences(2, 1)
		want := 2

		assert.Equal(t, got, want)
	})

	// Test Case 3: basic reference resolution
	t.Run("string with ref should replace the ref correctly", func(t *testing.T) {
		I.SetFileSystem(af.NewMemMapFs())

		makeNewJSON("testjson", baseContent)
		I.Regenerate()
		got := ResolveReferences("REF::testjson", 1)

		assert.Equal(t, got, baseContent)
	})

	// Test Case 4: error handling for missing references
	t.Run("string with non-existent ref should return error message", func(t *testing.T) {
		got := ResolveReferences("REF::nonexistent", 1)
		gotVal := reflect.ValueOf(got)

		if gotVal.Kind() != reflect.String {
			t.Errorf("the resolved value should have been a string but got type '%s'", gotVal.Kind())
		}

		assert.True(t, strings.Contains(gotVal.String(), "REF::ERR"))
	})

	// Test Case 5: reference resolution within slices
	t.Run("refs within a slice should all be replaced", func(t *testing.T) {
		I.SetFileSystem(af.NewMemMapFs())

		// set up two referenced JSON files
		makeNewJSON("testjson1", baseContent)
		makeNewJSON("testjson2", baseContent)

		I.Regenerate()

		// create a slice with mix of references and regular strings
		refSlice := []string{"test", "REF::testjson1", "notref", "REF::testjson2"}
		got := ResolveReferences(refSlice, 1)
		expectedSlice := []interface{}{"test", baseContent, "notref", baseContent}

		assert.Equal(t, got, expectedSlice)
	})

	// Test Case 6: reference resolution within maps
	t.Run("refs within map values should all be replaced", func(t *testing.T) {
		I.SetFileSystem(af.NewMemMapFs())

		// Set up two referenced JSON files
		makeNewJSON("testjson1", baseContent)
		makeNewJSON("testjson2", baseContent)

		I.Regenerate()

		// create a map with mix of references and regular values
		refMap := map[string]interface{}{
			"firstRef":  "REF::testjson1",
			"nonRef":    "nothing here",
			"secondRef": "REF::testjson2",
		}

		got := ResolveReferences(refMap, 1)

		expectedMap := map[string]interface{}{
			"firstRef":  baseContent,
			"nonRef":    "nothing here",
			"secondRef": baseContent,
		}

		assert.Equal(t, got, expectedMap)
	})

	// Test Case 7: nested reference resolution with sufficient depth
	t.Run("double nested refs should be resolved when depth permits", func(t *testing.T) {
		I.SetFileSystem(af.NewMemMapFs())

		// Set up chain of referenced files
		makeNewJSON("first", firstContentWithRef)
		makeNewJSON("second", secondContentWithRef)
		makeNewJSON("third", baseContent)

		I.Regenerate()

		// resolve with depth=2 to allow full resolution
		got := ResolveReferences(firstContentWithRef, 2)

		expectedMap := map[string]interface{}{
			"test": "testVal",
			"secondVal": map[string]interface{}{
				"just": "strings",
				"ref":  baseContent,
			},
		}

		assert.Equal(t, got, expectedMap)
	})

	// Test Case 8: nested reference resolution with limited depth
	t.Run("double nested refs only resolve one because of depth param", func(t *testing.T) {
		I.SetFileSystem(af.NewMemMapFs())

		// Set up chain of referenced files
		makeNewJSON("first", firstContentWithRef)
		makeNewJSON("second", secondContentWithRef)
		makeNewJSON("third", baseContent)

		I.Regenerate()

		// Resolve with depth=1 to limit resolution
		got := ResolveReferences(firstContentWithRef, 1)

		expectedMap := map[string]interface{}{
			"test": "testVal",
			"secondVal": map[string]interface{}{
				"just": "strings",
				"ref":  "REF::third", // this reference remains unresolved due to depth limit
			},
		}

		assert.Equal(t, got, expectedMap)
	})
}
