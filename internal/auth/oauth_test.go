package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

// Sample credentials for testing (not real credentials)
const testCredentials = `{
	"installed": {
		"client_id": "test-client-id.apps.googleusercontent.com",
		"project_id": "test-project",
		"auth_uri": "https://accounts.google.com/o/oauth2/auth",
		"token_uri": "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_secret": "test-client-secret",
		"redirect_uris": ["http://localhost"]
	}
}`

func TestNewAuthenticator(t *testing.T) {
	auth := NewAuthenticator("/path/to/creds.json", "/path/to/token.json")

	if auth.credentialsPath != "/path/to/creds.json" {
		t.Errorf("Expected credentialsPath to be '/path/to/creds.json', got '%s'", auth.credentialsPath)
	}

	if auth.tokenPath != "/path/to/token.json" {
		t.Errorf("Expected tokenPath to be '/path/to/token.json', got '%s'", auth.tokenPath)
	}
}

func TestLoadCredentials_Success(t *testing.T) {
	tmpDir := t.TempDir()
	credPath := filepath.Join(tmpDir, "credentials.json")

	if err := os.WriteFile(credPath, []byte(testCredentials), 0644); err != nil {
		t.Fatalf("Failed to write test credentials: %v", err)
	}

	auth := NewAuthenticator(credPath, filepath.Join(tmpDir, "token.json"))
	err := auth.LoadCredentials()

	if err != nil {
		t.Errorf("LoadCredentials failed: %v", err)
	}

	if auth.config == nil {
		t.Error("Expected config to be set after LoadCredentials")
	}
}

func TestLoadCredentials_FileNotFound(t *testing.T) {
	auth := NewAuthenticator("/nonexistent/credentials.json", "/path/to/token.json")
	err := auth.LoadCredentials()

	if err == nil {
		t.Error("Expected error for nonexistent credentials file")
	}
}

func TestLoadCredentials_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	credPath := filepath.Join(tmpDir, "credentials.json")

	// Write invalid JSON
	if err := os.WriteFile(credPath, []byte(`{"invalid": "format"}`), 0644); err != nil {
		t.Fatalf("Failed to write test credentials: %v", err)
	}

	auth := NewAuthenticator(credPath, filepath.Join(tmpDir, "token.json"))
	err := auth.LoadCredentials()

	if err == nil {
		t.Error("Expected error for invalid credentials format")
	}
}

func TestLoadToken_Success(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	token := &oauth2.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}

	data, _ := json.Marshal(token)
	if err := os.WriteFile(tokenPath, data, 0600); err != nil {
		t.Fatalf("Failed to write test token: %v", err)
	}

	auth := NewAuthenticator("/path/to/creds.json", tokenPath)
	loadedToken, err := auth.loadToken()

	if err != nil {
		t.Errorf("loadToken failed: %v", err)
	}

	if loadedToken.AccessToken != "test-access-token" {
		t.Errorf("Expected AccessToken to be 'test-access-token', got '%s'", loadedToken.AccessToken)
	}

	if loadedToken.RefreshToken != "test-refresh-token" {
		t.Errorf("Expected RefreshToken to be 'test-refresh-token', got '%s'", loadedToken.RefreshToken)
	}
}

func TestLoadToken_FileNotFound(t *testing.T) {
	auth := NewAuthenticator("/path/to/creds.json", "/nonexistent/token.json")
	_, err := auth.loadToken()

	if err == nil {
		t.Error("Expected error for nonexistent token file")
	}
}

func TestLoadToken_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	if err := os.WriteFile(tokenPath, []byte("not json"), 0600); err != nil {
		t.Fatalf("Failed to write test token: %v", err)
	}

	auth := NewAuthenticator("/path/to/creds.json", tokenPath)
	_, err := auth.loadToken()

	if err == nil {
		t.Error("Expected error for invalid token format")
	}
}

func TestSaveToken(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	auth := NewAuthenticator("/path/to/creds.json", tokenPath)

	token := &oauth2.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}

	err := auth.saveToken(token)
	if err != nil {
		t.Errorf("saveToken failed: %v", err)
	}

	// Verify the file was created with correct permissions
	info, err := os.Stat(tokenPath)
	if err != nil {
		t.Fatalf("Token file not created: %v", err)
	}

	// Check file permissions (should be 0600)
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected token file permissions to be 0600, got %o", info.Mode().Perm())
	}

	// Verify content
	loadedToken, err := auth.loadToken()
	if err != nil {
		t.Fatalf("Failed to load saved token: %v", err)
	}

	if loadedToken.AccessToken != token.AccessToken {
		t.Errorf("Saved token AccessToken mismatch")
	}
}

