package api

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/themillenniumfalcon/smolDB/index"
	"github.com/themillenniumfalcon/smolDB/log"
)

// CheckKeyIntegrity verifies the integrity of a specific key
func CheckKeyIntegrity(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	log.Info("checking integrity for key: %s", key)

	file, ok := index.I.Lookup(key)
	if !ok {
		http.Error(w, fmt.Sprintf("key '%s' not found", key), http.StatusNotFound)
		return
	}

	err := file.ValidateChecksum()
	if err != nil {
		http.Error(w, fmt.Sprintf("integrity check failed: %v", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "integrity check passed for key '%s'", key)
}

// RepairKeyIntegrity updates the checksum for a specific key
func RepairKeyIntegrity(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	log.Info("repairing integrity for key: %s", key)

	file, ok := index.I.Lookup(key)
	if !ok {
		http.Error(w, fmt.Sprintf("key '%s' not found", key), http.StatusNotFound)
		return
	}

	err := file.RepairChecksum()
	if err != nil {
		http.Error(w, fmt.Sprintf("integrity repair failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "integrity repaired for key '%s'", key)
}
