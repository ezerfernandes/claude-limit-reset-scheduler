# Tasks Specification

## Overview

This document contains all development tasks for the Go CLI calendar tool, organized by phase and linked to user stories from `features.md`.

---

## Phase 0: Project Setup and Configuration

### T-000: Initialize Go Module

**Description:** Initialize the Go module and set up the project structure.

**Linked User Stories:** None (Infrastructure)

**Subtasks:**
- [x] Run `go mod init` with appropriate module name
- [x] Create standard Go project directory structure:
  ```
  /cmd/calgo/         # Main application entry point
  /internal/          # Private application code
    /auth/            # OAuth2 authentication
    /calendar/        # Google Calendar API wrapper
    /config/          # Configuration management
    /cli/             # CLI command definitions
  /pkg/               # Public library code (if any)
  ```
- [x] Create initial `main.go` file

**Acceptance Criteria:**
- [x] `go build` runs successfully
- [x] Project follows standard Go project layout

**Status:** ✅ COMPLETED

---

### T-001: Add Core Dependencies

**Description:** Add required Go dependencies for the project.

**Linked User Stories:** None (Infrastructure)

**Subtasks:**
- [x] Add Google Calendar API client: `google.golang.org/api/calendar/v3`
- [x] Add OAuth2 library: `golang.org/x/oauth2/google`
- [x] Add CLI framework (Cobra): `github.com/spf13/cobra`
- [x] Add configuration library (Viper): `github.com/spf13/viper`
- [x] Add date parsing library: `github.com/araddon/dateparse`
- [x] Run `go mod tidy`

**Acceptance Criteria:**
- [x] All dependencies are added to `go.mod`
- [x] `go mod tidy` completes without errors

**Status:** ✅ COMPLETED

---

### T-002: Set Up Google Cloud Project

**Description:** Document the steps to create and configure a Google Cloud project.

**Linked User Stories:** US-001, US-002

**Subtasks:**
- [x] Create documentation for Google Cloud Console setup
- [x] Document OAuth2 credential creation process
- [x] Document required API enablement (Google Calendar API)
- [x] Create `.env.example` file with required environment variables
- [x] Add credentials and token files to `.gitignore`

**Acceptance Criteria:**
- [x] Clear setup instructions exist in README
- [x] Example environment file is provided
- [x] Sensitive files are gitignored

**Status:** ✅ COMPLETED

---

### T-003: Set Up Development Environment

**Description:** Configure development tooling and CI/CD basics.

**Linked User Stories:** None (Infrastructure)

**Subtasks:**
- [x] Create `Makefile` with common commands (build, test, lint, run)
- [x] Add `.editorconfig` for consistent formatting
- [x] Set up `golangci-lint` configuration
- [x] Create `Dockerfile` for containerized builds (optional)
- [x] Add GitHub Actions workflow for CI (optional)

**Acceptance Criteria:**
- [x] `make build` creates the binary
- [x] `make test` runs all tests
- [x] `make lint` runs linting

**Status:** ✅ COMPLETED

---

## Phase 1: Core Functionality (MVP)

### T-100: Implement Configuration Loading

**Description:** Implement environment variable and configuration file loading.

**Linked User Stories:** US-001, US-010

**Subtasks:**
- [x] Create `internal/config/config.go` with Config struct
- [x] Implement environment variable loading:
  - `GOOGLE_CALENDAR_CREDENTIALS` - path to credentials JSON
  - `GOOGLE_CALENDAR_TOKEN` - path to token file
  - `GOOGLE_CALENDAR_ID` - target calendar ID (default: "primary")
- [x] Implement configuration file loading from `~/.config/calgo/config.yaml`
- [x] Implement configuration priority: CLI flags > Env vars > Config file > Defaults
- [x] Add validation for required configuration values
- [x] Write unit tests for configuration loading

**Acceptance Criteria:**
- [x] Configuration loads from environment variables
- [x] Missing required config shows clear error
- [x] Priority order is respected

**Status:** ✅ COMPLETED

---

### T-101: Implement OAuth2 Authentication Flow

**Description:** Implement Google OAuth2 authentication and token management.

**Linked User Stories:** US-002

**Subtasks:**
- [x] Create `internal/auth/oauth.go` with authentication logic
- [x] Implement OAuth2 config loading from credentials JSON
- [x] Implement first-time authentication flow:
  - Generate authorization URL
  - Open browser for user authorization
  - Handle OAuth2 callback
  - Save token to file
- [x] Implement token loading from saved file
- [x] Implement automatic token refresh
- [x] Handle token refresh failures with re-authentication prompt
- [x] Write unit tests with mocked OAuth2 flow

**Acceptance Criteria:**
- [x] First-time auth opens browser and saves token
- [x] Subsequent runs use saved token
- [x] Expired tokens are automatically refreshed
- [x] Failed refresh prompts re-authentication

**Status:** ✅ COMPLETED

---

### T-102: Implement Google Calendar Client

**Description:** Create a wrapper around the Google Calendar API.

