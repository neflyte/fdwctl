package main

import (
	"fmt"
	"github.com/neflyte/fdwctl/cmd/fdwctl/cmd"
	"github.com/neflyte/fdwctl/internal/logger"
)

var AppVersion string

func main() {
	log := logger.Root().WithField("function", "main")
	fmt.Printf("fdwctl v%s\n", AppVersion)
	err := cmd.Execute()
	if err != nil {
		log.Errorf("program error: %s", err)
	}
}
