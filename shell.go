package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/themillenniumfalcon/smolDB/index"
	"github.com/themillenniumfalcon/smolDB/log"
)

func shell(dir string) error {
	log.IsShellMode = true
	log.Info("starting nanodb shell...")

	setup(dir)
	reader := bufio.NewReader(os.Stdin)

	for {
		log.Prompt("nanodb> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			log.Warn("err reading input: %s", err.Error())
		}

		if err = execInput(input); err != nil {
			log.Warn("err executing input: %s", err.Error())
		}
	}
}

func execInput(input string) (err error) {
	input = strings.TrimSuffix(input, "\n")
	args := strings.Split(input, " ")

	switch args[0] {
	case "index":
		files := index.I.ListKeys()
		log.Success("found %d files in index:", len(files))
		for _, f := range files {
			log.Info(f)
		}
	default:
		log.Warn("'%s' is not a valid command.", args[0])
		log.Info("valid commands: index, lookup <key>, delete <key>, exit")
	}

	return err
}
