package main

import "github.com/neflyte/fdwctl/cmd/fdwctl/cmd"

func main() {
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
