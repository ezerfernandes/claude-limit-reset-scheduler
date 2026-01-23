package calendar

import (
	"errors"
	"testing"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
)

func TestValidateEventParams(t *testing.T) {
	tests := []struct {
		name    string
		params  EventParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid params",
			params: EventParams{
				Title:     "Test Event",
				StartTime: time.Now(),
				Duration:  30 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "valid params with all fields",
			params: EventParams{
				Title:       "Test Event",
				StartTime:   time.Now(),
				Duration:    1 * time.Hour,
				Description: "Test description",
				Location:    "Test location",
			},
			wantErr: false,
		},
		{
			name: "missing title",
			params: EventParams{
				StartTime: time.Now(),
				Duration:  30 * time.Minute,
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "zero start time",
			params: EventParams{
				Title:    "Test Event",
				Duration: 30 * time.Minute,
			},
			wantErr: true,
			errMsg:  "start time is required",
		},
		{
			name: "zero duration",
			params: EventParams{
				Title:     "Test Event",
				StartTime: time.Now(),
				Duration:  0,
			},
			wantErr: true,
			errMsg:  "duration must be positive",
		},
		{
			name: "negative duration",
			params: EventParams{
				Title:     "Test Event",
				StartTime: time.Now(),
				Duration:  -30 * time.Minute,
			},
			wantErr: true,
			errMsg:  "duration must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEventParams(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateEventParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("validateEventParams() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestParseEventResult(t *testing.T) {
	tests := []struct {
		name    string
		event   *calendar.Event
		want    *EventResult
		wantErr bool
	}{
		{
			name: "valid datetime event",
			event: &calendar.Event{
				Id:          "test-id-123",
				Summary:     "Test Event",
				Description: "Test description",
				Location:    "Test location",
				HtmlLink:    "https://calendar.google.com/event?id=test",
				Start: &calendar.EventDateTime{
					DateTime: "2024-01-15T14:00:00Z",
				},
				End: &calendar.EventDateTime{
					DateTime: "2024-01-15T15:00:00Z",
				},
			},
			want: &EventResult{
				ID:          "test-id-123",
				Title:       "Test Event",
				Description: "Test description",
				Location:    "Test location",
				Link:        "https://calendar.google.com/event?id=test",
			},
			wantErr: false,
		},
		{
			name: "all-day event",
			event: &calendar.Event{
				Id:       "test-id-456",
				Summary:  "All Day Event",
				HtmlLink: "https://calendar.google.com/event?id=test2",
				Start: &calendar.EventDateTime{
					Date: "2024-01-15",
				},
				End: &calendar.EventDateTime{
					Date: "2024-01-16",
				},
			},
			want: &EventResult{
				ID:    "test-id-456",
				Title: "All Day Event",
				Link:  "https://calendar.google.com/event?id=test2",
			},
			wantErr: false,
		},
		{
			name: "minimal event",
			event: &calendar.Event{
				Id:      "test-id-789",
				Summary: "Minimal Event",
				Start: &calendar.EventDateTime{
					DateTime: "2024-01-15T14:00:00Z",
				},
				End: &calendar.EventDateTime{
					DateTime: "2024-01-15T14:30:00Z",
				},
			},
			want: &EventResult{
				ID:    "test-id-789",
				Title: "Minimal Event",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseEventResult(tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseEventResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if got.ID != tt.want.ID {
				t.Errorf("parseEventResult() ID = %v, want %v", got.ID, tt.want.ID)
			}
			if got.Title != tt.want.Title {
				t.Errorf("parseEventResult() Title = %v, want %v", got.Title, tt.want.Title)
			}
			if got.Description != tt.want.Description {
				t.Errorf("parseEventResult() Description = %v, want %v", got.Description, tt.want.Description)
			}
			if got.Location != tt.want.Location {
				t.Errorf("parseEventResult() Location = %v, want %v", got.Location, tt.want.Location)
			}
			if got.Link != tt.want.Link {
				t.Errorf("parseEventResult() Link = %v, want %v", got.Link, tt.want.Link)
			}
		})
	}
}

func TestWrapAPIError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantErr    error
		wantErrMsg string
	}{
		{
			name:       "400 bad request",
			err:        &googleapi.Error{Code: 400, Message: "Invalid request"},
			wantErr:    ErrEventCreationFailed,
			wantErrMsg: "invalid request",
		},
		{
			name:       "401 unauthorized",
			err:        &googleapi.Error{Code: 401, Message: "Unauthorized"},
			wantErr:    ErrPermissionDenied,
			wantErrMsg: "authentication expired",
		},
		{
			name:       "403 forbidden",
			err:        &googleapi.Error{Code: 403, Message: "Forbidden"},
			wantErr:    ErrPermissionDenied,
			wantErrMsg: "don't have permission",
		},
		{
			name: "403 quota exceeded",
			err: &googleapi.Error{
				Code:    403,
				Message: "Quota exceeded",
				Errors: []googleapi.ErrorItem{
					{Reason: "quotaExceeded"},
				},
			},
			wantErr:    ErrQuotaExceeded,
			wantErrMsg: "try again later",
		},
		{
			name:       "404 not found",
			err:        &googleapi.Error{Code: 404, Message: "Not found"},
			wantErr:    ErrCalendarNotFound,
			wantErrMsg: "calendar ID",
		},
		{
			name:       "429 rate limited",
			err:        &googleapi.Error{Code: 429, Message: "Rate limited"},
			wantErr:    ErrQuotaExceeded,
			wantErrMsg: "too many requests",
		},
		{
			name:       "500 server error",
			err:        &googleapi.Error{Code: 500, Message: "Internal error"},
			wantErr:    ErrEventCreationFailed,
			wantErrMsg: "code: 500",
		},
		{
			name:       "non-API error",
			err:        errors.New("network error"),
			wantErr:    ErrEventCreationFailed,
			wantErrMsg: "network error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := wrapAPIError(tt.err)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("wrapAPIError() error type = %v, want %v", err, tt.wantErr)
			}
			if !contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("wrapAPIError() error = %q, want to contain %q", err.Error(), tt.wantErrMsg)
			}
		})
	}
}

