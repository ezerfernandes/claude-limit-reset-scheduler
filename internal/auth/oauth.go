package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// Scopes required for Google Calendar access.
var Scopes = []string{
	calendar.CalendarEventsScope,
}

// Errors for authentication.
var (
	ErrInvalidCredentials  = errors.New("invalid credentials file format")
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrTokenRefreshFailed   = errors.New("token refresh failed")
)

// Authenticator handles OAuth2 authentication with Google.
type Authenticator struct {
	credentialsPath string
	tokenPath       string
	config          *oauth2.Config
}

// NewAuthenticator creates a new Authenticator with the given paths.
func NewAuthenticator(credentialsPath, tokenPath string) *Authenticator {
	return &Authenticator{
		credentialsPath: credentialsPath,
		tokenPath:       tokenPath,
	}
}

// LoadCredentials reads and parses the OAuth2 credentials file.
func (a *Authenticator) LoadCredentials() error {
	data, err := os.ReadFile(a.credentialsPath)
	if err != nil {
		return fmt.Errorf("failed to read credentials file: %w", err)
	}

	config, err := google.ConfigFromJSON(data, Scopes...)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidCredentials, err)
	}

	a.config = config
	return nil
}

// GetToken returns a valid OAuth2 token, either from cache or by authenticating.
func (a *Authenticator) GetToken(ctx context.Context) (*oauth2.Token, error) {
	if a.config == nil {
		if err := a.LoadCredentials(); err != nil {
			return nil, err
		}
	}

	// Try to load existing token
	token, err := a.loadToken()
	if err == nil {
		// Check if token needs refresh
		if token.Valid() {
			return token, nil
		}

		// Try to refresh the token
		tokenSource := a.config.TokenSource(ctx, token)
		newToken, err := tokenSource.Token()
		if err == nil {
			// Save refreshed token
			if saveErr := a.saveToken(newToken); saveErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to save refreshed token: %v\n", saveErr)
			}
			return newToken, nil
		}

		// Refresh failed, need to re-authenticate
		fmt.Println("Token refresh failed. Re-authentication required.")
	}

	// No valid token, need to authenticate
	return a.authenticate(ctx)
}

// GetClient returns an HTTP client configured with OAuth2 credentials.
func (a *Authenticator) GetClient(ctx context.Context) (*http.Client, error) {
	token, err := a.GetToken(ctx)
	if err != nil {
		return nil, err
	}

	return a.config.Client(ctx, token), nil
}

// authenticate performs the OAuth2 authentication flow.
func (a *Authenticator) authenticate(ctx context.Context) (*oauth2.Token, error) {
	// Create a channel to receive the authorization code
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Start local server to handle callback
	server, port, err := a.startCallbackServer(codeChan, errChan)
	if err != nil {
		return nil, fmt.Errorf("failed to start callback server: %w", err)
	}
	defer server.Close()

	// Update redirect URI to use the actual port
	a.config.RedirectURL = fmt.Sprintf("http://localhost:%d", port)

	// Generate authorization URL
	authURL := a.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Println("Opening browser for authentication...")
	fmt.Printf("If the browser doesn't open, visit this URL:\n%s\n\n", authURL)

	// Open browser
	if err := openBrowser(authURL); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to open browser: %v\n", err)
	}

	// Wait for the authorization code
	fmt.Println("Waiting for authorization...")
	var code string
	select {
	case code = <-codeChan:
		// Got the code
	case err := <-errChan:
		return nil, fmt.Errorf("%w: %v", ErrAuthenticationFailed, err)
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("%w: timeout waiting for authorization", ErrAuthenticationFailed)
	}

	// Exchange code for token
	token, err := a.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAuthenticationFailed, err)
	}

	// Save the token
	if err := a.saveToken(token); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save token: %v\n", err)
	}

	fmt.Println("Authentication successful!")
	return token, nil
}

// startCallbackServer starts a local HTTP server to handle the OAuth2 callback.
func (a *Authenticator) startCallbackServer(codeChan chan<- string, errChan chan<- error) (*http.Server, int, error) {
	// Find an available port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, 0, err
	}

	port := listener.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errMsg := r.URL.Query().Get("error")
			if errMsg == "" {
				errMsg = "no authorization code received"
			}
			errChan <- errors.New(errMsg)
			http.Error(w, "Authorization failed", http.StatusBadRequest)
			return
		}

		codeChan <- code
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head><title>Authorization Successful</title></head>
<body style="font-family: sans-serif; text-align: center; padding: 50px;">
<h1>Authorization Successful!</h1>
<p>You can close this window and return to the terminal.</p>
</body>
</html>
`)
	})

	server := &http.Server{Handler: mux}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	return server, port, nil
}

// loadToken reads the OAuth2 token from the token file.
func (a *Authenticator) loadToken() (*oauth2.Token, error) {
	data, err := os.ReadFile(a.tokenPath)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token file: %w", err)
	}

	return &token, nil
}

// saveToken writes the OAuth2 token to the token file.
func (a *Authenticator) saveToken(token *oauth2.Token) error {
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	if err := os.WriteFile(a.tokenPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

// openBrowser opens the specified URL in the default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// ClearToken removes the saved token file.
func (a *Authenticator) ClearToken() error {
	if err := os.Remove(a.tokenPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove token file: %w", err)
	}
	return nil
}

// HasSavedToken checks if a token file exists.
func (a *Authenticator) HasSavedToken() bool {
	_, err := os.Stat(a.tokenPath)
	return err == nil
}
