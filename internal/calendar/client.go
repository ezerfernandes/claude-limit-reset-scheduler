package calendar

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// Errors for calendar operations.
var (
	ErrEventCreationFailed = errors.New("failed to create event")
	ErrInvalidEventTime    = errors.New("invalid event time")
	ErrCalendarNotFound    = errors.New("calendar not found")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrQuotaExceeded       = errors.New("API quota exceeded")
)

// Client wraps the Google Calendar API service.
type Client struct {
	service    *calendar.Service
	calendarID string
}

// EventParams holds the parameters for creating a calendar event.
type EventParams struct {
	Title       string
	StartTime   time.Time
	Duration    time.Duration
	Description string
	Location    string
}

// EventResult contains the result of a successful event creation.
type EventResult struct {
	ID          string
	Title       string
	StartTime   time.Time
	EndTime     time.Time
	Description string
	Location    string
	Link        string
}

// NewClient creates a new Calendar client using the provided HTTP client.
// The httpClient should be configured with OAuth2 credentials.
func NewClient(ctx context.Context, httpClient *http.Client, calendarID string) (*Client, error) {
	service, err := calendar.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %w", err)
	}

	if calendarID == "" {
		calendarID = "primary"
	}

	return &Client{
		service:    service,
		calendarID: calendarID,
	}, nil
}

// CreateEvent creates a new event in the calendar.
func (c *Client) CreateEvent(ctx context.Context, params EventParams) (*EventResult, error) {
	if err := validateEventParams(params); err != nil {
		return nil, err
	}

	endTime := params.StartTime.Add(params.Duration)

	event := &calendar.Event{
		Summary:     params.Title,
		Description: params.Description,
		Location:    params.Location,
		Start: &calendar.EventDateTime{
			DateTime: params.StartTime.Format(time.RFC3339),
			TimeZone: params.StartTime.Location().String(),
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
			TimeZone: endTime.Location().String(),
		},
	}

	createdEvent, err := c.service.Events.Insert(c.calendarID, event).Context(ctx).Do()
	if err != nil {
		return nil, wrapAPIError(err)
	}

	return parseEventResult(createdEvent)
}

// validateEventParams validates the event parameters.
func validateEventParams(params EventParams) error {
	if params.Title == "" {
		return fmt.Errorf("%w: title is required", ErrInvalidEventTime)
	}

	if params.StartTime.IsZero() {
		return fmt.Errorf("%w: start time is required", ErrInvalidEventTime)
	}

	if params.Duration <= 0 {
		return fmt.Errorf("%w: duration must be positive", ErrInvalidEventTime)
	}

	return nil
}

// parseEventResult converts a Google Calendar event to our EventResult type.
func parseEventResult(event *calendar.Event) (*EventResult, error) {
	startTime, err := time.Parse(time.RFC3339, event.Start.DateTime)
	if err != nil {
		// Try parsing as date-only for all-day events
		startTime, err = time.Parse("2006-01-02", event.Start.Date)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start time: %w", err)
		}
	}

	endTime, err := time.Parse(time.RFC3339, event.End.DateTime)
	if err != nil {
		// Try parsing as date-only for all-day events
		endTime, err = time.Parse("2006-01-02", event.End.Date)
		if err != nil {
			return nil, fmt.Errorf("failed to parse end time: %w", err)
		}
	}

	return &EventResult{
		ID:          event.Id,
		Title:       event.Summary,
		StartTime:   startTime,
		EndTime:     endTime,
		Description: event.Description,
		Location:    event.Location,
		Link:        event.HtmlLink,
	}, nil
}

// wrapAPIError wraps Google API errors with user-friendly messages.
func wrapAPIError(err error) error {
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		switch apiErr.Code {
		case 400:
			return fmt.Errorf("%w: invalid request - %s", ErrEventCreationFailed, apiErr.Message)
		case 401:
			return fmt.Errorf("%w: authentication expired, please re-authenticate", ErrPermissionDenied)
		case 403:
			if containsQuotaError(apiErr) {
				return fmt.Errorf("%w: please try again later", ErrQuotaExceeded)
			}
			return fmt.Errorf("%w: you don't have permission to access this calendar", ErrPermissionDenied)
		case 404:
			return fmt.Errorf("%w: check that the calendar ID is correct", ErrCalendarNotFound)
		case 429:
			return fmt.Errorf("%w: too many requests, please try again later", ErrQuotaExceeded)
		default:
			return fmt.Errorf("%w: %s (code: %d)", ErrEventCreationFailed, apiErr.Message, apiErr.Code)
		}
	}

	return fmt.Errorf("%w: %v", ErrEventCreationFailed, err)
}

// containsQuotaError checks if the API error is related to quota.
func containsQuotaError(apiErr *googleapi.Error) bool {
	for _, e := range apiErr.Errors {
		if e.Reason == "quotaExceeded" || e.Reason == "rateLimitExceeded" || e.Reason == "userRateLimitExceeded" {
			return true
		}
	}
	return false
}
