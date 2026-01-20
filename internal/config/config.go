package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all configuration values for the application.
type Config struct {
	// CredentialsPath is the path to the OAuth2 credentials JSON file.
	CredentialsPath string `mapstructure:"credentials_path"`

	// TokenPath is the path where the OAuth2 token will be stored.
	TokenPath string `mapstructure:"token_path"`

	// CalendarID is the target calendar ID (defaults to "primary").
	CalendarID string `mapstructure:"calendar_id"`

	// DefaultDuration is the default event duration in minutes.
	DefaultDuration int `mapstructure:"default_duration"`

	// Timezone is the default timezone for events.
	Timezone string `mapstructure:"timezone"`
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() *Config {
	return &Config{
		CalendarID:      "primary",
		DefaultDuration: 30,
	}
}

// Errors for configuration validation.
var (
	ErrMissingCredentialsPath = errors.New("missing required configuration: credentials path (set GOOGLE_CALENDAR_CREDENTIALS or credentials_path in config)")
	ErrMissingTokenPath       = errors.New("missing required configuration: token path (set GOOGLE_CALENDAR_TOKEN or token_path in config)")
	ErrCredentialsNotFound    = errors.New("credentials file not found")
)

// Load loads configuration from all sources with the following priority:
// 1. CLI flags (passed via flagOverrides)
// 2. Environment variables
// 3. Configuration file (~/.config/calgo/config.yaml)
// 4. Default values
func Load(configPath string, flagOverrides map[string]interface{}) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("calendar_id", "primary")
	v.SetDefault("default_duration", 30)

	// Configure config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Default config location: ~/.config/calgo/config.yaml
		home, err := os.UserHomeDir()
		if err == nil {
			configDir := filepath.Join(home, ".config", "calgo")
			v.AddConfigPath(configDir)
			v.SetConfigName("config")
			v.SetConfigType("yaml")
		}
	}

	// Read config file (ignore if not found)
	if err := v.ReadInConfig(); err != nil {
		// Only return error if it's not a "file not found" error
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) && !os.IsNotExist(err) {
			// Check if it's a parsing error
			if _, ok := err.(viper.ConfigParseError); ok {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
		}
	}

	// Bind environment variables
	v.SetEnvPrefix("")
	v.AutomaticEnv()

	// Map environment variables to config keys
	v.BindEnv("credentials_path", "GOOGLE_CALENDAR_CREDENTIALS")
	v.BindEnv("token_path", "GOOGLE_CALENDAR_TOKEN")
	v.BindEnv("calendar_id", "GOOGLE_CALENDAR_ID")
	v.BindEnv("timezone", "TZ")

	// Apply flag overrides (highest priority)
	for key, value := range flagOverrides {
		if value != nil && value != "" {
			v.Set(key, value)
		}
	}

	// Unmarshal into Config struct
	cfg := DefaultConfig()
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

// Validate checks that all required configuration values are present.
func (c *Config) Validate() error {
	if c.CredentialsPath == "" {
		return ErrMissingCredentialsPath
	}

	if c.TokenPath == "" {
		return ErrMissingTokenPath
	}

	return nil
}

// ValidateCredentialsExist checks if the credentials file exists.
func (c *Config) ValidateCredentialsExist() error {
	if _, err := os.Stat(c.CredentialsPath); os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", ErrCredentialsNotFound, c.CredentialsPath)
	}
	return nil
}

// GetConfigDir returns the default configuration directory path.
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", "calgo"), nil
}

// EnsureConfigDir creates the configuration directory if it doesn't exist.
func EnsureConfigDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return configDir, nil
}