**Linked User Stories:** US-003

**Subtasks:**
- [x] Create `internal/calendar/client.go` with Calendar client struct
- [x] Implement `NewClient` function that takes OAuth2 token
- [x] Implement `CreateEvent` method with event parameters:
  - Title (summary)
  - Start time
  - End time (calculated from duration)
  - Description (optional)
  - Location (optional)
- [x] Implement response parsing to extract event details and link
- [x] Handle API errors with user-friendly messages
- [x] Write integration tests (with test calendar)

**Acceptance Criteria:**
- Events are created successfully in Google Calendar
- Event link is returned after creation
- API errors are handled gracefully

**Status:** ✅ COMPLETED

---

### T-103: Implement Date/Time Parsing

**Description:** Implement flexible date/time parsing for event times.

**Linked User Stories:** US-004

**Subtasks:**
- [x] Create `internal/calendar/datetime.go` with parsing functions
- [x] Support ISO 8601 format: "2024-01-15T14:00:00"
- [x] Support natural format: "2024-01-15 14:00"
- [x] Support time-only format: "14:00" (assumes today)
- [x] Support relative formats: "tomorrow 14:00", "in 2 hours"
- [x] Implement timezone handling from `TZ` environment variable
- [x] Fall back to system timezone if not configured
- [x] Write unit tests for all date formats

**Acceptance Criteria:**
- All specified date formats are parsed correctly
- Timezone is respected in parsing and display
- Invalid formats return helpful error messages

**Status:** ✅ COMPLETED

---

### T-104: Implement CLI Framework

**Description:** Set up the CLI command structure using Cobra.

**Linked User Stories:** US-007

**Subtasks:**
- [ ] Create `cmd/calgo/main.go` with root command
- [ ] Implement `--help` and `-h` flags
- [ ] Implement `--version` flag
- [ ] Add global flags:
  - `--config` for custom config file path
  - `--quiet` or `-q` for minimal output
  - `--json` for JSON output
- [ ] Create help templates with usage examples
- [ ] Wire up configuration loading in root command

**Acceptance Criteria:**
- `calgo --help` shows usage information
- `calgo --version` shows version
- Global flags are available to all subcommands

---

### T-105: Implement Create Command

**Description:** Implement the `create` subcommand for creating calendar events.

**Linked User Stories:** US-003, US-006

**Subtasks:**
- [ ] Create `internal/cli/create.go` with create command
- [ ] Add command flags:
  - `--title` or `-t` (required)
  - `--start` or `-s` (required)
  - `--duration` or `-d` (default: 30)
  - `--description` or `-D` (optional)
  - `--location` or `-l` (optional)
- [ ] Implement flag validation
- [ ] Wire up to calendar client for event creation
- [ ] Implement output formatting (standard, JSON, quiet)
- [ ] Write unit tests for command parsing
- [ ] Write integration tests for event creation

**Acceptance Criteria:**
- `calgo create --title "Test" --start "2024-01-15 14:00"` creates an event
- Missing required flags show clear error
- Created event displays confirmation with link

---

### T-106: Implement Error Handling

**Description:** Implement consistent error handling and user-friendly error messages.

**Linked User Stories:** US-001, US-002, US-003, US-006

**Subtasks:**
- [ ] Create `internal/errors/errors.go` with custom error types
- [ ] Implement error wrapper for configuration errors
- [ ] Implement error wrapper for authentication errors
- [ ] Implement error wrapper for API errors
- [ ] Add helpful suggestions in error messages
- [ ] Ensure errors are displayed consistently

**Acceptance Criteria:**
- All errors have clear, actionable messages
- Error messages suggest how to resolve the issue
- Exit codes are appropriate (0 for success, 1 for error)

---

## Phase 2: Enhanced Features

### T-200: Implement Quick Event Parsing

**Description:** Implement natural language parsing for quick event creation.

**Linked User Stories:** US-005

**Subtasks:**
- [ ] Create `internal/calendar/quickparse.go` with parsing logic
- [ ] Implement pattern matching for common formats:
  - "Meeting tomorrow at 2pm"
  - "Lunch with John on Friday at noon for 1 hour"
  - "Standup in 30 minutes"
- [ ] Extract title, date/time, and duration from input
- [ ] Implement `quick` subcommand
- [ ] Display parsed values before creating event
- [ ] Write unit tests for various input patterns

**Acceptance Criteria:**
- Common natural language formats are parsed correctly
- Parsed values are displayed for verification
- Unparseable input shows helpful error with examples

---

### T-201: Implement Recurring Events

**Description:** Add support for creating recurring events.

**Linked User Stories:** US-008

**Subtasks:**
- [ ] Extend `CreateEvent` to support recurrence rules
- [ ] Add `--repeat` or `-r` flag with values: daily, weekly, monthly, weekdays
- [ ] Add `--until` flag for recurrence end date
- [ ] Add `--count` flag for number of occurrences
- [ ] Validate that only one of `--until` or `--count` is used
- [ ] Generate RRULE string for Google Calendar API
- [ ] Write unit tests for recurrence rule generation

