package cmd

import (
	"context"
	"fmt"
	"os"
)

func Execute(args []string) error {
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}

	if len(args) == 0 {
		usage()
		return nil
	}

	if args[0] == "--help" || args[0] == "-h" {
		usage()
		return nil
	}

	switch args[0] {
	case "auth":
		return runAuth(context.Background(), args[1:])
	case "events":
		return runEvents(context.Background(), args[1:])
	default:
		usage()
		return &ExitError{Code: 2, Err: fmt.Errorf("comando no soportado: %s", args[0])}
	}
}

func usage() {
	_, _ = fmt.Fprint(os.Stdout, `Google Calendar CLI

Uso:
  dorean_g auth --credentials ./client_secret.json
  dorean_g events list [--calendar primary] [--days 7]
  dorean_g events create --summary "Daily" --start "2026-02-12T10:00:00-06:00" --end "2026-02-12T10:30:00-06:00" [--calendar primary] [--description "..."]

Variables:
  GCAL_CREDENTIALS  Ruta de client_secret.json (opcional si usas --credentials)
`)
}