func TestHasSavedToken(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	auth := NewAuthenticator("/path/to/creds.json", tokenPath)

	// Initially no token
	if auth.HasSavedToken() {
		t.Error("Expected HasSavedToken to return false when no token exists")
	}

	// Create token file
	if err := os.WriteFile(tokenPath, []byte("{}"), 0600); err != nil {
		t.Fatalf("Failed to create token file: %v", err)
	}

	// Now should have token
	if !auth.HasSavedToken() {
		t.Error("Expected HasSavedToken to return true when token exists")
	}
}

func TestClearToken(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	// Create token file
	if err := os.WriteFile(tokenPath, []byte("{}"), 0600); err != nil {
		t.Fatalf("Failed to create token file: %v", err)
	}

	auth := NewAuthenticator("/path/to/creds.json", tokenPath)

	// Clear token
	err := auth.ClearToken()
	if err != nil {
		t.Errorf("ClearToken failed: %v", err)
	}

	// Verify file is deleted
	if _, err := os.Stat(tokenPath); !os.IsNotExist(err) {
		t.Error("Expected token file to be deleted")
	}
}

func TestClearToken_NonexistentFile(t *testing.T) {
	auth := NewAuthenticator("/path/to/creds.json", "/nonexistent/token.json")

	// Should not error when file doesn't exist
	err := auth.ClearToken()
	if err != nil {
		t.Errorf("ClearToken should not error for nonexistent file: %v", err)
	}
}

func TestScopes(t *testing.T) {
	if len(Scopes) == 0 {
		t.Error("Scopes should not be empty")
	}

	// Should include calendar events scope
	found := false
	for _, scope := range Scopes {
		if scope == "https://www.googleapis.com/auth/calendar.events" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Scopes should include calendar.events scope")
	}
}

func TestStartCallbackServer(t *testing.T) {
	auth := NewAuthenticator("/path/to/creds.json", "/path/to/token.json")

	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	server, port, err := auth.startCallbackServer(codeChan, errChan)
	if err != nil {
		t.Fatalf("startCallbackServer failed: %v", err)
	}
	defer server.Close()

	if port <= 0 {
		t.Errorf("Expected positive port number, got %d", port)
	}
}

func TestCallbackServer_SuccessfulCode(t *testing.T) {
	auth := NewAuthenticator("/path/to/creds.json", "/path/to/token.json")

	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	server, port, err := auth.startCallbackServer(codeChan, errChan)
	if err != nil {
		t.Fatalf("startCallbackServer failed: %v", err)
	}
	defer server.Close()

	// Make request with authorization code
	url := fmt.Sprintf("http://localhost:%d/?code=test-auth-code", port)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check that code was received on channel
	select {
	case code := <-codeChan:
		if code != "test-auth-code" {
			t.Errorf("Expected code 'test-auth-code', got '%s'", code)
		}
	default:
		t.Error("Expected code to be sent on channel")
	}
}

func TestCallbackServer_MissingCode(t *testing.T) {
	auth := NewAuthenticator("/path/to/creds.json", "/path/to/token.json")

	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	server, port, err := auth.startCallbackServer(codeChan, errChan)
	if err != nil {
		t.Fatalf("startCallbackServer failed: %v", err)
	}
	defer server.Close()

	// Make request without code
	url := fmt.Sprintf("http://localhost:%d/", port)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status is error
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	// Check that error was received on channel
	select {
	case receivedErr := <-errChan:
		if receivedErr == nil {
			t.Error("Expected non-nil error")
		}
		if receivedErr.Error() != "no authorization code received" {
			t.Errorf("Expected 'no authorization code received', got '%s'", receivedErr.Error())
		}
	default:
		t.Error("Expected error to be sent on channel")
	}
}

