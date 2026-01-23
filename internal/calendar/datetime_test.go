package calendar

import (
	"os"
	"testing"
	"time"
)

func TestParseTime_ISO8601(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantYear int
		wantMon  time.Month
		wantDay  int
		wantHour int
		wantMin  int
		wantSec  int
	}{
		{
			name:     "ISO 8601 with T separator",
			input:    "2024-01-15T14:00:00",
			wantYear: 2024, wantMon: time.January, wantDay: 15,
			wantHour: 14, wantMin: 0, wantSec: 0,
		},
		{
			name:     "ISO 8601 without seconds",
			input:    "2024-01-15T14:30",
			wantYear: 2024, wantMon: time.January, wantDay: 15,
			wantHour: 14, wantMin: 30, wantSec: 0,
		},
		{
			name:     "ISO 8601 with Z timezone",
			input:    "2024-06-20T09:15:00Z",
			wantYear: 2024, wantMon: time.June, wantDay: 20,
			wantHour: 9, wantMin: 15, wantSec: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTime(tt.input, "UTC")
			if err != nil {
				t.Fatalf("ParseTime() error = %v", err)
			}
			got = got.UTC() // Normalize for comparison
			if got.Year() != tt.wantYear || got.Month() != tt.wantMon || got.Day() != tt.wantDay {
				t.Errorf("ParseTime() date = %d-%02d-%02d, want %d-%02d-%02d",
					got.Year(), got.Month(), got.Day(),
					tt.wantYear, tt.wantMon, tt.wantDay)
			}
			if got.Hour() != tt.wantHour || got.Minute() != tt.wantMin || got.Second() != tt.wantSec {
				t.Errorf("ParseTime() time = %02d:%02d:%02d, want %02d:%02d:%02d",
					got.Hour(), got.Minute(), got.Second(),
					tt.wantHour, tt.wantMin, tt.wantSec)
			}
		})
	}
}

func TestParseTime_NaturalFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantYear int
		wantMon  time.Month
		wantDay  int
		wantHour int
		wantMin  int
	}{
		{
			name:     "natural format with space",
			input:    "2024-01-15 14:00",
			wantYear: 2024, wantMon: time.January, wantDay: 15,
			wantHour: 14, wantMin: 0,
		},
		{
			name:     "natural format with seconds",
			input:    "2024-01-15 14:30:45",
			wantYear: 2024, wantMon: time.January, wantDay: 15,
			wantHour: 14, wantMin: 30,
		},
		{
			name:     "slash format",
			input:    "2024/03/20 10:00",
			wantYear: 2024, wantMon: time.March, wantDay: 20,
			wantHour: 10, wantMin: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTime(tt.input, "UTC")
			if err != nil {
				t.Fatalf("ParseTime() error = %v", err)
			}
			if got.Year() != tt.wantYear || got.Month() != tt.wantMon || got.Day() != tt.wantDay {
				t.Errorf("ParseTime() date = %d-%02d-%02d, want %d-%02d-%02d",
					got.Year(), got.Month(), got.Day(),
					tt.wantYear, tt.wantMon, tt.wantDay)
			}
			if got.Hour() != tt.wantHour || got.Minute() != tt.wantMin {
				t.Errorf("ParseTime() time = %02d:%02d, want %02d:%02d",
					got.Hour(), got.Minute(),
					tt.wantHour, tt.wantMin)
			}
		})
	}
}

