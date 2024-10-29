// provides the entry point for smolDB, implementing server and shell modes
// along with lock management and graceful shutdown handling.
package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/themillenniumfalcon/smolDB/api"
	"github.com/themillenniumfalcon/smolDB/log"
	"github.com/themillenniumfalcon/smolDB/sh"
	"github.com/urfave/cli/v2"
)

// initializes and starts the HTTP server with all API endpoints configured
func serve(port int, dir string) error {
	log.Info("initializing smolDB")
	// initialize database
	sh.Setup(dir)

	// set up HTTP router
	router := httprouter.New()

	// register API endpoints
	// base routes
	router.GET("/", api.Health)
	router.GET("/keys", api.GetKeys)
	router.POST("/regenerate", api.RegenerateIndex)

	// key-based routes
	router.GET("/key/:key", api.GetKey)
	router.PUT("/key/:key", api.UpdateKey)
	router.DELETE("/key/:key", api.DeleteKey)

	// field-based routes
	router.GET("/key/:key/field/:field", api.GetKeyField)
	router.PATCH("/key/:key/field/:field", api.PatchKeyField)

	log.Info("starting api server on port %d", port)
	// start HTTP server
	return http.ListenAndServe(fmt.Sprintf(":%d", port), router)
}

// sets up the CLI interface and handles both server and shell modes of operation
func main() {
	app := &cli.App{
		Name:  "smoldb",
		Usage: "an in-memory JSON database",
		// global flags for port and directory
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
		// command definitions for 'start' and 'shell'
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
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:        "port",
						Aliases:     []string{"p"},
						Value:       8080,
						Usage:       "port to run smoldb on",
						DefaultText: "8080",
					},
				},
				Action: func(c *cli.Context) error {
					return sh.Shell(c.String("dir"))
				},
			},
		},
	}

	// run the application
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