func TestCallbackServer_ErrorFromProvider(t *testing.T) {
	auth := NewAuthenticator("/path/to/creds.json", "/path/to/token.json")

	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	server, port, err := auth.startCallbackServer(codeChan, errChan)
	if err != nil {
		t.Fatalf("startCallbackServer failed: %v", err)
	}
	defer server.Close()

	// Make request with error parameter (simulating OAuth provider error)
	url := fmt.Sprintf("http://localhost:%d/?error=access_denied", port)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status is error
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	// Check that error was received on channel
	select {
	case receivedErr := <-errChan:
		if receivedErr == nil {
			t.Error("Expected non-nil error")
		}
		if receivedErr.Error() != "access_denied" {
			t.Errorf("Expected 'access_denied', got '%s'", receivedErr.Error())
		}
	default:
		t.Error("Expected error to be sent on channel")
	}
}

func TestSaveToken_InvalidPath(t *testing.T) {
	// Try to save to a directory that doesn't exist
	auth := NewAuthenticator("/path/to/creds.json", "/nonexistent/directory/token.json")

	token := &oauth2.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}

	err := auth.saveToken(token)
	if err == nil {
		t.Error("Expected error when saving to invalid path")
	}
}

func TestSaveToken_VerifyJSONFormat(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	auth := NewAuthenticator("/path/to/creds.json", tokenPath)

	expiry := time.Now().Add(time.Hour).Truncate(time.Second)
	token := &oauth2.Token{
		AccessToken:  "my-access-token",
		RefreshToken: "my-refresh-token",
		TokenType:    "Bearer",
		Expiry:       expiry,
	}

	err := auth.saveToken(token)
	if err != nil {
		t.Fatalf("saveToken failed: %v", err)
	}

	// Read raw file content
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		t.Fatalf("Failed to read token file: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("Token file is not valid JSON: %v", err)
	}

	// Verify expected fields exist
	if _, ok := parsed["access_token"]; !ok {
		t.Error("Token file missing 'access_token' field")
	}
	if _, ok := parsed["refresh_token"]; !ok {
		t.Error("Token file missing 'refresh_token' field")
	}
	if _, ok := parsed["token_type"]; !ok {
		t.Error("Token file missing 'token_type' field")
	}
}

func TestLoadToken_ExpiredToken(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	// Create an expired token
	token := &oauth2.Token{
		AccessToken:  "expired-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(-time.Hour), // Expired 1 hour ago
	}

	data, _ := json.Marshal(token)
	if err := os.WriteFile(tokenPath, data, 0600); err != nil {
		t.Fatalf("Failed to write test token: %v", err)
	}

	auth := NewAuthenticator("/path/to/creds.json", tokenPath)
	loadedToken, err := auth.loadToken()

	if err != nil {
		t.Errorf("loadToken should succeed even for expired token: %v", err)
	}

	// Token should be loaded but marked as expired
	if loadedToken.Valid() {
		t.Error("Expected token to be invalid (expired)")
	}

	if loadedToken.AccessToken != "expired-access-token" {
		t.Error("Token should still have correct access token value")
	}
}

func TestLoadToken_EmptyJSON(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	// Write empty JSON object
	if err := os.WriteFile(tokenPath, []byte("{}"), 0600); err != nil {
		t.Fatalf("Failed to write test token: %v", err)
	}

	auth := NewAuthenticator("/path/to/creds.json", tokenPath)
	loadedToken, err := auth.loadToken()

	if err != nil {
		t.Errorf("loadToken should succeed for empty JSON: %v", err)
	}

	// Token should be loaded but invalid (no access token)
	if loadedToken.Valid() {
		t.Error("Expected empty token to be invalid")
	}
}

func TestErrorTypes(t *testing.T) {
	// Verify error variables are properly defined
	if ErrInvalidCredentials == nil {
		t.Error("ErrInvalidCredentials should not be nil")
	}
	if ErrAuthenticationFailed == nil {
		t.Error("ErrAuthenticationFailed should not be nil")
	}
	if ErrTokenRefreshFailed == nil {
		t.Error("ErrTokenRefreshFailed should not be nil")
	}

	// Test error wrapping
	wrappedErr := fmt.Errorf("%w: some detail", ErrInvalidCredentials)
	if !errors.Is(wrappedErr, ErrInvalidCredentials) {
		t.Error("Wrapped error should match ErrInvalidCredentials")
	}
}

