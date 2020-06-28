package main

import (
	"github.com/neflyte/fdwctl/cmd/fdwctl/cmd"
	"github.com/neflyte/fdwctl/internal/logger"
)

func main() {
	log := logger.Root().WithField("function", "main")
	err := cmd.Execute()
	if err != nil {
		log.Errorf("program error: %s", err)
	}
}
