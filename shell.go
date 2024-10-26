package main

import (
	"bufio"
	"os"
	"strings"

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

func execInput(input string) error {
	args := strings.Split(input, " ")
	log.Info("%+v", args)
	return nil
}
