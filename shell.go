// provides the shell interface for smolDB, allowing interactive
// command-line operations on the database
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/themillenniumfalcon/smolDB/api"
	"github.com/themillenniumfalcon/smolDB/index"
	"github.com/themillenniumfalcon/smolDB/log"
)

// DefaultDepth defines the default depth for resolving nested references
// when no depth parameter is provided
const DefaultDepth = 0

// shell initializes and runs the interactive shell interface for smolDB,
// it takes a directory path as input where the database files are stored
func shell(dir string) error {
	log.IsShellMode = true
	log.Info("starting smoldb shell...")

	setup(dir)
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
