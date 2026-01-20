package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.CalendarID != "primary" {
		t.Errorf("Expected CalendarID to be 'primary', got '%s'", cfg.CalendarID)
	}

	if cfg.DefaultDuration != 30 {
		t.Errorf("Expected DefaultDuration to be 30, got %d", cfg.DefaultDuration)
	}

	if cfg.CredentialsPath != "" {
		t.Errorf("Expected CredentialsPath to be empty, got '%s'", cfg.CredentialsPath)
	}

	if cfg.TokenPath != "" {
		t.Errorf("Expected TokenPath to be empty, got '%s'", cfg.TokenPath)
	}
}

func TestLoadWithDefaults(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("GOOGLE_CALENDAR_CREDENTIALS")
	os.Unsetenv("GOOGLE_CALENDAR_TOKEN")
	os.Unsetenv("GOOGLE_CALENDAR_ID")

	cfg, err := Load("", nil)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.CalendarID != "primary" {
		t.Errorf("Expected CalendarID to be 'primary', got '%s'", cfg.CalendarID)
	}

	if cfg.DefaultDuration != 30 {
		t.Errorf("Expected DefaultDuration to be 30, got %d", cfg.DefaultDuration)
	}
}

func TestLoadFromEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("GOOGLE_CALENDAR_CREDENTIALS", "/path/to/credentials.json")
	os.Setenv("GOOGLE_CALENDAR_TOKEN", "/path/to/token.json")
	os.Setenv("GOOGLE_CALENDAR_ID", "test-calendar-id")
	defer func() {
		os.Unsetenv("GOOGLE_CALENDAR_CREDENTIALS")
		os.Unsetenv("GOOGLE_CALENDAR_TOKEN")
		os.Unsetenv("GOOGLE_CALENDAR_ID")
	}()

	cfg, err := Load("", nil)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.CredentialsPath != "/path/to/credentials.json" {
		t.Errorf("Expected CredentialsPath to be '/path/to/credentials.json', got '%s'", cfg.CredentialsPath)
	}

	if cfg.TokenPath != "/path/to/token.json" {
		t.Errorf("Expected TokenPath to be '/path/to/token.json', got '%s'", cfg.TokenPath)
	}

	if cfg.CalendarID != "test-calendar-id" {
		t.Errorf("Expected CalendarID to be 'test-calendar-id', got '%s'", cfg.CalendarID)
	}
}

func TestLoadFromConfigFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
credentials_path: /config/credentials.json
token_path: /config/token.json
calendar_id: config-calendar-id
default_duration: 60
timezone: America/New_York
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Clear environment variables to ensure config file is used
	os.Unsetenv("GOOGLE_CALENDAR_CREDENTIALS")
	os.Unsetenv("GOOGLE_CALENDAR_TOKEN")
	os.Unsetenv("GOOGLE_CALENDAR_ID")

	cfg, err := Load(configPath, nil)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.CredentialsPath != "/config/credentials.json" {
		t.Errorf("Expected CredentialsPath to be '/config/credentials.json', got '%s'", cfg.CredentialsPath)
	}

	if cfg.TokenPath != "/config/token.json" {
		t.Errorf("Expected TokenPath to be '/config/token.json', got '%s'", cfg.TokenPath)
	}

	if cfg.CalendarID != "config-calendar-id" {
		t.Errorf("Expected CalendarID to be 'config-calendar-id', got '%s'", cfg.CalendarID)
	}

	if cfg.DefaultDuration != 60 {
		t.Errorf("Expected DefaultDuration to be 60, got %d", cfg.DefaultDuration)
	}

	if cfg.Timezone != "America/New_York" {
		t.Errorf("Expected Timezone to be 'America/New_York', got '%s'", cfg.Timezone)
	}
}

func TestConfigPriority_EnvOverridesConfigFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
credentials_path: /config/credentials.json
token_path: /config/token.json
calendar_id: config-calendar-id
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set environment variable to override config file
	os.Setenv("GOOGLE_CALENDAR_ID", "env-calendar-id")
	defer os.Unsetenv("GOOGLE_CALENDAR_ID")

	cfg, err := Load(configPath, nil)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Environment variable should override config file
	if cfg.CalendarID != "env-calendar-id" {
		t.Errorf("Expected CalendarID to be 'env-calendar-id' (from env), got '%s'", cfg.CalendarID)
	}

	// Config file values should still be used for non-overridden fields
	if cfg.CredentialsPath != "/config/credentials.json" {
		t.Errorf("Expected CredentialsPath to be '/config/credentials.json', got '%s'", cfg.CredentialsPath)
	}
}

func TestConfigPriority_FlagsOverrideAll(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
credentials_path: /config/credentials.json
token_path: /config/token.json
calendar_id: config-calendar-id
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set environment variable
	os.Setenv("GOOGLE_CALENDAR_ID", "env-calendar-id")
	defer os.Unsetenv("GOOGLE_CALENDAR_ID")

	// Pass flag overrides
	flagOverrides := map[string]interface{}{
		"calendar_id": "flag-calendar-id",
	}

	cfg, err := Load(configPath, flagOverrides)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Flag should override both env and config file
	if cfg.CalendarID != "flag-calendar-id" {
		t.Errorf("Expected CalendarID to be 'flag-calendar-id' (from flag), got '%s'", cfg.CalendarID)
	}
}

func TestValidate_MissingCredentialsPath(t *testing.T) {
	cfg := &Config{
		TokenPath:  "/path/to/token.json",
		CalendarID: "primary",
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected validation error for missing credentials path")
	}
	if err != ErrMissingCredentialsPath {
		t.Errorf("Expected ErrMissingCredentialsPath, got: %v", err)
	}
}

func TestValidate_MissingTokenPath(t *testing.T) {
	cfg := &Config{
		CredentialsPath: "/path/to/credentials.json",
		CalendarID:      "primary",
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected validation error for missing token path")
	}
	if err != ErrMissingTokenPath {
		t.Errorf("Expected ErrMissingTokenPath, got: %v", err)
	}
}

func TestValidate_Success(t *testing.T) {
	cfg := &Config{
		CredentialsPath: "/path/to/credentials.json",
		TokenPath:       "/path/to/token.json",
		CalendarID:      "primary",
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Expected no validation error, got: %v", err)
	}
}

func TestValidateCredentialsExist_FileNotFound(t *testing.T) {
	cfg := &Config{
		CredentialsPath: "/nonexistent/path/credentials.json",
	}

	err := cfg.ValidateCredentialsExist()
	if err == nil {
		t.Error("Expected error for nonexistent credentials file")
	}
}

func TestValidateCredentialsExist_FileExists(t *testing.T) {
	// Create a temporary credentials file
	tmpDir := t.TempDir()
	credentialsPath := filepath.Join(tmpDir, "credentials.json")

	if err := os.WriteFile(credentialsPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create credentials file: %v", err)
	}

	cfg := &Config{
		CredentialsPath: credentialsPath,
	}

	err := cfg.ValidateCredentialsExist()
	if err != nil {
		t.Errorf("Expected no error for existing credentials file, got: %v", err)
	}
}

func TestGetConfigDir(t *testing.T) {
	configDir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir failed: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "calgo")

	if configDir != expected {
		t.Errorf("Expected config dir to be '%s', got '%s'", expected, configDir)
	}
}

func TestEnsureConfigDir(t *testing.T) {
	// This test actually creates the directory, which is fine for testing
	configDir, err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir failed: %v", err)
	}

	// Check that the directory exists
	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("Config directory does not exist: %v", err)
	}

	if !info.IsDir() {
		t.Error("Config path is not a directory")
	}
}
