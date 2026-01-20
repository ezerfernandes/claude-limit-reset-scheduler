# calgo - Google Calendar CLI

A Go CLI tool for creating Google Calendar events directly from the terminal.

## Features

- Create calendar events with a simple command
- Flexible date/time input formats (ISO 8601, natural language, relative)
- OAuth2 authentication with Google Calendar API
- Environment variable and configuration file support
- JSON output for scripting

## Installation

### From Source

```bash
git clone https://github.com/ezer/calgo.git
cd calgo
go build -o calgo ./cmd/calgo
```

## Google Cloud Setup

Before using calgo, you need to set up a Google Cloud project and enable the Google Calendar API.

### Step 1: Create a Google Cloud Project

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Click the project dropdown in the top navigation bar
3. Click **New Project**
4. Enter a project name (e.g., "calgo")
5. Click **Create**
6. Wait for the project to be created, then select it from the project dropdown

### Step 2: Enable the Google Calendar API

1. In the Google Cloud Console, go to **APIs & Services** > **Library**
2. Search for "Google Calendar API"
3. Click on **Google Calendar API** in the results
4. Click **Enable**
5. Wait for the API to be enabled

### Step 3: Configure OAuth Consent Screen

1. Go to **Menu** > **Google Auth platform** > **Branding**
2. If you see "Google Auth platform not configured yet", click **Get Started**
3. Under **App Information**:
   - Enter **App name**: calgo
   - Enter **User support email**: your email address
   - Click **Next**
4. Under **Audience**:
   - Select **External** (unless you have a Google Workspace account for Internal)
   - Click **Next**
5. Under **Contact Information**:
   - Enter an email address for developer notifications
   - Click **Next**
6. Under **Finish**:
   - Review the Google API Services User Data Policy
   - Check the agreement checkbox
   - Click **Continue**, then **Create**
7. Navigate to **Google Auth platform** > **Data Access**
8. Click **Add or Remove Scopes**
9. Find and select `https://www.googleapis.com/auth/calendar.events`
10. Click **Save**
11. Navigate to **Google Auth platform** > **Audience**
12. Under **Test users**, click **Add users**
13. Add your Google email address and click **Save**

> **Note:** Apps in "Testing" status are limited to 100 test users, and authorizations expire after 7 days. For personal use, this is sufficient.

### Step 4: Create OAuth2 Credentials

1. Go to **APIs & Services** > **Credentials**
2. Click **Create Credentials** > **OAuth client ID**
3. Select **Desktop app** as the application type
4. Enter a name (e.g., "calgo CLI")
5. Click **Create**
6. In the dialog that appears, click **Download JSON**
7. Save the file as `credentials.json` in your working directory
8. Click **OK** to close the dialog

> **Important**: Keep your `credentials.json` file secure and never commit it to version control. It's already listed in `.gitignore`.

## Configuration

### Environment Variables

calgo uses environment variables for configuration. Copy the example file and fill in your values:

```bash
cp .env.example .env
```

Required environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `GOOGLE_CALENDAR_CREDENTIALS` | Path to OAuth2 credentials JSON file | None (required) |
| `GOOGLE_CALENDAR_TOKEN` | Path where OAuth2 token will be stored | None (required) |
| `GOOGLE_CALENDAR_ID` | Target calendar ID | `primary` |

Example `.env` file:

```bash
GOOGLE_CALENDAR_CREDENTIALS=./credentials.json
GOOGLE_CALENDAR_TOKEN=./token.json
GOOGLE_CALENDAR_ID=primary
```

### Configuration File (Optional)

calgo also supports a configuration file at `~/.config/calgo/config.yaml`:

```yaml
calendar_id: primary
default_duration: 30
timezone: America/New_York
```

Configuration priority (highest to lowest):
1. Command-line flags
2. Environment variables
3. Configuration file
4. Built-in defaults

## First-Time Authentication

When you run calgo for the first time:

1. calgo will open your default browser
2. Sign in to your Google account
3. Grant calgo access to manage your calendar events
4. The browser will redirect to a local page confirming success
5. A token file will be saved for future use

Subsequent runs will use the saved token automatically. If the token expires, calgo will refresh it automatically.

## Usage

### Basic Usage

```bash
# Show help
calgo --help

# Show version
calgo --version

# Create a simple event
calgo create --title "Team Meeting" --start "2024-01-15 14:00"

# Create an event with duration
calgo create --title "Lunch" --start "tomorrow 12:00" --duration 60

# Create an event with all options
calgo create \
  --title "Project Review" \
  --start "2024-01-15T14:00:00" \
  --duration 90 \
  --description "Quarterly project review" \
  --location "Conference Room A"
```

### Date/Time Formats

calgo supports multiple date/time formats:

- ISO 8601: `2024-01-15T14:00:00`
- Natural: `2024-01-15 14:00`
- Time only (assumes today): `14:00`
- Relative: `tomorrow 14:00`, `in 2 hours`

### Output Formats

```bash
# Standard output (default)
calgo create --title "Meeting" --start "14:00"

# JSON output for scripting
calgo create --title "Meeting" --start "14:00" --json

# Quiet mode (only outputs event ID)
calgo create --title "Meeting" --start "14:00" --quiet
```

## Troubleshooting

### "Missing required environment variable"

Ensure you have set all required environment variables. See the Configuration section above.

### "Invalid credentials file"

Make sure your `credentials.json` file:
- Exists at the path specified by `GOOGLE_CALENDAR_CREDENTIALS`
- Is a valid JSON file downloaded from Google Cloud Console
- Contains OAuth2 Desktop application credentials

### "Token refresh failed"

Delete your token file and run calgo again to re-authenticate:

```bash
rm ./token.json
calgo create --title "Test" --start "14:00"
```

### "Access denied" or "Insufficient permissions"

1. Verify you added your email to the test users in OAuth consent screen
2. Ensure the Google Calendar API is enabled
3. Delete your token file and re-authenticate

## Development

```bash
# Build
go build -o calgo ./cmd/calgo

# Run tests
go test ./...

# Run linting
golangci-lint run
```

## License

MIT License - see LICENSE file for details.
