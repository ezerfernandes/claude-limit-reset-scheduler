// authtest is a simple program to test OAuth2 authentication.
// Run: go run ./cmd/authtest
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ezer/calgo/internal/auth"
	"github.com/ezer/calgo/internal/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load("", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Use default paths if not set
	if cfg.CredentialsPath == "" {
		cfg.CredentialsPath = "./credentials.json"
	}
	if cfg.TokenPath == "" {
		cfg.TokenPath = "./token.json"
	}

	fmt.Printf("Credentials path: %s\n", cfg.CredentialsPath)
	fmt.Printf("Token path: %s\n", cfg.TokenPath)

	// Check if credentials file exists
	if _, err := os.Stat(cfg.CredentialsPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Credentials file not found: %s\n", cfg.CredentialsPath)
		os.Exit(1)
	}

	// Create authenticator
	authenticator := auth.NewAuthenticator(cfg.CredentialsPath, cfg.TokenPath)

	// Load credentials
	fmt.Println("\nLoading credentials...")
	if err := authenticator.LoadCredentials(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load credentials: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Credentials loaded successfully!")

	// Check for existing token
	if authenticator.HasSavedToken() {
		fmt.Println("\nExisting token found. Testing token validity...")
	} else {
		fmt.Println("\nNo existing token. Will initiate OAuth flow...")
	}

	// Get token (this will trigger OAuth flow if needed)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Println("\nGetting OAuth token...")
	token, err := authenticator.GetToken(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nAuthentication successful!")
	fmt.Printf("Access token: %s...\n", token.AccessToken[:min(20, len(token.AccessToken))])
	fmt.Printf("Token type: %s\n", token.TokenType)
	fmt.Printf("Expiry: %s\n", token.Expiry.Format(time.RFC3339))

	if token.RefreshToken != "" {
		fmt.Printf("Refresh token: %s...\n", token.RefreshToken[:min(20, len(token.RefreshToken))])
	}

	fmt.Println("\nToken saved to:", cfg.TokenPath)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
