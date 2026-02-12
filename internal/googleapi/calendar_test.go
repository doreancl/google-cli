package googleapi

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	calendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func mustCalendarService(t *testing.T, handler http.HandlerFunc) (*calendar.Service, func()) {
	t.Helper()
	ts := httptest.NewServer(handler)
	svc, err := calendar.NewService(context.Background(),
		option.WithHTTPClient(ts.Client()),
		option.WithEndpoint(ts.URL+"/"),
		option.WithoutAuthentication(),
	)
	if err != nil {
		ts.Close()
		t.Fatalf("new service: %v", err)
	}
	return svc, ts.Close
}

func TestNewCalendarClientNilClientChecks(t *testing.T) {
	var c *CalendarClient
	_, err := c.ListEvents(context.Background(), ListEventsRequest{})
	if err == nil {
		t.Fatal("expected nil client error")
	}
	_, err = c.CreateEvent(context.Background(), CreateEventRequest{})
	if err == nil {
		t.Fatal("expected nil client error")
	}
}

func TestNewCalendarClient(t *testing.T) {
	orig := newCalendarServiceFn
	defer func() { newCalendarServiceFn = orig }()

	newCalendarServiceFn = func(context.Context) (*calendar.Service, error) {
		return nil, errors.New("svc boom")
	}
	if _, err := NewCalendarClient(context.Background()); err == nil {
		t.Fatal("expected service creation error")
	}

	svc, closeFn := mustCalendarService(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"items":[]}`)
	})
	defer closeFn()
	newCalendarServiceFn = func(context.Context) (*calendar.Service, error) { return svc, nil }
	c, err := NewCalendarClient(context.Background())
	if err != nil || c == nil || c.svc == nil {
		t.Fatalf("expected initialized client, got c=%v err=%v", c, err)
	}
}

func TestListEventsValidations(t *testing.T) {
	c := &CalendarClient{}
	_, err := c.ListEvents(context.Background(), ListEventsRequest{From: time.Now(), To: time.Now().Add(-time.Hour)})
	if err == nil {
		t.Fatal("expected invalid range error")
	}
}

func TestCreateEventValidations(t *testing.T) {
	c := &CalendarClient{}
	_, err := c.CreateEvent(context.Background(), CreateEventRequest{Summary: "", StartRFC3339: "x", EndRFC3339: "y"})
	if err == nil {
		t.Fatal("expected summary error")
	}
	_, err = c.CreateEvent(context.Background(), CreateEventRequest{Summary: "s", StartRFC3339: "bad", EndRFC3339: "2026-01-01T10:00:00Z"})
	if err == nil {
		t.Fatal("expected invalid start")
	}
	_, err = c.CreateEvent(context.Background(), CreateEventRequest{Summary: "s", StartRFC3339: "2026-01-01T10:00:00Z", EndRFC3339: "bad"})
	if err == nil {
		t.Fatal("expected invalid end")
	}
	_, err = c.CreateEvent(context.Background(), CreateEventRequest{Summary: "s", StartRFC3339: "2026-01-01T10:00:00Z", EndRFC3339: "2026-01-01T09:00:00Z"})
	if err == nil {
		t.Fatal("expected end before start")
	}
}

func TestListEventsSuccess(t *testing.T) {
	svc, closeFn := mustCalendarService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.Contains(r.URL.Path, "/events") {
			http.NotFound(w, r)
			return
		}
		_, _ = io.WriteString(w, `{"items":[{"id":"1","summary":"A","start":{"dateTime":"2026-01-01T10:00:00Z"},"end":{"dateTime":"2026-01-01T11:00:00Z"},"htmlLink":"https://x"},{"id":"2","summary":"B","start":{"date":"2026-01-02"},"end":{"date":"2026-01-03"}}]}`)
	})
	defer closeFn()

	c := &CalendarClient{svc: svc}
	out, err := c.ListEvents(context.Background(), ListEventsRequest{
		CalendarID: "",
		From:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		To:         time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}
	if len(out) != 2 || out[0].ID != "1" || out[1].Start != "2026-01-02" {
		t.Fatalf("unexpected output: %+v", out)
	}
}

func TestListEventsServiceError(t *testing.T) {
	svc, closeFn := mustCalendarService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":"boom"}`)
	})
	defer closeFn()

	c := &CalendarClient{svc: svc}
	_, err := c.ListEvents(context.Background(), ListEventsRequest{
		From: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
	})
	if err == nil {
		t.Fatal("expected api error")
	}
}

func TestCreateEventSuccess(t *testing.T) {
	svc, closeFn := mustCalendarService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.Contains(r.URL.Path, "/events") {
			http.NotFound(w, r)
			return
		}
		_, _ = io.WriteString(w, `{"id":"evt-1","summary":"Daily","start":{"dateTime":"2026-01-01T10:00:00Z"},"end":{"dateTime":"2026-01-01T10:30:00Z"},"htmlLink":"https://event"}`)
	})
	defer closeFn()

	c := &CalendarClient{svc: svc}
	out, err := c.CreateEvent(context.Background(), CreateEventRequest{
		Summary:      "Daily",
		StartRFC3339: "2026-01-01T10:00:00Z",
		EndRFC3339:   "2026-01-01T10:30:00Z",
	})
	if err != nil {
		t.Fatalf("CreateEvent: %v", err)
	}
	if out.ID != "evt-1" || out.HTMLLink != "https://event" {
		t.Fatalf("unexpected output: %+v", out)
	}
}

func TestCreateEventServiceError(t *testing.T) {
	svc, closeFn := mustCalendarService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":"boom"}`)
	})
	defer closeFn()

	c := &CalendarClient{svc: svc}
	_, err := c.CreateEvent(context.Background(), CreateEventRequest{
		Summary:      "Daily",
		StartRFC3339: "2026-01-01T10:00:00Z",
		EndRFC3339:   "2026-01-01T10:30:00Z",
	})
	if err == nil {
		t.Fatal("expected api error")
	}
}

func TestOauthRetryTransportFallbackAndErrors(t *testing.T) {
	errBoom := errors.New("boom")
	tpt := &oauthRetryTransport{fallbackBase: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errBoom
	})}
	resp, err := tpt.RoundTrip(httptest.NewRequest(http.MethodGet, "http://example.com", nil))
	if resp != nil && resp.Body != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	if !errors.Is(err, errBoom) {
		t.Fatalf("expected boom, got %v", err)
	}
}

func TestOauthRetryTransportNoRetryForNonGET(t *testing.T) {
	var calls int32
	tpt := &oauthRetryTransport{base: roundTripFunc(func(*http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return &http.Response{StatusCode: http.StatusInternalServerError, Body: io.NopCloser(strings.NewReader("x")), Header: make(http.Header)}, nil
	})}
	resp, err := tpt.RoundTrip(httptest.NewRequest(http.MethodPost, "http://example.com", nil))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestOauthRetryTransportRetriesGET(t *testing.T) {
	var calls int32
	tpt := &oauthRetryTransport{base: roundTripFunc(func(*http.Request) (*http.Response, error) {
		c := atomic.AddInt32(&calls, 1)
		if c < 3 {
			return &http.Response{StatusCode: http.StatusTooManyRequests, Body: io.NopCloser(strings.NewReader("retry")), Header: make(http.Header)}, nil
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("ok")), Header: make(http.Header)}, nil
	})}
	resp, err := tpt.RoundTrip(httptest.NewRequest(http.MethodGet, "http://example.com", nil))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}
