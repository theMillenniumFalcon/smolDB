package main

import (
	"github.com/themillenniumfalcon/smolDB/api"
	"github.com/themillenniumfalcon/smolDB/index"
	"github.com/themillenniumfalcon/smolDB/log"
)

func main() {
	log.SetLoggingLevel(log.INFO)
	log.Info("initializing smolDB")
	index.I.Regenerate("db")

	api.Serve()
}
