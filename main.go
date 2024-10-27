// provides the entry point for smolDB, implementing server and shell modes
// along with lock management and graceful shutdown handling.
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

// constructs the path for the lock file based on the provided directory,
// if dir is empty or ".", the lock file is created in the current directory
func getLockLocation(dir string) string {
	base := "smoldb_lock"
	if dir == "" || dir == "." {
		return base
	}

	return dir + "/" + base
}

// attempts to create a lock file to ensure only one instance
// of smolDB is running against a specific database directory,
// returns an error if the lock already exists or cannot be created
func acquireLock(dir string) error {
	_, err := index.I.FileSystem.Stat(getLockLocation(dir))

	// create lock if it doesn't exist
	if os.IsNotExist(err) {
		_, err = index.I.FileSystem.Create(getLockLocation(dir))
		return err
	}

	return fmt.Errorf("couldn't acquire lock on %s", dir)
}

// removes the lock file, allowing other instances to access the database
func releaseLock(dir string) error {
	lockdir := getLockLocation(dir)
	return index.I.FileSystem.Remove(lockdir)
}

// performs graceful shutdown operations when the program is terminated,
// releases the lock file and logs any errors that occur during cleanup
func cleanup(dir string) {
	log.Info("\ncaught term signal! cleaning up...")

	// handles graceful shutdown and releases lock file
	err := releaseLock(dir)
	if err != nil {
		log.Warn("couldn't remove lock")
		log.Fatal(err)
		return
	}
}

// initializes the database, acquires the lock, and sets up signal handling
// for graceful shutdown, also ensures the index is up to date
func setup(dir string) {
	// initialize database setup
	log.Info("initializing smolDB")
	index.I = index.NewFileIndex(dir)
	index.I.Regenerate()

	// lock acquisition
	err := acquireLock(dir)
	if err != nil {
		log.Fatal(err)
		return
	}

	// generating index once again, ensures the index is fresh and accounts
	// for any changes that might have occurred during startup
	index.I.Regenerate()

	// creates a buffered channel c to receive OS signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// launches a goroutine that waits for a signal at channel c
	// when a signal is received, it calls cleanup function to release the lock and exits the program
	go func() {
		<-c
		cleanup(dir)
		os.Exit(1)
	}()
}

// initializes and starts the HTTP server with all API endpoints configured
func serve(port int, dir string) error {
	log.Info("initializing smolDB")
	// initialize database
	setup(dir)

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
					return shell(c.String("dir"))
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
