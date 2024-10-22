package main

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/julienschmidt/httprouter"
)

func GetIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func GetKey(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	fmt.Fprintf(w, "attempt to get key %s", key)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "404 page not found", http.StatusNotFound)
}

func main() {
	router := httprouter.New()
	router.GET("/", GetIndex)
	router.GET("/:key", GetKey)

	router.NotFound = http.HandlerFunc(NotFound)

	log.Info("Starting server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
