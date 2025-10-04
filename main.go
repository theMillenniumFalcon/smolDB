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
func serve(port int, dir string, durability string, groupMs int, groupBatch int, syncMode string) error {
	log.Info("initializing smolDB")
	// initialize database
	sh.SetupWithOptions(dir, durability, groupMs, groupBatch, syncMode)

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

	// integrity routes
	router.GET("/integrity/:key", api.CheckKeyIntegrity)
	router.POST("/integrity/:key/repair", api.RepairKeyIntegrity)

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
			&cli.StringFlag{
				Name:        "durability",
				Usage:       "durability level: none|commit|grouped",
				Value:       "commit",
				DefaultText: "commit",
				EnvVars:     []string{"SMOLDB_DURABILITY"},
			},
			&cli.IntFlag{
				Name:        "group-commit-ms",
				Usage:       "group commit fsync interval in ms (used when durability=grouped)",
				Value:       0,
				DefaultText: "0",
				EnvVars:     []string{"SMOLDB_GROUP_COMMIT_MS"},
			},
			&cli.IntFlag{
				Name:        "group-commit-batch",
				Usage:       "group commit fsync after this many appends (used when durability=grouped)",
				Value:       0,
				DefaultText: "0",
				EnvVars:     []string{"SMOLDB_GROUP_COMMIT_BATCH"},
			},
			&cli.StringFlag{
				Name:        "sync-mode",
				Usage:       "sync mode: none|fsync|dsync (dsync best-effort)",
				Value:       "fsync",
				DefaultText: "fsync",
				EnvVars:     []string{"SMOLDB_SYNC_MODE"},
			},
		},
		// command definitions for 'start' and 'shell'
		Commands: []*cli.Command{
			{
				Name:    "start",
				Aliases: []string{"st"},
				Usage:   "start a smoldb server",
				Action: func(c *cli.Context) error {
					return serve(
						c.Int("port"),
						c.String("dir"),
						c.String("durability"),
						c.Int("group-commit-ms"),
						c.Int("group-commit-batch"),
						c.String("sync-mode"),
					)
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
					&cli.StringFlag{
						Name:        "durability",
						Usage:       "durability level: none|commit|grouped",
						Value:       "commit",
						DefaultText: "commit",
						EnvVars:     []string{"SMOLDB_DURABILITY"},
					},
					&cli.IntFlag{
						Name:        "group-commit-ms",
						Usage:       "group commit fsync interval in ms (used when durability=grouped)",
						Value:       0,
						DefaultText: "0",
						EnvVars:     []string{"SMOLDB_GROUP_COMMIT_MS"},
					},
					&cli.IntFlag{
						Name:        "group-commit-batch",
						Usage:       "group commit fsync after this many appends (used when durability=grouped)",
						Value:       0,
						DefaultText: "0",
						EnvVars:     []string{"SMOLDB_GROUP_COMMIT_BATCH"},
					},
					&cli.StringFlag{
						Name:        "sync-mode",
						Usage:       "sync mode: none|fsync|dsync (dsync best-effort)",
						Value:       "fsync",
						DefaultText: "fsync",
						EnvVars:     []string{"SMOLDB_SYNC_MODE"},
					},
				},
				Action: func(c *cli.Context) error {
					return sh.ShellWithOptions(
						c.String("dir"),
						c.String("durability"),
						c.Int("group-commit-ms"),
						c.Int("group-commit-batch"),
						c.String("sync-mode"),
					)
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
