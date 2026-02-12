package cmd

import (
	"context"
	"errors"
	"testing"
)

func TestExecuteUsageAndHelp(t *testing.T) {
	if err := Execute(nil); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if err := Execute([]string{"--help"}); err != nil {
		t.Fatalf("expected nil help, got %v", err)
	}
	if err := Execute([]string{"-h"}); err != nil {
		t.Fatalf("expected nil -h, got %v", err)
	}
	if err := Execute([]string{"--", "--help"}); err != nil {
		t.Fatalf("expected nil with -- passthrough, got %v", err)
	}
}

func TestExecuteUnsupportedCommand(t *testing.T) {
	err := Execute([]string{"nope"})
	if err == nil {
		t.Fatal("expected error")
	}
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 2 {
		t.Fatalf("expected ExitError code 2, got %#v", err)
	}
}

func TestExecuteDispatchesCommands(t *testing.T) {
	origAuth := runAuthCommand
	origEvents := runEventsCommand
	defer func() {
		runAuthCommand = origAuth
		runEventsCommand = origEvents
	}()

	runAuthCommand = func(context.Context, []string) error { return errors.New("auth path") }
	err := Execute([]string{"auth"})
	if err == nil || err.Error() != "auth path" {
		t.Fatalf("expected auth dispatch, got %v", err)
	}

	runEventsCommand = func(context.Context, []string) error { return errors.New("events path") }
	err = Execute([]string{"events"})
	if err == nil || err.Error() != "events path" {
		t.Fatalf("expected events dispatch, got %v", err)
	}
}
