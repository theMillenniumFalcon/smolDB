package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/themillenniumfalcon/smolDB/api"
	"github.com/themillenniumfalcon/smolDB/index"
)

func main() {
	log.Infof("initializing smolDB")
	index.I.Regenerate("db")

	api.Serve()
}
