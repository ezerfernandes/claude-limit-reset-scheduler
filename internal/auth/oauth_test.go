package auth

import (
	"encoding/json"
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
