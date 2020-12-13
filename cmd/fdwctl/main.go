package main

import (
	"os"

	"github.com/neflyte/fdwctl/cmd/fdwctl/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
