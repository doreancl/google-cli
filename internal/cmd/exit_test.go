package cmd

import (
	"errors"
	"testing"
)

func TestExitErrorMethods(t *testing.T) {
	var nilErr *ExitError
	if nilErr.Error() != "" {
		t.Fatal("nil ExitError should render empty error string")
	}
	if nilErr.Unwrap() != nil {
		t.Fatal("nil ExitError should unwrap nil")
	}

	err := &ExitError{Code: 2, Err: errors.New("boom")}
	if err.Error() != "boom" {
		t.Fatalf("unexpected Error: %q", err.Error())
	}
	if !errors.Is(err, err.Err) {
		t.Fatal("unwrap should expose wrapped error")
	}
}

func TestExitCode(t *testing.T) {
	if got := ExitCode(nil); got != 0 {
		t.Fatalf("got %d", got)
	}
	if got := ExitCode(errors.New("x")); got != 1 {
		t.Fatalf("got %d", got)
	}
	if got := ExitCode(&ExitError{Code: 3, Err: errors.New("x")}); got != 3 {
		t.Fatalf("got %d", got)
	}
	if got := ExitCode(&ExitError{Code: 0, Err: errors.New("x")}); got != 1 {
		t.Fatalf("got %d", got)
	}
}
