package main

import (
	"fmt"
	"os"

	"google-cli/internal/cmd"
)

var (
	executeFn  = cmd.Execute
	exitCodeFn = cmd.ExitCode
	exitFn     = os.Exit
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
	if err := executeFn(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		exitFn(exitCodeFn(err))
	}
}
