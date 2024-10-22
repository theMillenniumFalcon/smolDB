package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/themillenniumfalcon/smolDB/index"

	"github.com/julienschmidt/httprouter"
)

func Health(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "smolDB is ok")
}

func RegenerateIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "hit regenerate index")
	index.I.Regenerate(index.I.Dir)
}

func GetKeys(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
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
	log.Infof("get key '%s'", key)

	file, ok := index.I.Lookup(key)
	if ok {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, file.ResolvePath())
		return
	}

	write404(w, key)
}

func UpdateKey(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	file, _ := index.I.Lookup(key)

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	file.ReplaceContent(string(bodyBytes))
	fmt.Fprintf(w, "update '%s' successful", key)
}

func write404(w http.ResponseWriter, key string) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "key '%s' not found", key)
	log.Warnf("key '%s' not found", key)
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

	router.NotFound = http.HandlerFunc(NotFound)

	log.Info("Starting API server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
