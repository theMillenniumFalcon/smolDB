package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/themillenniumfalcon/smolDB/index"
	"github.com/themillenniumfalcon/smolDB/log"

	"github.com/julienschmidt/httprouter"
)

func Health(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.WInfo(w, "smolDB is working fine!")
}

func RegenerateIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	index.I.Regenerate(index.I.Dir)
	log.WInfo(w, "regenerated index")
}

func GetKeys(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	log.Info("retrieving index")
	files := index.I.ListKeys()

	data := struct {
		Files []string `json:"files"`
	}{
		Files: files,
	}

	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "%+v", string(jsonData))
}

func GetKey(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	log.Info("get key '%s'", key)

	file, ok := index.I.Lookup(key)
	if ok {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, file.ResolvePath())
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
		log.Warn("err reading body when key '%s': '%s'", key, err.Error())
	}

	err = index.I.Put(file, bodyBytes)
	if err != nil {
		log.Warn("err updating key '%s': '%s'", key, err.Error())
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
			log.Warn("unable to delete key '%s': '%s'", key, err.Error())
		}
		log.WInfo(w, "key '%s' deleted successfully", key)
		return
	}

	log.WInfo(w, "key '%s' does not exist", key)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "404 page not found", http.StatusNotFound)
}

func Serve() {
	router := httprouter.New()

	router.GET("/", Health)
	router.POST("/regenerate", RegenerateIndex)
	router.GET("/getKeys", GetKeys)
	router.GET("/get/:key", GetKey)
	router.PUT("/put/:key", UpdateKey)
	router.DELETE("/delete/:key", DeleteKey)

	router.NotFound = http.HandlerFunc(NotFound)

	log.Info("Starting API server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