**Acceptance Criteria:**
- Recurring events are created with correct recurrence pattern
- `--until` and `--count` work correctly
- Conflicting flags show appropriate error

---

### T-202: Implement Event Reminders

**Description:** Add support for event reminders.

**Linked User Stories:** US-009

**Subtasks:**
- [ ] Extend `CreateEvent` to support reminders
- [ ] Add `--reminder` flag (can be used multiple times)
- [ ] Add `--no-reminder` flag to disable default reminders
- [ ] Parse reminder value as minutes
- [ ] Configure reminder in Google Calendar API request
- [ ] Write unit tests for reminder configuration

**Acceptance Criteria:**
- Events are created with specified reminders
- Multiple reminders can be set
- `--no-reminder` creates event without reminders

---

### T-203: Implement Configuration File Support

**Description:** Add full configuration file support with defaults.

**Linked User Stories:** US-010

**Subtasks:**
- [ ] Implement config file creation on first run
- [ ] Add `calgo config init` command to create default config
- [ ] Add `calgo config show` command to display current config
- [ ] Implement config file validation
- [ ] Support all configurable options in config file
- [ ] Document config file format in README

**Acceptance Criteria:**
- Config file is created at `~/.config/calgo/config.yaml`
- Config values are respected as defaults
- Config commands work correctly

---

## Phase 3: Polish and Release

### T-300: Improve Output Formatting

**Description:** Enhance CLI output with colors and formatting.

**Linked User Stories:** US-006

**Subtasks:**
- [ ] Add color support for terminal output
- [ ] Implement success/error coloring
- [ ] Add emoji indicators (optional, disabled by default)
- [ ] Implement table formatting for list outputs
- [ ] Respect `NO_COLOR` environment variable

**Acceptance Criteria:**
- Output is visually clear and easy to read
- Colors are disabled when not supported or `NO_COLOR` is set

---

### T-301: Add Comprehensive Help

**Description:** Enhance help documentation with examples and guides.

**Linked User Stories:** US-007

**Subtasks:**
- [ ] Add detailed examples to all command help
- [ ] Create `calgo examples` command showing common use cases
- [ ] Add date format examples in `--start` flag help
- [ ] Create man page (optional)
- [ ] Update README with full usage documentation

**Acceptance Criteria:**
- Help text includes practical examples
- Users can learn the tool from help output alone

---

### T-302: Implement Dry Run Mode

**Description:** Add a dry run mode to preview actions without executing.

**Linked User Stories:** US-003

**Subtasks:**
- [ ] Add `--dry-run` flag to create command
- [ ] Display what would be created without making API call
- [ ] Show parsed date/time and calculated end time
- [ ] Validate all inputs in dry run mode

**Acceptance Criteria:**
- `--dry-run` shows event details without creating
- All validation runs in dry run mode

---

### T-303: Build and Release Automation

**Description:** Set up automated builds and releases.

**Linked User Stories:** None (Infrastructure)

**Subtasks:**
- [ ] Configure GoReleaser for multi-platform builds
- [ ] Set up GitHub Actions for automated releases
- [ ] Create Homebrew formula (macOS)
- [ ] Create installation script
- [ ] Add checksums for release binaries

**Acceptance Criteria:**
- Releases are automated via GitHub tags
- Binaries available for Linux, macOS, Windows
- Installation instructions for all platforms

---

## Task Dependencies

```
T-000 ─┬─> T-001 ─┬─> T-100 ─┬─> T-101 ─> T-102 ─┬─> T-105
       │         │          │                    │
       │         │          └─> T-103 ───────────┘
       │         │
       │         └─> T-104 ──────────────────────────> T-105
       │
       └─> T-002
       │
       └─> T-003

T-105 ─┬─> T-106
       │
       ├─> T-200
       │
       ├─> T-201
       │
       ├─> T-202
       │
       └─> T-203 ─> T-300 ─> T-301 ─> T-302 ─> T-303
```

---

## Effort Estimation

| Task  | Effort   | Priority |
|-------|----------|----------|
| T-000 | Small    | P0       |
| T-001 | Small    | P0       |
| T-002 | Small    | P0       |
| T-003 | Medium   | P1       |
| T-100 | Medium   | P0       |
| T-101 | Large    | P0       |
| T-102 | Medium   | P0       |
| T-103 | Medium   | P0       |
| T-104 | Medium   | P0       |
| T-105 | Medium   | P0       |
| T-106 | Small    | P0       |
| T-200 | Large    | P2       |
| T-201 | Medium   | P1       |
| T-202 | Small    | P1       |
| T-203 | Medium   | P2       |
| T-300 | Small    | P2       |
| T-301 | Small    | P1       |
| T-302 | Small    | P2       |
| T-303 | Medium   | P1       |

**Legend:**
- Small: < 2 hours
- Medium: 2-4 hours
- Large: 4-8 hours
- P0: Required for MVP
- P1: Important, post-MVP
- P2: Nice to have