func TestParseTime_TimeOnly(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		input    string
		wantHour int
		wantMin  int
		wantSec  int
	}{
		{
			name:     "time only HH:MM",
			input:    "14:00",
			wantHour: 14, wantMin: 0, wantSec: 0,
		},
		{
			name:     "time only HH:MM:SS",
			input:    "09:30:15",
			wantHour: 9, wantMin: 30, wantSec: 15,
		},
		{
			name:     "time only single digit hour",
			input:    "9:00",
			wantHour: 9, wantMin: 0, wantSec: 0,
		},
		{
			name:     "midnight",
			input:    "00:00",
			wantHour: 0, wantMin: 0, wantSec: 0,
		},
		{
			name:     "end of day",
			input:    "23:59",
			wantHour: 23, wantMin: 59, wantSec: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTime(tt.input, "")
			if err != nil {
				t.Fatalf("ParseTime() error = %v", err)
			}
			// Should be today's date
			if got.Year() != now.Year() || got.Month() != now.Month() || got.Day() != now.Day() {
				t.Errorf("ParseTime() date = %d-%02d-%02d, want today %d-%02d-%02d",
					got.Year(), got.Month(), got.Day(),
					now.Year(), now.Month(), now.Day())
			}
			if got.Hour() != tt.wantHour || got.Minute() != tt.wantMin || got.Second() != tt.wantSec {
				t.Errorf("ParseTime() time = %02d:%02d:%02d, want %02d:%02d:%02d",
					got.Hour(), got.Minute(), got.Second(),
					tt.wantHour, tt.wantMin, tt.wantSec)
			}
		})
	}
}

func TestParseTime_RelativeFormats(t *testing.T) {
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)

	tests := []struct {
		name        string
		input       string
		wantDay     int
		wantHour    int
		wantMin     int
		checkOffset bool // if true, check relative to now
		hourOffset  int
		minOffset   int
	}{
		{
			name:     "today 14:00",
			input:    "today 14:00",
			wantDay:  now.Day(),
			wantHour: 14, wantMin: 0,
		},
		{
			name:     "today at 14:00",
			input:    "today at 14:00",
			wantDay:  now.Day(),
			wantHour: 14, wantMin: 0,
		},
		{
			name:     "tomorrow 10:30",
			input:    "tomorrow 10:30",
			wantDay:  tomorrow.Day(),
			wantHour: 10, wantMin: 30,
		},
		{
			name:     "tomorrow at 9:00",
			input:    "tomorrow at 9:00",
			wantDay:  tomorrow.Day(),
			wantHour: 9, wantMin: 0,
		},
		{
			name:        "in 2 hours",
			input:       "in 2 hours",
			checkOffset: true,
			hourOffset:  2,
		},
		{
			name:        "in 1 hour",
			input:       "in 1 hour",
			checkOffset: true,
			hourOffset:  1,
		},
		{
			name:        "in 30 minutes",
			input:       "in 30 minutes",
			checkOffset: true,
			minOffset:   30,
		},
		{
			name:        "in 45 mins",
			input:       "in 45 mins",
			checkOffset: true,
			minOffset:   45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTime(tt.input, "")
			if err != nil {
				t.Fatalf("ParseTime() error = %v", err)
			}

			if tt.checkOffset {
				expected := now.Add(time.Duration(tt.hourOffset)*time.Hour + time.Duration(tt.minOffset)*time.Minute)
				// Allow 2 second tolerance for test execution time
				diff := got.Sub(expected)
				if diff < -2*time.Second || diff > 2*time.Second {
					t.Errorf("ParseTime() = %v, want approximately %v (diff: %v)", got, expected, diff)
				}
			} else {
				if got.Day() != tt.wantDay {
					t.Errorf("ParseTime() day = %d, want %d", got.Day(), tt.wantDay)
				}
				if got.Hour() != tt.wantHour || got.Minute() != tt.wantMin {
					t.Errorf("ParseTime() time = %02d:%02d, want %02d:%02d",
						got.Hour(), got.Minute(), tt.wantHour, tt.wantMin)
				}
			}
		})
	}
}

func TestParseTime_Timezone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		timezone string
		wantTZ   string
	}{
		{
			name:     "explicit UTC timezone",
			input:    "2024-01-15 14:00",
			timezone: "UTC",
			wantTZ:   "UTC",
		},
		{
			name:     "explicit America/New_York timezone",
			input:    "2024-01-15 14:00",
			timezone: "America/New_York",
			wantTZ:   "EST",
		},
		{
			name:     "explicit Europe/London timezone",
			input:    "2024-01-15 14:00",
			timezone: "Europe/London",
			wantTZ:   "GMT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTime(tt.input, tt.timezone)
			if err != nil {
				t.Fatalf("ParseTime() error = %v", err)
			}
			gotTZ, _ := got.Zone()
			if gotTZ != tt.wantTZ {
				t.Errorf("ParseTime() timezone = %s, want %s", gotTZ, tt.wantTZ)
			}
		})
	}
}

