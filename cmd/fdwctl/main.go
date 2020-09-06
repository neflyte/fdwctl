package main

import (
	"github.com/neflyte/fdwctl/cmd/fdwctl/cmd"
	"os"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
