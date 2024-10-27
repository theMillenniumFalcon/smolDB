// provides functionality for resolving references within JSON-like data structures
package index

import (
	"fmt"
	"reflect"
	"strings"
)

// traverses through a JSON-like data structure and resolves any references
// marked with "REF::" prefix, it handles nested structures including maps and slices
// Parameters:
//   - jsonVal: input value to process (can be any JSON-compatible type)
//   - depthLeft: maximum depth of recursive reference resolution to prevent infinite loops
func ResolveReferences(jsonVal interface{}, depthLeft int) interface{} {
	// if no more depth allowed, return value as-is
	if depthLeft < 1 {
		return jsonVal
	}

	val := reflect.ValueOf(jsonVal)

	switch val.Kind() {
	case reflect.String:
		valString := val.String()

		// check if string contains a reference marker
		if strings.Contains(valString, "REF::") {
			resolvedString := resolveString(valString, depthLeft)
			return resolvedString
		}
		return valString

	case reflect.Slice:
		// create a new slice to hold resolved values
		numberOfValues := val.Len()
		newSlice := make([]interface{}, numberOfValues)

		// recursively resolve each element in the slice
		for i := 0; i < numberOfValues; i++ {
			pointer := val.Index(i)
			newSlice[i] = ResolveReferences(pointer.Interface(), depthLeft)
		}
		return newSlice

	case reflect.Map:
		// create a new map to hold resolved values
		newMap := make(map[string]interface{})

		// recursively resolve each value in the map
		for _, key := range val.MapKeys() {
			nestedVal := val.MapIndex(key).Interface()
			newMap[key.String()] = ResolveReferences(nestedVal, depthLeft)
		}
		return newMap

	default:
		return jsonVal
	}
}

// handles the resolution of a single reference string,
// expects strings in the format "REF::key" where key is the lookup key
// Parameters:
//   - valString: the reference string to resolve (must start with "REF::")
//   - depthLeft: remaining depth for nested reference resolution
func resolveString(valString string, depthLeft int) interface{} {
	// extract the key by removing the "REF::" prefix
	key := strings.Replace(valString, "REF::", "", 1)

	// look up the key in the index
	file, ok := I.Lookup(key)
	if ok {
		jsonMap, err := file.ToMap()
		if err != nil {
			errMessage := fmt.Sprintf("REF::ERR key '%s' cannot be parsed into json: %s", key, err.Error())
			return errMessage
		}

		// recursively resolve any references in the found map
		return ResolveReferences(jsonMap, depthLeft-1)
	}

	return fmt.Sprintf("REF::ERR key '%s' not found", key)
}