func TestLoadCredentials_SetsConfig(t *testing.T) {
	tmpDir := t.TempDir()
	credPath := filepath.Join(tmpDir, "credentials.json")

	if err := os.WriteFile(credPath, []byte(testCredentials), 0644); err != nil {
		t.Fatalf("Failed to write test credentials: %v", err)
	}

	auth := NewAuthenticator(credPath, filepath.Join(tmpDir, "token.json"))

	// Config should be nil before loading
	if auth.config != nil {
		t.Error("Expected config to be nil before LoadCredentials")
	}

	err := auth.LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials failed: %v", err)
	}

	// Config should be set after loading
	if auth.config == nil {
		t.Error("Expected config to be set after LoadCredentials")
	}

	// Verify config has expected values
	if auth.config.ClientID != "test-client-id.apps.googleusercontent.com" {
		t.Errorf("Unexpected ClientID: %s", auth.config.ClientID)
	}
}

func TestLoadCredentials_CalledTwice(t *testing.T) {
	tmpDir := t.TempDir()
	credPath := filepath.Join(tmpDir, "credentials.json")

	if err := os.WriteFile(credPath, []byte(testCredentials), 0644); err != nil {
		t.Fatalf("Failed to write test credentials: %v", err)
	}

	auth := NewAuthenticator(credPath, filepath.Join(tmpDir, "token.json"))

	// Load credentials twice - should work both times
	err := auth.LoadCredentials()
	if err != nil {
		t.Fatalf("First LoadCredentials failed: %v", err)
	}

	err = auth.LoadCredentials()
	if err != nil {
		t.Fatalf("Second LoadCredentials failed: %v", err)
	}
}

func TestHasSavedToken_AfterClear(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	auth := NewAuthenticator("/path/to/creds.json", tokenPath)

	// Create and then clear token
	if err := os.WriteFile(tokenPath, []byte("{}"), 0600); err != nil {
		t.Fatalf("Failed to create token file: %v", err)
	}

	if !auth.HasSavedToken() {
		t.Error("Should have saved token after creation")
	}

	if err := auth.ClearToken(); err != nil {
		t.Fatalf("ClearToken failed: %v", err)
	}

	if auth.HasSavedToken() {
		t.Error("Should not have saved token after clear")
	}
}

func TestCallbackServer_ContentType(t *testing.T) {
	auth := NewAuthenticator("/path/to/creds.json", "/path/to/token.json")

	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	server, port, err := auth.startCallbackServer(codeChan, errChan)
	if err != nil {
		t.Fatalf("startCallbackServer failed: %v", err)
	}
	defer server.Close()

	// Make request with authorization code
	url := fmt.Sprintf("http://localhost:%d/?code=test-code", port)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check Content-Type header
	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("Expected Content-Type 'text/html', got '%s'", contentType)
	}
}

func TestNewAuthenticator_EmptyPaths(t *testing.T) {
	auth := NewAuthenticator("", "")

	if auth.credentialsPath != "" {
		t.Error("Expected empty credentialsPath")
	}
	if auth.tokenPath != "" {
		t.Error("Expected empty tokenPath")
	}
}

func TestLoadToken_TokenWithAllFields(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	expiry := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
	token := &oauth2.Token{
		AccessToken:  "access-123",
		RefreshToken: "refresh-456",
		TokenType:    "Bearer",
		Expiry:       expiry,
	}

	data, _ := json.Marshal(token)
	if err := os.WriteFile(tokenPath, data, 0600); err != nil {
		t.Fatalf("Failed to write test token: %v", err)
	}

	auth := NewAuthenticator("/path/to/creds.json", tokenPath)
	loadedToken, err := auth.loadToken()

	if err != nil {
		t.Fatalf("loadToken failed: %v", err)
	}

	if loadedToken.AccessToken != "access-123" {
		t.Errorf("AccessToken mismatch: got %s", loadedToken.AccessToken)
	}
	if loadedToken.RefreshToken != "refresh-456" {
		t.Errorf("RefreshToken mismatch: got %s", loadedToken.RefreshToken)
	}
	if loadedToken.TokenType != "Bearer" {
		t.Errorf("TokenType mismatch: got %s", loadedToken.TokenType)
	}
	if !loadedToken.Expiry.Equal(expiry) {
		t.Errorf("Expiry mismatch: got %v, want %v", loadedToken.Expiry, expiry)
	}
}

// TestOpenBrowser is a compile-time check that openBrowser exists
// We can't easily test the actual browser opening behavior
func TestOpenBrowser_FunctionExists(t *testing.T) {
	// This test just verifies the function signature compiles
	var _ func(string) error = openBrowser
}

