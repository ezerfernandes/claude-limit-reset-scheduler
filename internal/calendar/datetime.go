package calendar

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
)

// Errors for date/time parsing.
var (
	ErrInvalidDateFormat = errors.New("invalid date/time format")
	ErrInvalidTimezone   = errors.New("invalid timezone")
)

// ParseTime parses a date/time string into a time.Time value.
// Supported formats:
//   - ISO 8601: "2024-01-15T14:00:00", "2024-01-15T14:00:00Z", "2024-01-15T14:00:00+05:00"
//   - Natural: "2024-01-15 14:00", "2024-01-15 14:00:00"
//   - Time only: "14:00", "14:00:00" (assumes today)
//   - Relative: "tomorrow 14:00", "today 14:00", "in 2 hours", "in 30 minutes"
//
// The timezone is determined by:
//  1. Timezone embedded in the input string (for ISO 8601 with offset)
//  2. The provided timezone string (from config)
//  3. The TZ environment variable
//  4. The system's local timezone
func ParseTime(input string, timezone string) (time.Time, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return time.Time{}, fmt.Errorf("%w: empty input", ErrInvalidDateFormat)
	}

	// Get the location to use for parsing
	loc, err := getLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}

	// Try relative formats first
	if t, ok := parseRelative(input, loc); ok {
		return t, nil
	}

	// Try time-only format
	if t, ok := parseTimeOnly(input, loc); ok {
		return t, nil
	}

	// Try standard formats using dateparse
	return parseStandard(input, loc)
}

// getLocation returns the time.Location based on the provided timezone string,
// falling back to TZ environment variable, then system local timezone.
func getLocation(timezone string) (*time.Location, error) {
	// Priority 1: Provided timezone string
	if timezone != "" {
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidTimezone, timezone)
		}
		return loc, nil
	}

	// Priority 2: TZ environment variable
	if tz := os.Getenv("TZ"); tz != "" {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			return nil, fmt.Errorf("%w: %s (from TZ environment variable)", ErrInvalidTimezone, tz)
		}
		return loc, nil
	}

	// Priority 3: System local timezone
	return time.Local, nil
}

// parseRelative attempts to parse relative date/time formats.
// Supported formats:
//   - "today 14:00", "today at 14:00"
//   - "tomorrow 14:00", "tomorrow at 14:00"
//   - "in 2 hours", "in 30 minutes", "in 1 hour"
func parseRelative(input string, loc *time.Location) (time.Time, bool) {
	input = strings.ToLower(input)
	now := time.Now().In(loc)

	// Pattern: "in X hours/minutes"
	if strings.HasPrefix(input, "in ") {
		if t, ok := parseInDuration(input, now); ok {
			return t, true
		}
	}

	// Pattern: "today [at] HH:MM"
	if strings.HasPrefix(input, "today") {
		if t, ok := parseDayWithTime(input, now, 0, loc); ok {
			return t, true
		}
	}

	// Pattern: "tomorrow [at] HH:MM"
	if strings.HasPrefix(input, "tomorrow") {
		if t, ok := parseDayWithTime(input, now, 1, loc); ok {
			return t, true
		}
	}

	return time.Time{}, false
}

// parseInDuration parses "in X hours/minutes" format.
var inDurationRegex = regexp.MustCompile(`^in\s+(\d+)\s*(hours?|minutes?|mins?|hrs?)$`)

func parseInDuration(input string, now time.Time) (time.Time, bool) {
	matches := inDurationRegex.FindStringSubmatch(input)
	if matches == nil {
		return time.Time{}, false
	}

	amount, err := strconv.Atoi(matches[1])
	if err != nil {
		return time.Time{}, false
	}

	unit := matches[2]
	var duration time.Duration

	switch {
	case strings.HasPrefix(unit, "hour"), strings.HasPrefix(unit, "hr"):
		duration = time.Duration(amount) * time.Hour
	case strings.HasPrefix(unit, "min"):
		duration = time.Duration(amount) * time.Minute
	default:
		return time.Time{}, false
	}

	return now.Add(duration), true
}

// parseDayWithTime parses "today/tomorrow [at] HH:MM" format.
var dayTimeRegex = regexp.MustCompile(`^(?:today|tomorrow)\s*(?:at\s+)?(\d{1,2}):(\d{2})(?::(\d{2}))?$`)