func TestContainsQuotaError(t *testing.T) {
	tests := []struct {
		name   string
		apiErr *googleapi.Error
		want   bool
	}{
		{
			name: "quota exceeded",
			apiErr: &googleapi.Error{
				Code: 403,
				Errors: []googleapi.ErrorItem{
					{Reason: "quotaExceeded"},
				},
			},
			want: true,
		},
		{
			name: "rate limit exceeded",
			apiErr: &googleapi.Error{
				Code: 403,
				Errors: []googleapi.ErrorItem{
					{Reason: "rateLimitExceeded"},
				},
			},
			want: true,
		},
		{
			name: "user rate limit exceeded",
			apiErr: &googleapi.Error{
				Code: 403,
				Errors: []googleapi.ErrorItem{
					{Reason: "userRateLimitExceeded"},
				},
			},
			want: true,
		},
		{
			name: "other error",
			apiErr: &googleapi.Error{
				Code: 403,
				Errors: []googleapi.ErrorItem{
					{Reason: "forbidden"},
				},
			},
			want: false,
		},
		{
			name: "no errors array",
			apiErr: &googleapi.Error{
				Code: 403,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsQuotaError(tt.apiErr)
			if got != tt.want {
				t.Errorf("containsQuotaError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventParamsDefaults(t *testing.T) {
	// Test that EventParams with minimal fields can be used
	params := EventParams{
		Title:     "Test",
		StartTime: time.Now(),
		Duration:  30 * time.Minute,
	}

	// These should be their zero values
	if params.Description != "" {
		t.Error("Expected empty description by default")
	}
	if params.Location != "" {
		t.Error("Expected empty location by default")
	}
}

// contains checks if a string contains a substring (case-insensitive would need strings.Contains with ToLower).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
