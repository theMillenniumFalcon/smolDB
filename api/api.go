package api

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/themillenniumfalcon/smolDB/index"

	"github.com/julienschmidt/httprouter"
)

func Health(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "smolDB is ok")
}

func GetKeys(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	files := index.I.ListKeys()
	fmt.Fprintf(w, "%+v", files)
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

	w.WriteHeader(http.StatusNotFound)

	fmt.Fprintf(w, "key '%s' not found", key)
	log.Warnf("key '%s' not found", key)
}

func RegenerateIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "hit regenerate index")
	index.I.Regenerate(index.I.Dir)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "404 page not found", http.StatusNotFound)
}

func Serve() {
	router := httprouter.New()

	router.GET("/", Health)
	router.GET("/getKeys", GetKeys)
	router.GET("/get/:key", GetKey)
	router.POST("/regenerate", RegenerateIndex)

	router.NotFound = http.HandlerFunc(NotFound)

	log.Info("Starting API server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
