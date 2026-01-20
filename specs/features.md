# Features Specification

## Overview

A Go CLI tool for creating Google Calendar events directly from the terminal, using environment variables for configuration.

---

## User Stories

### US-001: Environment Variable Configuration

**As a** user
**I want to** configure the CLI tool using environment variables
**So that** I can securely store my Google Calendar API credentials without hardcoding them

#### Acceptance Criteria (EARS Notation)

1. **WHEN** the CLI starts, the system shall read the `GOOGLE_CALENDAR_CREDENTIALS` environment variable for the OAuth2 credentials JSON.
2. **WHEN** the CLI starts, the system shall read the `GOOGLE_CALENDAR_TOKEN` environment variable for the stored OAuth2 token path.
3. **WHEN** the CLI starts, the system shall read the `GOOGLE_CALENDAR_ID` environment variable to determine the target calendar (defaults to "primary").
4. **IF** required environment variables are missing, **THEN** the system shall display a clear error message indicating which variables are required.
5. **IF** the credentials file path is invalid, **THEN** the system shall display an error message with the expected file format.

---

### US-002: OAuth2 Authentication

**As a** user
**I want to** authenticate with Google Calendar API using OAuth2
**So that** I can securely access my calendar without exposing sensitive credentials

#### Acceptance Criteria (EARS Notation)

1. **WHEN** the user runs the CLI for the first time, the system shall initiate the OAuth2 authentication flow.
2. **WHEN** OAuth2 authentication is initiated, the system shall open a browser URL for the user to authorize the application.
3. **WHEN** the user completes authorization, the system shall store the OAuth2 token for future use.
4. **WHILE** a valid token exists, the system shall use the stored token without requiring re-authentication.
5. **IF** the stored token is expired, **THEN** the system shall automatically refresh the token using the refresh token.
6. **IF** token refresh fails, **THEN** the system shall prompt the user to re-authenticate.

---

### US-003: Create Calendar Event

**As a** user
**I want to** create a Google Calendar event with a simple command
**So that** I can quickly schedule events without leaving the terminal

#### Acceptance Criteria (EARS Notation)

1. **WHEN** the user runs `calgo create --title "Event" --start "2024-01-15 14:00" --duration 60`, the system shall create a calendar event with the specified parameters.
2. **WHEN** an event is created successfully, the system shall display a confirmation message with the event details and a link to the event.
3. The system shall support the following event parameters:
   - `--title` or `-t`: Event title (required)
   - `--start` or `-s`: Start date/time (required)
   - `--duration` or `-d`: Duration in minutes (optional, defaults to 30)
   - `--description` or `-D`: Event description (optional)
   - `--location` or `-l`: Event location (optional)
4. **IF** required parameters are missing, **THEN** the system shall display a usage message showing required parameters.
5. **IF** the date/time format is invalid, **THEN** the system shall display an error with supported formats.

---

### US-004: Flexible Date/Time Input

**As a** user
**I want to** specify event times in multiple formats
**So that** I can use the format most convenient for my workflow

#### Acceptance Criteria (EARS Notation)

1. **WHEN** the user provides a date/time, the system shall accept ISO 8601 format (e.g., "2024-01-15T14:00:00").
2. **WHEN** the user provides a date/time, the system shall accept natural format (e.g., "2024-01-15 14:00").
3. **WHEN** the user provides a relative time, the system shall accept formats like "tomorrow 14:00" or "in 2 hours".
4. **WHEN** the user provides only a time (e.g., "14:00"), the system shall assume today's date.
5. **WHERE** timezone configuration is provided via `TZ` environment variable, the system shall use that timezone for parsing and display.
6. **IF** no timezone is configured, **THEN** the system shall use the system's local timezone.

---

### US-005: Quick Event Creation

**As a** user
**I want to** create events using a quick natural language format
**So that** I can schedule events even faster

#### Acceptance Criteria (EARS Notation)

1. **WHEN** the user runs `calgo quick "Team meeting tomorrow at 2pm for 1 hour"`, the system shall parse and create the event.
2. **WHEN** parsing a quick event string, the system shall extract:
   - Event title (required)
   - Date/time (required)
   - Duration (optional, defaults to 30 minutes)
3. **IF** the quick event string cannot be parsed, **THEN** the system shall display an error with examples of valid formats.
4. **WHEN** a quick event is created, the system shall display the parsed values for user confirmation.

