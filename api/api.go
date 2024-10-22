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
	log.Infof("attempt to get key %s", key)

	file, ok := index.I.Lookup(key)
	fmt.Fprintf(w, "%+v, %t", file.FileName, ok)
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