func parseDayWithTime(input string, now time.Time, daysOffset int, loc *time.Location) (time.Time, bool) {
	matches := dayTimeRegex.FindStringSubmatch(input)
	if matches == nil {
		return time.Time{}, false
	}

	hour, err := strconv.Atoi(matches[1])
	if err != nil || hour < 0 || hour > 23 {
		return time.Time{}, false
	}

	minute, err := strconv.Atoi(matches[2])
	if err != nil || minute < 0 || minute > 59 {
		return time.Time{}, false
	}

	second := 0
	if matches[3] != "" {
		second, err = strconv.Atoi(matches[3])
		if err != nil || second < 0 || second > 59 {
			return time.Time{}, false
		}
	}

	targetDate := now.AddDate(0, 0, daysOffset)
	return time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(),
		hour, minute, second, 0, loc), true
}

// parseTimeOnly attempts to parse time-only formats like "14:00" or "14:00:00".
// Returns a time.Time for today at the specified time.
var timeOnlyRegex = regexp.MustCompile(`^(\d{1,2}):(\d{2})(?::(\d{2}))?$`)

func parseTimeOnly(input string, loc *time.Location) (time.Time, bool) {
	matches := timeOnlyRegex.FindStringSubmatch(input)
	if matches == nil {
		return time.Time{}, false
	}

	hour, err := strconv.Atoi(matches[1])
	if err != nil || hour < 0 || hour > 23 {
		return time.Time{}, false
	}

	minute, err := strconv.Atoi(matches[2])
	if err != nil || minute < 0 || minute > 59 {
		return time.Time{}, false
	}

	second := 0
	if matches[3] != "" {
		second, err = strconv.Atoi(matches[3])
		if err != nil || second < 0 || second > 59 {
			return time.Time{}, false
		}
	}

	now := time.Now().In(loc)
	return time.Date(now.Year(), now.Month(), now.Day(),
		hour, minute, second, 0, loc), true
}

// parseStandard uses the dateparse library to parse standard date/time formats.
func parseStandard(input string, loc *time.Location) (time.Time, error) {
	// First, try parsing with the dateparse library in the specified location
	t, err := dateparse.ParseIn(input, loc)
	if err == nil {
		return t, nil
	}

	// Try some additional common formats that dateparse might not handle well
	formats := []string{
		"2006-01-02T15:04:05",       // ISO 8601 without timezone
		"2006-01-02T15:04",          // ISO 8601 without seconds
		"2006-01-02 15:04:05",       // Natural with seconds
		"2006-01-02 15:04",          // Natural without seconds
		"2006/01/02 15:04:05",       // Slash format with seconds
		"2006/01/02 15:04",          // Slash format without seconds
		"01/02/2006 15:04:05",       // US format with seconds
		"01/02/2006 15:04",          // US format without seconds
		"02/01/2006 15:04:05",       // European format with seconds
		"02/01/2006 15:04",          // European format without seconds
		"Jan 2, 2006 15:04:05",      // Month name format
		"Jan 2, 2006 15:04",         // Month name format without seconds
		"January 2, 2006 15:04:05",  // Full month name format
		"January 2, 2006 15:04",     // Full month name format without seconds
		"2006-01-02",                // Date only (midnight)
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, input, loc); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("%w: could not parse '%s'. Try formats like '2024-01-15 14:00', '14:00', 'tomorrow 14:00', or 'in 2 hours'", ErrInvalidDateFormat, input)
}

// FormatTime formats a time.Time value for display.
func FormatTime(t time.Time) string {
	return t.Format("Mon, Jan 2, 2006 at 3:04 PM MST")
}

// FormatTimeShort formats a time.Time value in a shorter format.
func FormatTimeShort(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

// ParseDuration parses a duration string into a time.Duration.
// Supports formats like "30m", "1h", "1h30m", "90" (minutes as default).
func ParseDuration(input string) (time.Duration, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return 0, fmt.Errorf("empty duration")
	}

	// If it's just a number, treat it as minutes
	if minutes, err := strconv.Atoi(input); err == nil {
		return time.Duration(minutes) * time.Minute, nil
	}

	// Try standard Go duration parsing
	d, err := time.ParseDuration(input)
	if err != nil {
		return 0, fmt.Errorf("invalid duration '%s': use formats like '30m', '1h', '1h30m', or just '30' for minutes", input)
	}

	return d, nil
}