func TestParseTime_TZEnvironment(t *testing.T) {
	// Save original TZ
	origTZ := os.Getenv("TZ")
	defer os.Setenv("TZ", origTZ)

	// Set TZ to a known timezone
	os.Setenv("TZ", "America/Los_Angeles")

	got, err := ParseTime("14:00", "")
	if err != nil {
		t.Fatalf("ParseTime() error = %v", err)
	}

	gotTZ, _ := got.Zone()
	// Could be PST or PDT depending on date
	if gotTZ != "PST" && gotTZ != "PDT" {
		t.Errorf("ParseTime() with TZ env timezone = %s, want PST or PDT", gotTZ)
	}
}

func TestParseTime_InvalidFormats(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"random text", "not a date"},
		{"invalid time", "25:00"},
		{"invalid minutes", "14:60"},
		{"invalid month", "2024-13-15 14:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTime(tt.input, "")
			if err == nil {
				t.Errorf("ParseTime() expected error for input %q", tt.input)
			}
		})
	}
}

func TestParseTime_InvalidTimezone(t *testing.T) {
	_, err := ParseTime("14:00", "Invalid/Timezone")
	if err == nil {
		t.Error("ParseTime() expected error for invalid timezone")
	}
}

func TestGetLocation(t *testing.T) {
	// Save original TZ
	origTZ := os.Getenv("TZ")
	defer os.Setenv("TZ", origTZ)

	tests := []struct {
		name     string
		timezone string
		envTZ    string
		wantName string
	}{
		{
			name:     "explicit timezone takes priority",
			timezone: "UTC",
			envTZ:    "America/New_York",
			wantName: "UTC",
		},
		{
			name:     "env TZ used when no explicit timezone",
			timezone: "",
			envTZ:    "America/Chicago",
			wantName: "America/Chicago",
		},
		{
			name:     "local used when nothing set",
			timezone: "",
			envTZ:    "",
			wantName: "Local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TZ", tt.envTZ)
			loc, err := getLocation(tt.timezone)
			if err != nil {
				t.Fatalf("getLocation() error = %v", err)
			}
			if tt.wantName == "Local" {
				if loc != time.Local {
					t.Errorf("getLocation() = %v, want Local", loc)
				}
			} else if loc.String() != tt.wantName {
				t.Errorf("getLocation() = %v, want %v", loc.String(), tt.wantName)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     time.Duration
		wantErr  bool
	}{
		{
			name:  "minutes only as number",
			input: "30",
			want:  30 * time.Minute,
		},
		{
			name:  "minutes with m suffix",
			input: "30m",
			want:  30 * time.Minute,
		},
		{
			name:  "hours with h suffix",
			input: "2h",
			want:  2 * time.Hour,
		},
		{
			name:  "hours and minutes",
			input: "1h30m",
			want:  90 * time.Minute,
		},
		{
			name:  "seconds",
			input: "90s",
			want:  90 * time.Second,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	loc, _ := time.LoadLocation("UTC")
	testTime := time.Date(2024, time.January, 15, 14, 30, 0, 0, loc)

	got := FormatTime(testTime)
	want := "Mon, Jan 15, 2024 at 2:30 PM UTC"

	if got != want {
		t.Errorf("FormatTime() = %q, want %q", got, want)
	}
}

func TestFormatTimeShort(t *testing.T) {
	loc, _ := time.LoadLocation("UTC")
	testTime := time.Date(2024, time.January, 15, 14, 30, 0, 0, loc)

	got := FormatTimeShort(testTime)
	want := "2024-01-15 14:30"

	if got != want {
		t.Errorf("FormatTimeShort() = %q, want %q", got, want)
	}
}

func TestParseTimeOnly_InvalidTimes(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"hour too high", "24:00"},
		{"minute too high", "14:60"},
		{"second too high", "14:00:60"},
		{"negative hour", "-1:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := time.Local
			_, ok := parseTimeOnly(tt.input, loc)
			if ok {
				t.Errorf("parseTimeOnly() should return false for invalid input %q", tt.input)
			}
		})
	}
}
