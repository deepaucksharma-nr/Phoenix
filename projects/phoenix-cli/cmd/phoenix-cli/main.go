package main

import (
	"os"

	"github.com/phoenix/platform/projects/phoenix-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