// mockHTTPHandler is a test helper for creating mock handlers
func createTestHandler(codeChan chan<- string, errChan chan<- error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		w.Write([]byte("<html><body>Success</body></html>"))
	}
}

func TestCallbackHandler_UsingTestServer(t *testing.T) {
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	handler := createTestHandler(codeChan, errChan)
	server := httptest.NewServer(handler)
	defer server.Close()

	// Test successful code
	resp, err := http.Get(server.URL + "?code=my-test-code")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	select {
	case code := <-codeChan:
		if code != "my-test-code" {
			t.Errorf("Expected 'my-test-code', got '%s'", code)
		}
	default:
		t.Error("Code not received")
	}
}

func TestGetToken_LoadsCredentialsWhenConfigNil(t *testing.T) {
	tmpDir := t.TempDir()
	credPath := filepath.Join(tmpDir, "credentials.json")
	tokenPath := filepath.Join(tmpDir, "token.json")

	// Write valid credentials
	if err := os.WriteFile(credPath, []byte(testCredentials), 0644); err != nil {
		t.Fatalf("Failed to write test credentials: %v", err)
	}

	// Write a valid, non-expired token
	token := &oauth2.Token{
		AccessToken:  "valid-access-token",
		RefreshToken: "valid-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}
	tokenData, _ := json.Marshal(token)
	if err := os.WriteFile(tokenPath, tokenData, 0600); err != nil {
		t.Fatalf("Failed to write test token: %v", err)
	}

	auth := NewAuthenticator(credPath, tokenPath)

	// Verify config is nil initially
	if auth.config != nil {
		t.Error("Config should be nil initially")
	}

	// GetToken should load credentials and return valid token
	ctx := context.Background()
	resultToken, err := auth.GetToken(ctx)
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	// Config should now be loaded
	if auth.config == nil {
		t.Error("Config should be set after GetToken")
	}

	if resultToken.AccessToken != "valid-access-token" {
		t.Errorf("Expected access token 'valid-access-token', got '%s'", resultToken.AccessToken)
	}
}

func TestGetToken_ReturnsValidCachedToken(t *testing.T) {
	tmpDir := t.TempDir()
	credPath := filepath.Join(tmpDir, "credentials.json")
	tokenPath := filepath.Join(tmpDir, "token.json")

	// Write valid credentials
	if err := os.WriteFile(credPath, []byte(testCredentials), 0644); err != nil {
		t.Fatalf("Failed to write test credentials: %v", err)
	}

	// Write a valid, non-expired token
	token := &oauth2.Token{
		AccessToken:  "cached-access-token",
		RefreshToken: "cached-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(2 * time.Hour),
	}
	tokenData, _ := json.Marshal(token)
	if err := os.WriteFile(tokenPath, tokenData, 0600); err != nil {
		t.Fatalf("Failed to write test token: %v", err)
	}

	auth := NewAuthenticator(credPath, tokenPath)

	// Pre-load credentials
	if err := auth.LoadCredentials(); err != nil {
		t.Fatalf("LoadCredentials failed: %v", err)
	}

	ctx := context.Background()
	resultToken, err := auth.GetToken(ctx)
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if resultToken.AccessToken != "cached-access-token" {
		t.Errorf("Expected cached token, got '%s'", resultToken.AccessToken)
	}
}

func TestGetToken_FailsWithInvalidCredentials(t *testing.T) {
	tmpDir := t.TempDir()
	credPath := filepath.Join(tmpDir, "credentials.json")
	tokenPath := filepath.Join(tmpDir, "token.json")

	// Write invalid credentials
	if err := os.WriteFile(credPath, []byte(`{"invalid": "format"}`), 0644); err != nil {
		t.Fatalf("Failed to write test credentials: %v", err)
	}

	auth := NewAuthenticator(credPath, tokenPath)

	ctx := context.Background()
	_, err := auth.GetToken(ctx)
	if err == nil {
		t.Error("Expected error for invalid credentials")
	}
}

func TestGetToken_FailsWithMissingCredentials(t *testing.T) {
	auth := NewAuthenticator("/nonexistent/credentials.json", "/nonexistent/token.json")

	ctx := context.Background()
	_, err := auth.GetToken(ctx)
	if err == nil {
		t.Error("Expected error for missing credentials")
	}
}

func TestGetClient_ReturnsClientWithValidToken(t *testing.T) {
	tmpDir := t.TempDir()
	credPath := filepath.Join(tmpDir, "credentials.json")
	tokenPath := filepath.Join(tmpDir, "token.json")

	// Write valid credentials
	if err := os.WriteFile(credPath, []byte(testCredentials), 0644); err != nil {
		t.Fatalf("Failed to write test credentials: %v", err)
	}

	// Write a valid, non-expired token
	token := &oauth2.Token{
		AccessToken:  "client-access-token",
		RefreshToken: "client-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}
	tokenData, _ := json.Marshal(token)
	if err := os.WriteFile(tokenPath, tokenData, 0600); err != nil {
		t.Fatalf("Failed to write test token: %v", err)
	}

	auth := NewAuthenticator(credPath, tokenPath)

	ctx := context.Background()
	client, err := auth.GetClient(ctx)
	if err != nil {
		t.Fatalf("GetClient failed: %v", err)
	}

	if client == nil {
		t.Error("Expected non-nil HTTP client")
	}
}

func TestGetClient_FailsWithInvalidCredentials(t *testing.T) {
	auth := NewAuthenticator("/nonexistent/credentials.json", "/nonexistent/token.json")

	ctx := context.Background()
	_, err := auth.GetClient(ctx)
	if err == nil {
		t.Error("Expected error for invalid credentials")
	}
}

func TestClearToken_ReturnsNilForSuccessfulDelete(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token.json")

	// Create a token file
	if err := os.WriteFile(tokenPath, []byte("{}"), 0600); err != nil {
		t.Fatalf("Failed to create token file: %v", err)
	}

	auth := NewAuthenticator("/path/to/creds.json", tokenPath)

	// First call should succeed
	err := auth.ClearToken()
	if err != nil {
		t.Errorf("First ClearToken failed: %v", err)
	}

	// Second call should also succeed (file already gone)
	err = auth.ClearToken()
	if err != nil {
		t.Errorf("Second ClearToken failed: %v", err)
	}
}

func TestSaveToken_CreatesParentDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// Try to save to a nested path that doesn't exist
	tokenPath := filepath.Join(tmpDir, "nested", "dir", "token.json")

	auth := NewAuthenticator("/path/to/creds.json", tokenPath)

	token := &oauth2.Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
	}

	// This should fail because parent directory doesn't exist
	err := auth.saveToken(token)
	if err == nil {
		t.Error("Expected error when parent directory doesn't exist")
	}
}

