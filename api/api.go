package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/themillenniumfalcon/smolDB/index"
	"github.com/themillenniumfalcon/smolDB/log"

	"github.com/julienschmidt/httprouter"
)

func resolveReferences(w http.ResponseWriter, jsonVal interface{}) interface{} {
	val := reflect.ValueOf(jsonVal)

	switch val.Kind() {
	case reflect.String:
		valString := val.String()

		if strings.Contains(valString, "REF::") {
			key := strings.Replace(valString, "REF::", "", 1)
			file, ok := index.I.Lookup(key)

			if ok {
				jsonMap, err := file.ToMap()
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					log.WWarn(w, "err key '%s' cannot be parsed into json: %s", key, err.Error())
				}

				return resolveReferences(w, jsonMap)
			}

			w.WriteHeader(http.StatusNotFound)
			log.WWarn(w, "key '%s' not found", key)
		}

		return valString
	case reflect.Slice:
		numberOfValues := val.Len()
		newSlice := make([]interface{}, numberOfValues)

		for i := 0; i < numberOfValues; i++ {
			pointer := val.Index(i)
			newSlice[i] = resolveReferences(w, pointer.Interface())
		}

		return newSlice
	case reflect.Map:
		newMap := make(map[string]interface{})

		for _, key := range val.MapKeys() {
			nestedVal := val.MapIndex(key).Interface()
			newMap[key.String()] = resolveReferences(w, nestedVal)
		}

		return newMap
	default:
		return val
	}
}

func Health(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	response := map[string]string{"message": "smolDB is working fine!"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func RegenerateIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	index.I.Regenerate()
	log.WInfo(w, "regenerated index")
}

func GetKeys(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Info("retrieving index")
	files := index.I.ListKeys()

	data := struct {
		Files []string `json:"files"`
	}{
		Files: files,
	}

	w.Header().Set("Content-Type", "application/json")

	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "%+v", string(jsonData))
}

func GetKey(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	log.Info("get key '%s'", key)

	file, ok := index.I.Lookup(key)
	if ok {
		w.Header().Set("Content-Type", "application/json")

		jsonMap, err := file.ToMap()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.WWarn(w, "err key '%s' cannot be parsed into json: %s", key, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		resolvedJsonMap := resolveReferences(w, jsonMap)
		jsonData, _ := json.Marshal(resolvedJsonMap)
		fmt.Fprintf(w, "%+v", string(jsonData))
		return
	}

	w.WriteHeader(http.StatusNotFound)
	log.WWarn(w, "key '%s' not found", key)
}

func GetKeyField(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	field := ps.ByName("field")

	log.Info("get field '%s' in key '%s'", field, key)

	file, ok := index.I.Lookup(key)
	if ok {
		jsonMap, err := file.ToMap()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.WWarn(w, "err key '%s' cannot be parsed into json: %s", key, err.Error())
			return
		}

		val, ok := jsonMap[field]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			log.WWarn(w, "err key '%s' does not have field '%s'", key, field)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		resolvedValue := resolveReferences(w, val)
		jsonData, _ := json.Marshal(resolvedValue)
		fmt.Fprintf(w, "%+v", string(jsonData))
		return
	}

	w.WriteHeader(http.StatusNotFound)
	log.WWarn(w, "key '%s' not found", key)
}

func UpdateKey(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	log.Info("put key '%s'", key)
	file, ok := index.I.Lookup(key)

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.WWarn(w, "err reading body when key '%s': %s", key, err.Error())
		return
	}

	err = index.I.Put(file, bodyBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.WWarn(w, "err updating key '%s': %s", key, err.Error())
		return
	}

	if ok {
		log.WInfo(w, "update '%s' successful", key)
		return
	}
	log.WInfo(w, "create '%s' successful", key)
}

func DeleteKey(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	log.Info("delete key '%s'", key)

	file, ok := index.I.Lookup(key)
	if ok {
		err := index.I.Delete(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.WWarn(w, "err unable to delete key '%s': '%s'", key, err.Error())
			return
		}
		log.WInfo(w, "delete '%s' successful", key)
		return
	}

	w.WriteHeader(http.StatusNotFound)
	log.WWarn(w, "key '%s' does not exist", key)
}

func PatchKeyField(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	field := ps.ByName("field")
	log.Info("patch field '%s' in key '%s'", field, key)

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.WWarn(w, "err reading body with key '%s': %s", key, err.Error())
		return
	}

	file, ok := index.I.Lookup(key)
	if ok {
		jsonMap, err := file.ToMap()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.WWarn(w, "err key '%s' cannot be parsed into json: %s", key, err.Error())
			return
		}

		var parsedJSON map[string]interface{}

		err = json.Unmarshal(bodyBytes, &parsedJSON)
		if err != nil {
			jsonMap[field] = string(bodyBytes)
		} else {
			jsonMap[field] = parsedJSON
		}
		jsonData, _ := json.Marshal(jsonMap)

		err = file.ReplaceContent(string(jsonData))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.WWarn(w, "err setting content of key '%s': %s", key, err.Error())
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		log.WInfo(w, "patch field '%s' of key '%s' successful", field, key)
		return
	}

	w.WriteHeader(http.StatusNotFound)
	log.WWarn(w, "key '%s' not found", key)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "404 page not found", http.StatusNotFound)
}
