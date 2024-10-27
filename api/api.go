// provides the core API functions that are used for smolDB server and shell
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/themillenniumfalcon/smolDB/index"
	"github.com/themillenniumfalcon/smolDB/log"

	"github.com/julienschmidt/httprouter"
)

// HTTP status codes
const (
	successStatus     = http.StatusOK
	notFoundStatus    = http.StatusNotFound
	badRequestStatus  = http.StatusBadRequest
	serverErrorStatus = http.StatusInternalServerError
)

// extracts the 'depth' query parameter from the request URL
// returns default value of 3 if not specified or if parsing fails
func getMaxDepthParam(r *http.Request) int {
	maxDepth := 3
	maxDepthStr := r.URL.Query().Get("depth")

	parsedInt, err := strconv.Atoi(maxDepthStr)
	if err == nil {
		maxDepth = parsedInt
	}

	return maxDepth
}

// handles GET /health
// returns a simple status message indicating the service is running
func Health(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	response := map[string]string{"message": "smolDB is working fine!"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handles GET /keys
// returns a list of all keys in the database
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

// handles POST /regenerate
// rebuilds the entire database index
func RegenerateIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	index.I.Regenerate()
	log.WInfo(w, "regenerated index")
}

// handles GET /key/:key
// returns the full JSON content for a specific key
// supports recursive resolution of references up to specified depth
func GetKey(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	log.Info("get key '%s'", key)

	file, ok := index.I.Lookup(key)
	if ok {
		w.Header().Set("Content-Type", "application/json")

		jsonMap, err := file.ToMap()
		if err != nil {
			w.WriteHeader(badRequestStatus)
			log.WWarn(w, "err key '%s' cannot be parsed into json: %s", key, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		maxDepth := getMaxDepthParam(r)
		resolvedJsonMap := index.ResolveReferences(jsonMap, maxDepth)

		jsonData, _ := json.Marshal(resolvedJsonMap)
		fmt.Fprintf(w, "%+v", string(jsonData))
		return
	}

	w.WriteHeader(notFoundStatus)
	log.WWarn(w, "key '%s' not found", key)
}

// handles PUT /key/:key
// creates or updates the content for a specific key
func UpdateKey(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	log.Info("put key '%s'", key)
	file, ok := index.I.Lookup(key)

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(badRequestStatus)
		log.WWarn(w, "err reading body when key '%s': %s", key, err.Error())
		return
	}

	err = index.I.Put(file, bodyBytes)
	if err != nil {
		w.WriteHeader(serverErrorStatus)
		log.WWarn(w, "err updating key '%s': %s", key, err.Error())
		return
	}

	if ok {
		log.WInfo(w, "update '%s' successful", key)
		return
	}
	log.WInfo(w, "create '%s' successful", key)
}

// handles DELETE /key/:key
// removes a key and its associated content from the database
func DeleteKey(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	log.Info("delete key '%s'", key)

	file, ok := index.I.Lookup(key)
	if ok {
		err := index.I.Delete(file)
		if err != nil {
			w.WriteHeader(serverErrorStatus)
			log.WWarn(w, "err unable to delete key '%s': '%s'", key, err.Error())
			return
		}
		log.WInfo(w, "delete '%s' successful", key)
		return
	}

	w.WriteHeader(notFoundStatus)
	log.WWarn(w, "key '%s' does not exist", key)
}

// handles GET /key/:key/field/:field
// returns the value of a specific field within a key's JSON content
// supports recursive resolution of references up to specified depth
func GetKeyField(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	field := ps.ByName("field")

	log.Info("get field '%s' in key '%s'", field, key)

	file, ok := index.I.Lookup(key)
	if ok {
		jsonMap, err := file.ToMap()
		if err != nil {
			w.WriteHeader(badRequestStatus)
			log.WWarn(w, "err key '%s' cannot be parsed into json: %s", key, err.Error())
			return
		}

		// look up the specified field
		val, ok := jsonMap[field]
		if !ok {
			w.WriteHeader(badRequestStatus)
			log.WWarn(w, "err key '%s' does not have field '%s'", key, field)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		maxDepth := getMaxDepthParam(r)
		resolvedValue := index.ResolveReferences(val, maxDepth)

		jsonData, _ := json.Marshal(resolvedValue)
		fmt.Fprintf(w, "%+v", string(jsonData))
		return
	}

	w.WriteHeader(notFoundStatus)
	log.WWarn(w, "key '%s' not found", key)
}

// handles PATCH /key/:key/field/:field
// updates a specific field within a key's JSON content
// handles both primitive values and nested JSON objects
func PatchKeyField(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	field := ps.ByName("field")
	log.Info("patch field '%s' in key '%s'", field, key)

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(badRequestStatus)
		log.WWarn(w, "err reading body with key '%s': %s", key, err.Error())
		return
	}

	file, ok := index.I.Lookup(key)
	if ok {
		jsonMap, err := file.ToMap()
		if err != nil {
			w.WriteHeader(badRequestStatus)
			log.WWarn(w, "err key '%s' cannot be parsed into json: %s", key, err.Error())
			return
		}

		var parsedJSON map[string]interface{}

		err = json.Unmarshal(bodyBytes, &parsedJSON)
		if err != nil {
			// if parsing as JSON fails, treat it as a primitive value
			jsonMap[field] = string(bodyBytes)
		} else {
			// if parsing succeeds, it's a JSON object
			jsonMap[field] = parsedJSON
		}
		jsonData, _ := json.Marshal(jsonMap)

		err = file.ReplaceContent(string(jsonData))
		if err != nil {
			w.WriteHeader(serverErrorStatus)
			log.WWarn(w, "err setting content of key '%s': %s", key, err.Error())
			return
		}

		w.WriteHeader(successStatus)
		log.WInfo(w, "patch field '%s' of key '%s' successful", field, key)
		return
	}

	w.WriteHeader(notFoundStatus)
	log.WWarn(w, "key '%s' not found", key)
}

// handles all unmatched routes
// returns a standard 404 error
func NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "404 page not found", notFoundStatus)
}
