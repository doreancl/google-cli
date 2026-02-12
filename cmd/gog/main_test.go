package main

import (
	"errors"
	"os"
	"testing"
)

var errBoom = errors.New("boom")

func TestMainHelpPath(t *testing.T) {
	t.Helper()
	origExec := executeFn
	origExit := exitFn
	defer func() {
		executeFn = origExec
		exitFn = origExit
	}()

	origArgs := os.Args
	os.Args = []string{"gog", "--help"}
	defer func() { os.Args = origArgs }()

	executeFn = func([]string) error { return nil }
	exitFn = func(int) {}
	main()
}

func TestMainErrorPath(t *testing.T) {
	origExec := executeFn
	origExit := exitFn
	defer func() {
		executeFn = origExec
		exitFn = origExit
	}()

	var exitCode int
	executeFn = func([]string) error { return errBoom }
	exitFn = func(code int) { exitCode = code }

	main()
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
}