---

### US-006: Event Confirmation and Output

**As a** user
**I want to** receive clear feedback when events are created
**So that** I can verify the event was created correctly

#### Acceptance Criteria (EARS Notation)

1. **WHEN** an event is created successfully, the system shall display:
   - Event title
   - Start date/time (in local timezone)
   - End date/time (in local timezone)
   - Event link (Google Calendar URL)
2. **WHERE** the `--json` flag is provided, the system shall output the event details in JSON format.
3. **WHERE** the `--quiet` or `-q` flag is provided, the system shall only output the event ID.
4. **IF** event creation fails, **THEN** the system shall display a descriptive error message.

---

### US-007: Help and Usage Information

**As a** user
**I want to** access help documentation from the CLI
**So that** I can learn how to use the tool effectively

#### Acceptance Criteria (EARS Notation)

1. **WHEN** the user runs `calgo --help` or `calgo -h`, the system shall display general usage information.
2. **WHEN** the user runs `calgo create --help`, the system shall display detailed help for the create command.
3. **WHEN** the user runs `calgo version` or `calgo --version`, the system shall display the current version.
4. The system shall include example commands in the help output.

---

### US-008: Recurring Events

**As a** user
**I want to** create recurring calendar events
**So that** I can schedule repeated events with a single command

#### Acceptance Criteria (EARS Notation)

1. **WHERE** the `--repeat` or `-r` flag is provided, the system shall create a recurring event.
2. **WHEN** creating a recurring event, the system shall support the following recurrence patterns:
   - `daily`: Repeats every day
   - `weekly`: Repeats every week on the same day
   - `monthly`: Repeats every month on the same date
   - `weekdays`: Repeats Monday through Friday
3. **WHERE** the `--until` flag is provided, the system shall set an end date for the recurrence.
4. **WHERE** the `--count` flag is provided, the system shall limit the number of occurrences.
5. **IF** both `--until` and `--count` are provided, **THEN** the system shall display an error indicating only one can be used.

---

### US-009: Event Reminders

**As a** user
**I want to** set reminders for calendar events
**So that** I receive notifications before the event starts

#### Acceptance Criteria (EARS Notation)

1. **WHERE** the `--reminder` flag is provided, the system shall add a reminder to the event.
2. **WHEN** setting a reminder, the system shall accept values in minutes (e.g., `--reminder 15` for 15 minutes before).
3. **WHERE** multiple `--reminder` flags are provided, the system shall add multiple reminders.
4. **IF** no reminder is specified, **THEN** the system shall use the calendar's default reminder settings.
5. **WHERE** `--no-reminder` flag is provided, the system shall create the event without any reminders.

---

### US-010: Configuration File Support

**As a** user
**I want to** optionally use a configuration file
**So that** I can set default values and avoid repetitive command-line options

#### Acceptance Criteria (EARS Notation)

1. **WHEN** the CLI starts, the system shall check for a configuration file at `~/.config/calgo/config.yaml`.
2. **WHERE** a configuration file exists, the system shall read default values for:
   - Default calendar ID
   - Default event duration
   - Default reminder settings
   - Default timezone
3. **WHEN** both config file and environment variables are set, the system shall prioritize environment variables.
4. **WHEN** both config file/env vars and CLI flags are set, the system shall prioritize CLI flags.
5. **WHERE** the `--config` flag is provided, the system shall use the specified configuration file path.

---

## Priority Matrix

| User Story | Priority | Complexity | MVP |
|------------|----------|------------|-----|
| US-001     | High     | Low        | Yes |
| US-002     | High     | Medium     | Yes |
| US-003     | High     | Medium     | Yes |
| US-004     | Medium   | Medium     | Yes |
| US-005     | Low      | High       | No  |
| US-006     | High     | Low        | Yes |
| US-007     | High     | Low        | Yes |
| US-008     | Medium   | Medium     | No  |
| US-009     | Medium   | Low        | No  |
| US-010     | Low      | Medium     | No  |

---

## Glossary

- **EARS**: Easy Approach to Requirements Syntax - a notation for writing clear, unambiguous requirements
- **OAuth2**: Open Authorization 2.0 - the authorization framework used by Google APIs
- **MVP**: Minimum Viable Product - the core features needed for initial release
