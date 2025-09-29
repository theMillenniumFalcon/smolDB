// provides the shell interface for smolDB, allowing interactive
// command-line operations on the database
package sh

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/julienschmidt/httprouter"
	"github.com/themillenniumfalcon/smolDB/api"
	"github.com/themillenniumfalcon/smolDB/index"
	"github.com/themillenniumfalcon/smolDB/log"
)

// DefaultDepth defines the default depth for resolving nested references
// when no depth parameter is provided
const DefaultDepth = 0

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

// execInput processes the user input and executes the corresponding command,
// it supports commands: index, listAll, lookup, delete, regenerate, and exit
func execInput(input string, dir string) (err error) {
	input = strings.TrimSuffix(input, "\n")
	args := strings.Split(input, " ")

	switch args[0] {
	case "index":
		healthWrapper()
	case "listAll":
		listAllWrapper()
	case "lookup":
		return lookupWrapper(args)
	case "delete":
		return deleteWrapper(args)
	case "regenerate":
		index.I.Regenerate()
	case "exit":
		cleanup(dir)
		os.Exit(0)
	default:
		log.Warn("'%s' is not a valid command.", args[0])
		log.Info("valid commands: index, listAll, lookup <key> <depth>, delete <key>, regenerate, exit")
	}

	return err
}

// initializes the database, acquires the lock, and sets up signal handling
// for graceful shutdown, also ensures the index is up to date
func Setup(dir string) {
	// initialize database setup
	log.Info("initializing smolDB")
	index.I = index.NewFileIndex(dir)
	// initialize WAL with commit durability and replay
	if err := index.I.InitWAL(index.DurabilityCommit); err != nil {
		log.Warn("failed to init WAL: %s", err.Error())
	} else {
		if index.I != nil && index.I.WALAvailable() {
			_ = index.I.WALReplay()
		}
	}
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

// SetupWithOptions is like Setup, but allows configuring durability and group commit interval
func SetupWithOptions(dir string, durability string, groupCommitMs int, groupCommitBatch int, syncMode string) {
	log.Info("initializing smolDB")
	index.I = index.NewFileIndex(dir)

	// pick durability level
	level := index.DurabilityCommit
	switch durability {
	case "none":
		level = index.DurabilityNone
	case "commit":
		level = index.DurabilityCommit
	case "grouped":
		level = index.DurabilityGrouped
	}

	// set sync mode
	switch syncMode {
	case "none":
		index.I.SetSyncMode(index.SyncNone)
	case "fsync":
		index.I.SetSyncMode(index.SyncFsync)
	case "dsync":
		index.I.SetSyncMode(index.SyncDSync)
	default:
		index.I.SetSyncMode(index.SyncFsync)
	}

	// initialize WAL with chosen durability and grouped interval/batch
	if err := index.I.InitWALWithOptions(level, groupCommitMs, groupCommitBatch); err != nil {
		log.Warn("failed to init WAL: %s", err.Error())
	} else {
		if index.I != nil && index.I.WALAvailable() {
			_ = index.I.WALReplay()
		}
	}
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

// ShellWithOptions runs the shell with durability configuration
func ShellWithOptions(dir string, durability string, groupCommitMs int, groupCommitBatch int, syncMode string) error {
	log.IsShellMode = true
	log.Info("starting smoldb shell...")

	SetupWithOptions(dir, durability, groupCommitMs, groupCommitBatch, syncMode)
	reader := bufio.NewReader(os.Stdin)

	// the main shell loop, displays a prompt, read user input, and executes input
	for {
		log.Prompt("smoldb> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			log.Warn("err reading input: %s", err.Error())
		}

		if err = execInput(input, dir); err != nil {
			log.Warn("err executing input: %s", err.Error())
		}
	}
}

// shell initializes and runs the interactive shell interface for smolDB,
// it takes a directory path as input where the database files are stored
func Shell(dir string) error {
	log.IsShellMode = true
	log.Info("starting smoldb shell...")

	Setup(dir)
	reader := bufio.NewReader(os.Stdin)

	// the main shell loop, displays a prompt, read user input, and executes input
	for {
		log.Prompt("smoldb> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			log.Warn("err reading input: %s", err.Error())
		}

		if err = execInput(input, dir); err != nil {
			log.Warn("err executing input: %s", err.Error())
		}
	}
}

// parseDepthFromArgs extracts and parses the depth parameter from command arguments,
// returns DefaultDepth if no valid depth parameter is provided
func parseDepthFromArgs(args []string) int {
	if len(args) < 3 {
		return DefaultDepth
	}
	if parsedInt, err := strconv.Atoi(args[2]); err == nil {
		return parsedInt
	}
	return DefaultDepth
}

// healthWrapper checks the database health by making a mock HTTP request
// to the health endpoint
func healthWrapper() error {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/health", nil)

	api.Health(w, r, httprouter.Params{})

	fmt.Println(w.Body.String())
	return nil
}

// listAllWrapper displays all keys present in the database index
func listAllWrapper() {
	files := index.I.ListKeys()
	log.Success("found %d files in index:", len(files))

	for _, f := range files {
		log.Info(f)
	}
}

// lookupWrapper handles the lookup command, which retrieves and displays
// the value associated with a given key, resolving nested references
// up to the specified depth
func lookupWrapper(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("no key provided")
	}

	key := args[1]
	f, ok := index.I.Lookup(key)
	if !ok {
		return fmt.Errorf("key doesn't exist")
	}

	log.Success("found key %s:", key)

	m, err := f.ToMap()
	if err != nil {
		return err
	}

	depth := parseDepthFromArgs(args)
	log.Info("resolving reference to depth %d...", depth)
	resolvedMap := index.ResolveReferences(m, depth)

	// pretty print the JSON output
	b, err := json.Marshal(resolvedMap)
	if err != nil {
		return err
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, b, "", "\t")
	if err != nil {
		return err
	}

	log.Info("%s", prettyJSON.String())
	return nil
}

// deleteWrapper handles the delete command, which removes a key-value
// pair from the database
func deleteWrapper(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("no key provided")
	}

	key := args[1]
	f, ok := index.I.Lookup(key)
	if !ok {
		return fmt.Errorf("key doesn't exist")
	}

	err := index.I.Delete(f)
	if err != nil {
		return err
	}

	log.Success("deleted key %s", key)
	return nil
}
