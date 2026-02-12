package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"google-cli/internal/googleapi"
)

var newCalendarClient = googleapi.NewCalendarClient

func runEvents(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("falta subcomando para events: list | create")
	}
	switch args[0] {
	case "list":
		return runEventsList(ctx, args[1:])
	case "create":
		return runEventsCreate(ctx, args[1:])
	default:
		return fmt.Errorf("subcomando no soportado: %s", args[0])
	}
}

func runEventsList(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("events list", flag.ContinueOnError)
	calendarID := fs.String("calendar", "primary", "ID del calendario")
	days := fs.Int("days", 7, "Numero de dias hacia adelante")
	if err := fs.Parse(args); err != nil {
		return &ExitError{Code: 2, Err: err}
	}

	if *days <= 0 {
		return errors.New("--days debe ser mayor a 0")
	}

	client, err := newCalendarClient(ctx)
	if err != nil {
		return err
	}

	start := time.Now().UTC()
	end := start.Add(time.Duration(*days) * 24 * time.Hour)
	events, err := client.ListEvents(ctx, googleapi.ListEventsRequest{
		CalendarID: *calendarID,
		From:       start,
		To:         end,
	})
	if err != nil {
		return err
	}
	if len(events) == 0 {
		_, _ = fmt.Fprintln(os.Stdout, "No hay eventos en el rango indicado.")
		return nil
	}

	for _, item := range events {
		_, _ = fmt.Fprintf(os.Stdout, "- %s | %s | %s\n", item.Start, item.Summary, item.ID)
	}
	return nil
}

func runEventsCreate(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("events create", flag.ContinueOnError)
	calendarID := fs.String("calendar", "primary", "ID del calendario")
	summary := fs.String("summary", "", "Titulo del evento")
	description := fs.String("description", "", "Descripcion del evento")
	start := fs.String("start", "", "Inicio RFC3339 (ej: 2026-02-12T10:00:00-06:00)")
	end := fs.String("end", "", "Fin RFC3339 (ej: 2026-02-12T10:30:00-06:00)")
	location := fs.String("location", "", "Ubicacion")
	if err := fs.Parse(args); err != nil {
		return &ExitError{Code: 2, Err: err}
	}

	if *summary == "" || *start == "" || *end == "" {
		return errors.New("flags requeridos: --summary --start --end")
	}

	client, err := newCalendarClient(ctx)
	if err != nil {
		return err
	}

	created, err := client.CreateEvent(ctx, googleapi.CreateEventRequest{
		CalendarID:   *calendarID,
		Summary:      *summary,
		Description:  *description,
		Location:     *location,
		StartRFC3339: *start,
		EndRFC3339:   *end,
	})
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(os.Stdout, "Evento creado: %s\nLink: %s\n", created.ID, created.HTMLLink)
	return nil
}