// Helper for context tests
func TestGetToken_WithCancelledContext(t *testing.T) {
	tmpDir := t.TempDir()
	credPath := filepath.Join(tmpDir, "credentials.json")
	tokenPath := filepath.Join(tmpDir, "token.json")

	// Write valid credentials
	if err := os.WriteFile(credPath, []byte(testCredentials), 0644); err != nil {
		t.Fatalf("Failed to write test credentials: %v", err)
	}

	// Write a valid token so we don't trigger the auth flow
	token := &oauth2.Token{
		AccessToken:  "context-test-token",
		RefreshToken: "context-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}
	tokenData, _ := json.Marshal(token)
	if err := os.WriteFile(tokenPath, tokenData, 0600); err != nil {
		t.Fatalf("Failed to write test token: %v", err)
	}

	auth := NewAuthenticator(credPath, tokenPath)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// GetToken with valid cached token should still work despite cancelled context
	resultToken, err := auth.GetToken(ctx)
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if resultToken.AccessToken != "context-test-token" {
		t.Errorf("Expected token, got '%s'", resultToken.AccessToken)
	}
}

func TestLoadCredentials_PermissionDenied(t *testing.T) {
	// Skip on Windows as permission handling is different
	if os.Getenv("GOOS") == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	tmpDir := t.TempDir()
	credPath := filepath.Join(tmpDir, "credentials.json")

	// Write credentials then make them unreadable
	if err := os.WriteFile(credPath, []byte(testCredentials), 0000); err != nil {
		t.Fatalf("Failed to write test credentials: %v", err)
	}
	defer os.Chmod(credPath, 0644) // Restore for cleanup

	auth := NewAuthenticator(credPath, filepath.Join(tmpDir, "token.json"))
	err := auth.LoadCredentials()

	if err == nil {
		t.Error("Expected error for permission denied")
	}
}
