// Package main is the entrypoint for the Bourbon Framework CLI.
package main

import (
	"os"

	"github.com/ishubhamsingh2e/bourbon/cmd/bourbon/internal/commands"
)

func main() {
	runner := commands.NewRunner(os.Stdout, os.Stderr)
	exitCode := runner.Run(os.Args[1:])
	os.Exit(exitCode)
}
