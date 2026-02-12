package cmd

import "errors"

// ExitError carries a specific process exit code.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e == nil || e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func (e *ExitError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// ExitCode maps command errors to process exit codes.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var ee *ExitError
	if errors.As(err, &ee) && ee.Code > 0 {
		return ee.Code
	}
	return 1
}
