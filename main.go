// provides the entry point for smolDB, implementing server and shell modes
// along with lock management and graceful shutdown handling.
package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/themillenniumfalcon/smolDB/admin"
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
				Name: "admin",
				Subcommands: []*cli.Command{
					{
						Name:  "compact",
						Usage: "compact the database by rewriting JSON files and trimming WAL",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "force",
								Usage: "force compaction even if database is locked",
								Value: false,
							},
						},
						Action: func(c *cli.Context) error {
							stats, err := admin.CompactDB(c.String("dir"), c.Bool("force"))
							if err != nil {
								return err
							}
							log.Info("Compaction complete:")
							log.Info("- Files processed: %d", stats.FilesProcessed)
							log.Info("- Size before: %d bytes", stats.BytesBefore)
							log.Info("- Size after: %d bytes", stats.BytesAfter)
							log.Info("- Space saved: %d bytes", stats.BytesBefore-stats.BytesAfter)
							if stats.WalEntriesTrimmed > 0 {
								log.Info("- WAL entries trimmed: %d", stats.WalEntriesTrimmed)
							}
							return nil
						},
					},
					{
						Name:  "verify",
						Usage: "verify database integrity",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "repair",
								Usage: "attempt to repair issues found",
								Value: false,
							},
						},
						Action: func(c *cli.Context) error {
							report, err := admin.VerifyDB(c.String("dir"), c.Bool("repair"))
							if err != nil {
								return err
							}
							log.Info("Integrity check complete:")
							log.Info("- Total files: %d", report.TotalFiles)
							log.Info("- Valid files: %d", report.ValidFiles)
							if len(report.InvalidFiles) > 0 {
								log.Warn("- Invalid files (%d):", len(report.InvalidFiles))
								for _, f := range report.InvalidFiles {
									log.Warn("  - %s", f)
								}
							}
							if len(report.Repairs) > 0 {
								log.Info("- Repaired files (%d):", len(report.Repairs))
								for _, f := range report.Repairs {
									log.Info("  - %s", f)
								}
							}
							if len(report.IndexMismatches) > 0 {
								log.Warn("- Index mismatches (%d):", len(report.IndexMismatches))
								for _, f := range report.IndexMismatches {
									log.Warn("  - %s", f)
								}
							}
							return nil
						},
					},
				},
			},
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
