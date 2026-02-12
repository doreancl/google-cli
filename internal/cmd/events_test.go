package cmd

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"google-cli/internal/googleapi"
)

type mockCalendarClient struct {
	listOut   []googleapi.Event
	listErr   error
	createOut *googleapi.Event
	createErr error
}

func (m *mockCalendarClient) ListEvents(context.Context, googleapi.ListEventsRequest) ([]googleapi.Event, error) {
	return m.listOut, m.listErr
}

func (m *mockCalendarClient) CreateEvent(context.Context, googleapi.CreateEventRequest) (*googleapi.Event, error) {
	return m.createOut, m.createErr
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	fn()
	_ = w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func TestRunEventsDispatch(t *testing.T) {
	if err := runEvents(context.Background(), nil); err == nil {
		t.Fatal("expected missing subcommand error")
	}
	if err := runEvents(context.Background(), []string{"nope"}); err == nil {
		t.Fatal("expected unsupported subcommand error")
	}
}

func TestRunEventsListValidation(t *testing.T) {
	if err := runEventsList(context.Background(), []string{"--days", "0"}); err == nil || !strings.Contains(err.Error(), "--days") {
		t.Fatalf("expected --days error, got %v", err)
	}
	err := runEventsList(context.Background(), []string{"--bad-flag"})
	if err == nil {
		t.Fatal("expected parse error")
	}
	var ee *ExitError
	if !errors.As(err, &ee) {
		t.Fatalf("expected ExitError, got %T", err)
	}
}

func TestRunEventsCreateValidation(t *testing.T) {
	if err := runEventsCreate(context.Background(), nil); err == nil || !strings.Contains(err.Error(), "flags requeridos") {
		t.Fatalf("expected required flags error, got %v", err)
	}
	err := runEventsCreate(context.Background(), []string{"--bad-flag"})
	if err == nil {
		t.Fatal("expected parse error")
	}
	var ee *ExitError
	if !errors.As(err, &ee) {
		t.Fatalf("expected ExitError, got %T", err)
	}
}

func TestRunEventsListNoEvents(t *testing.T) {
	old := newCalendarClient
	defer func() { newCalendarClient = old }()

	newCalendarClient = func(context.Context) (calendarClient, error) {
		return &mockCalendarClient{listOut: []googleapi.Event{}}, nil
	}

	out := captureStdout(t, func() {
		if err := runEventsList(context.Background(), []string{"--days", "1"}); err != nil {
			t.Fatalf("runEventsList: %v", err)
		}
	})
	if !strings.Contains(out, "No hay eventos") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRunEventsListAndCreateSuccess(t *testing.T) {
	old := newCalendarClient
	defer func() { newCalendarClient = old }()

	newCalendarClient = func(context.Context) (calendarClient, error) {
		return &mockCalendarClient{
			listOut:   []googleapi.Event{{ID: "1", Summary: "Daily", Start: "2026-01-01T10:00:00Z"}},
			createOut: &googleapi.Event{ID: "evt-1", HTMLLink: "https://event"},
		}, nil
	}

	outList := captureStdout(t, func() {
		if err := runEventsList(context.Background(), []string{"--days", "1", "--calendar", "primary"}); err != nil {
			t.Fatalf("runEventsList: %v", err)
		}
	})
	if !strings.Contains(outList, "Daily") {
		t.Fatalf("unexpected list output: %q", outList)
	}

	outCreate := captureStdout(t, func() {
		err := runEventsCreate(context.Background(), []string{
			"--summary", "Daily",
			"--start", time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
			"--end", time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC).Format(time.RFC3339),
		})
		if err != nil {
			t.Fatalf("runEventsCreate: %v", err)
		}
	})
	if !strings.Contains(outCreate, "Evento creado: evt-1") {
		t.Fatalf("unexpected create output: %q", outCreate)
	}
}

func TestRunEventsClientErrors(t *testing.T) {
	old := newCalendarClient
	defer func() { newCalendarClient = old }()

	newCalendarClient = func(context.Context) (calendarClient, error) {
		return nil, errors.New("boom")
	}
	if err := runEventsList(context.Background(), []string{"--days", "1"}); err == nil {
		t.Fatal("expected list client error")
	}
	if err := runEventsCreate(context.Background(), []string{
		"--summary", "Daily",
		"--start", "2026-01-01T10:00:00Z",
		"--end", "2026-01-01T10:30:00Z",
	}); err == nil {
		t.Fatal("expected create client error")
	}
}

func TestRunEventsListAndCreateSubcommands(t *testing.T) {
	old := newCalendarClient
	defer func() { newCalendarClient = old }()

	newCalendarClient = func(context.Context) (calendarClient, error) {
		return &mockCalendarClient{
			listOut:   []googleapi.Event{},
			createOut: &googleapi.Event{ID: "evt-1"},
		}, nil
	}

	if err := runEvents(context.Background(), []string{"list", "--days", "1"}); err != nil {
		t.Fatalf("runEvents list: %v", err)
	}
	if err := runEvents(context.Background(), []string{
		"create",
		"--summary", "Daily",
		"--start", "2026-01-01T10:00:00Z",
		"--end", "2026-01-01T10:30:00Z",
	}); err != nil {
		t.Fatalf("runEvents create: %v", err)
	}
}
