package index

import (
	"fmt"
	"reflect"
	"strings"
)

func ResolveReferences(jsonVal interface{}, depthLeft int) interface{} {
	if depthLeft < 1 {
		return jsonVal
	}

	val := reflect.ValueOf(jsonVal)

	switch val.Kind() {
	case reflect.String:
		valString := val.String()

		if strings.Contains(valString, "REF::") {
			resolvedString := resolveString(valString, depthLeft)
			return resolvedString
		}
		return valString

	case reflect.Slice:
		numberOfValues := val.Len()
		newSlice := make([]interface{}, numberOfValues)

		for i := 0; i < numberOfValues; i++ {
			pointer := val.Index(i)
			newSlice[i] = ResolveReferences(pointer.Interface(), depthLeft)
		}
		return newSlice

	case reflect.Map:
		newMap := make(map[string]interface{})

		for _, key := range val.MapKeys() {
			nestedVal := val.MapIndex(key).Interface()
			newMap[key.String()] = ResolveReferences(nestedVal, depthLeft)
		}
		return newMap

	default:
		return jsonVal
	}
}

func resolveString(valString string, depthLeft int) interface{} {
	key := strings.Replace(valString, "REF::", "", 1)

	file, ok := I.Lookup(key)
	if ok {
		jsonMap, err := file.ToMap()
		if err != nil {
			errMessage := fmt.Sprintf("REF::ERR key '%s' cannot be parsed into json: %s", key, err.Error())
			return errMessage
		}

		return ResolveReferences(jsonMap, depthLeft-1)
	}

	return fmt.Sprintf("REF::ERR key '%s' not found", key)
}
