package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/julienschmidt/httprouter"
	"github.com/themillenniumfalcon/smolDB/api"
	"github.com/themillenniumfalcon/smolDB/index"
	"github.com/themillenniumfalcon/smolDB/log"
	"github.com/urfave/cli/v2"
)

func cleanup(lockdir string) {
	log.Info("caught term signal! cleaning up...")

	err := index.I.FileSystem.Remove(lockdir)
	if err != nil {
		log.Warn("couldn't remove lock")
		log.Fatal(err)
		return
	}
}

func getLockLocation(dir string) string {
	base := "smoldb_lock"
	if dir == "" || dir == "." {
		return base
	}

	return dir + "/" + base
}

func setup(dir string) {
	log.Info("initializing smolDB")
	index.I = index.NewFileIndex(dir)
	index.I.Regenerate()

	_, err := index.I.FileSystem.Create(getLockLocation(dir))
	if err != nil {
		log.Warn("couldn't acquire lock on %s", dir)
		log.Fatal(err)
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup(getLockLocation(dir))
		os.Exit(1)
	}()
}

func serve(port int, dir string) error {
	log.SetLoggingLevel(log.INFO)
	setup(dir)

	router := httprouter.New()

	router.GET("/", api.Health)
	router.POST("/regenerate", api.RegenerateIndex)
	router.GET("/getKeys", api.GetKeys)
	router.GET("/:key", api.GetKey)
	router.GET("/:key/:field", api.GetKeyField)
	router.PUT("/:key", api.UpdateKey)
	router.DELETE("/:key", api.DeleteKey)
	router.PATCH("/:key/:field", api.PatchKeyField)

	log.Info("starting api server on port %d", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), router)
}

func main() {
	app := &cli.App{
		Name:  "smoldb",
		Usage: "a in-memory JSON database",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "port",
				Aliases:     []string{"p"},
				Value:       8080,
				Usage:       "port to run smoldb on",
				DefaultText: "8080",
			},
			&cli.StringFlag{
				Name:        "dir",
				Aliases:     []string{"d"},
				Value:       "db",
				Usage:       "directory to look for keys",
				DefaultText: "db",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "start",
				Aliases: []string{"st"},
				Usage:   "start a smoldb server",
				Action: func(c *cli.Context) error {
					return serve(c.Int("port"), c.String("dir"))
				},
			}, {
				Name:    "shell",
				Aliases: []string{"sh"},
				Usage:   "start an interactive smoldb shell",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
