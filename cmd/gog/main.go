package main

import (
	"fmt"
	"os"

	"google-cli/internal/cmd"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
)

var (
	_ = version
	_ = commit
	_ = date
)

func main() {
	if err := cmd.Execute(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(cmd.ExitCode(err))
	}
}
