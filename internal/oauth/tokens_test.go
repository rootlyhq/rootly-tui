package oauth

import (
	"testing"
	"time"

	"golang.org/x/oauth2"

	"github.com/rootlyhq/rootly-tui/internal/config"
)

func setupTestConfig(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create a base config so Load() works
	cfg := &config.Config{Endpoint: "api.rootly.com"}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("failed to save base config: %v", err)
	}
}

func TestTokensSaveLoadRoundtrip(t *testing.T) {
	setupTestConfig(t)

	td := &TokenData{
		AccessToken:  "access-123",
		RefreshToken: "refresh-456",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(1 * time.Hour).Truncate(time.Second),
	}

	if err := SaveTokens(td); err != nil {
		t.Fatalf("SaveTokens() error: %v", err)
	}

	loaded, err := LoadTokens()
	if err != nil {
		t.Fatalf("LoadTokens() error: %v", err)
	}

	if loaded.AccessToken != td.AccessToken {
		t.Errorf("AccessToken = %q, want %q", loaded.AccessToken, td.AccessToken)
	}
	if loaded.RefreshToken != td.RefreshToken {
		t.Errorf("RefreshToken = %q, want %q", loaded.RefreshToken, td.RefreshToken)
	}
	if loaded.TokenType != td.TokenType {
		t.Errorf("TokenType = %q, want %q", loaded.TokenType, td.TokenType)
	}
	if !loaded.ExpiresAt.Equal(td.ExpiresAt) {
		t.Errorf("ExpiresAt = %v, want %v", loaded.ExpiresAt, td.ExpiresAt)
	}

	// Verify config file has use_oauth set
	cfg, _ := config.Load()
	if !cfg.UseOAuth {
		t.Error("expected UseOAuth to be true after saving tokens")
	}
	if cfg.OAuthAccessToken != td.AccessToken {
		t.Error("expected config to contain access token")
	}
}

func TestClearTokens(t *testing.T) {
	setupTestConfig(t)

	td := &TokenData{AccessToken: "test", RefreshToken: "test", TokenType: "Bearer"}
	if err := SaveTokens(td); err != nil {
		t.Fatalf("SaveTokens() error: %v", err)
	}

	if err := ClearTokens(); err != nil {
		t.Fatalf("ClearTokens() error: %v", err)
	}

	loaded, err := LoadTokens()
	if err != nil {
		t.Fatalf("LoadTokens() after clear error: %v", err)
	}
	if loaded.HasValidTokens() {
		t.Error("expected no valid tokens after clear")
	}

	cfg, _ := config.Load()
	if cfg.UseOAuth {
		t.Error("expected UseOAuth to be false after clear")
	}
}

func TestClearTokensWhenNoConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	if err := ClearTokens(); err != nil {
		t.Errorf("ClearTokens() with no config should return nil, got: %v", err)
	}
}

func TestIsExpired(t *testing.T) {
	// Expired token
	td := &TokenData{ExpiresAt: time.Now().Add(-1 * time.Minute)}
	if !td.IsExpired() {
		t.Error("expected token to be expired")
	}

	// Token expiring within 30s buffer
	td = &TokenData{ExpiresAt: time.Now().Add(15 * time.Second)}
	if !td.IsExpired() {
		t.Error("expected token to be expired (within 30s buffer)")
	}

	// Valid token
	td = &TokenData{ExpiresAt: time.Now().Add(1 * time.Hour)}
	if td.IsExpired() {
		t.Error("expected token to be valid")
	}
}

func TestTokenDataOAuth2Conversion(t *testing.T) {
	tok := &oauth2.Token{
		AccessToken:  "access",
		RefreshToken: "refresh",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(1 * time.Hour),
	}

	td := TokenDataFromOAuth2(tok)
	if td.AccessToken != tok.AccessToken {
		t.Error("AccessToken mismatch")
	}

	back := td.ToOAuth2Token()
	if back.AccessToken != tok.AccessToken {
		t.Error("roundtrip AccessToken mismatch")
	}
	if back.RefreshToken != tok.RefreshToken {
		t.Error("roundtrip RefreshToken mismatch")
	}
}

func TestHasValidTokens(t *testing.T) {
	td := &TokenData{AccessToken: "a", RefreshToken: "r"}
	if !td.HasValidTokens() {
		t.Error("expected valid tokens")
	}

	td = &TokenData{AccessToken: "", RefreshToken: "r"}
	if td.HasValidTokens() {
		t.Error("expected invalid tokens (no access token)")
	}

	td = &TokenData{AccessToken: "a", RefreshToken: ""}
	if td.HasValidTokens() {
		t.Error("expected invalid tokens (no refresh token)")
	}
}

func TestSaveOAuth2Token(t *testing.T) {
	setupTestConfig(t)

	tok := &oauth2.Token{
		AccessToken:  "at",
		RefreshToken: "rt",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(1 * time.Hour).Truncate(time.Second),
	}

	if err := SaveOAuth2Token(tok); err != nil {
		t.Fatalf("SaveOAuth2Token() error: %v", err)
	}

	loaded, err := LoadTokens()
	if err != nil {
		t.Fatalf("LoadTokens() error: %v", err)
	}
	if loaded.AccessToken != "at" {
		t.Errorf("AccessToken = %q, want %q", loaded.AccessToken, "at")
	}
}
