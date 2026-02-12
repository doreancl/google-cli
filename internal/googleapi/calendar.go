package googleapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	calendar "google.golang.org/api/calendar/v3"
)

type CalendarClient struct {
	svc *calendar.Service
}

type ListEventsRequest struct {
	CalendarID string
	From       time.Time
	To         time.Time
}

type Event struct {
	ID       string `json:"id"`
	Summary  string `json:"summary"`
	Start    string `json:"start"`
	End      string `json:"end"`
	HTMLLink string `json:"htmlLink,omitempty"`
}

type CreateEventRequest struct {
	CalendarID   string
	Summary      string
	Description  string
	Location     string
	StartRFC3339 string
	EndRFC3339   string
}

var newCalendarServiceFn = func(ctx context.Context) (*calendar.Service, error) {
	return newService(ctx, calendar.NewService)
}

func NewCalendarClient(ctx context.Context) (*CalendarClient, error) {
	svc, err := newCalendarServiceFn(ctx)
	if err != nil {
		return nil, fmt.Errorf("calendar service: %w", err)
	}

	return &CalendarClient{svc: svc}, nil
}

func (c *CalendarClient) ListEvents(ctx context.Context, req ListEventsRequest) ([]Event, error) {
	if c == nil || c.svc == nil {
		return nil, errors.New("calendar client no inicializado")
	}

	calendarID := strings.TrimSpace(req.CalendarID)
	if calendarID == "" {
		calendarID = "primary"
	}
	if req.To.Before(req.From) {
		return nil, errors.New("rango invalido: --to debe ser mayor o igual a --from")
	}

	resp, err := c.svc.Events.List(calendarID).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(req.From.Format(time.RFC3339)).
		TimeMax(req.To.Format(time.RFC3339)).
		OrderBy("startTime").
		Fields("items(id,summary,start,end,htmlLink)").
		Context(ctx).
		Do()
	if err != nil {
		return nil, err
	}

	out := make([]Event, 0, len(resp.Items))
	for _, item := range resp.Items {
		start := item.Start.DateTime
		if start == "" {
			start = item.Start.Date
		}
		end := item.End.DateTime
		if end == "" {
			end = item.End.Date
		}
		out = append(out, Event{
			ID:       item.Id,
			Summary:  item.Summary,
			Start:    start,
			End:      end,
			HTMLLink: item.HtmlLink,
		})
	}
	return out, nil
}

func (c *CalendarClient) CreateEvent(ctx context.Context, req CreateEventRequest) (*Event, error) {
	if c == nil || c.svc == nil {
		return nil, errors.New("calendar client no inicializado")
	}

	calendarID := strings.TrimSpace(req.CalendarID)
	if calendarID == "" {
		calendarID = "primary"
	}
	if strings.TrimSpace(req.Summary) == "" {
		return nil, errors.New("summary requerido")
	}
	start, err := time.Parse(time.RFC3339, req.StartRFC3339)
	if err != nil {
		return nil, fmt.Errorf("start invalido: %w", err)
	}
	end, err := time.Parse(time.RFC3339, req.EndRFC3339)
	if err != nil {
		return nil, fmt.Errorf("end invalido: %w", err)
	}
	if !end.After(start) {
		return nil, errors.New("end debe ser mayor a start")
	}

	created, err := c.svc.Events.Insert(calendarID, &calendar.Event{
		Summary:     req.Summary,
		Description: req.Description,
		Location:    req.Location,
		Start:       &calendar.EventDateTime{DateTime: req.StartRFC3339},
		End:         &calendar.EventDateTime{DateTime: req.EndRFC3339},
	}).
		Fields("id,summary,start,end,htmlLink").
		Context(ctx).
		Do()
	if err != nil {
		return nil, err
	}

	startOut := created.Start.DateTime
	if startOut == "" {
		startOut = created.Start.Date
	}
	endOut := created.End.DateTime
	if endOut == "" {
		endOut = created.End.Date
	}
	return &Event{
		ID:       created.Id,
		Summary:  created.Summary,
		Start:    startOut,
		End:      endOut,
		HTMLLink: created.HtmlLink,
	}, nil
}

type oauthRetryTransport struct {
	base         http.RoundTripper
	fallbackBase http.RoundTripper
}

func (t *oauthRetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.base
	if base == nil {
		base = t.fallbackBase
	}
	if base == nil {
		base = http.DefaultTransport
	}

	maxRetries := 2
	if req != nil && req.Method != http.MethodGet {
		maxRetries = 0
	}
	delay := 500 * time.Millisecond
	var resp *http.Response
	var err error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err = base.RoundTrip(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			if attempt == maxRetries {
				return resp, nil
			}
			_ = resp.Body.Close()
			time.Sleep(delay)
			delay *= 2
			continue
		}
		return resp, nil
	}
	return resp, err
}
