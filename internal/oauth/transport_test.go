package oauth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/oauth2"

	"github.com/rootlyhq/rootly-tui/internal/config"
)

func TestUserAgentTransport(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	tr := &userAgentTransport{base: http.DefaultTransport, userAgent: "rootly-tui/test"}
	client := &http.Client{Transport: tr}

	req, _ := http.NewRequestWithContext(t.Context(), "GET", srv.URL, http.NoBody)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	defer resp.Body.Close()

	if gotUA != "rootly-tui/test" {
		t.Errorf("User-Agent = %q, want %q", gotUA, "rootly-tui/test")
	}
}

func TestUserAgentTransportNilBase(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	tr := &userAgentTransport{base: nil, userAgent: "test"}
	client := &http.Client{Transport: tr}

	req, _ := http.NewRequestWithContext(t.Context(), "GET", srv.URL, http.NoBody)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestNewHTTPClientNoTokens(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := &config.Config{Endpoint: "api.rootly.com"}
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}

	oauthCfg := NewConfig("https://app.rootly.com")
	td, _ := LoadTokens()
	client := NewHTTPClientWithTokens(oauthCfg, td, http.DefaultTransport, "rootly-tui/test")

	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewHTTPClientWithTokens(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Start a token server for refresh
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"refreshed","token_type":"Bearer","expires_in":3600,"refresh_token":"new-refresh"}`))
	}))
	defer tokenSrv.Close()

	cfg := &config.Config{
		Endpoint:          "api.rootly.com",
		UseOAuth:          true,
		OAuthAccessToken:  "test-access",
		OAuthRefreshToken: "test-refresh",
		OAuthTokenType:    "Bearer",
		OAuthExpiresAt:    time.Now().Add(1 * time.Hour),
	}
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}

	oauthCfg := NewConfig(tokenSrv.URL)
	td2, _ := LoadTokens()
	client := NewHTTPClientWithTokens(oauthCfg, td2, http.DefaultTransport, "rootly-tui/test")

	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestPersistingTokenSource(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := &config.Config{Endpoint: "api.rootly.com"}
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}

	tok := &oauth2.Token{
		AccessToken:  "new-access",
		RefreshToken: "new-refresh",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(1 * time.Hour),
	}

	src := &persistingTokenSource{
		src: oauth2.StaticTokenSource(tok),
	}

	got, err := src.Token()
	if err != nil {
		t.Fatalf("Token() error: %v", err)
	}
	if got.AccessToken != "new-access" {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, "new-access")
	}

	// Verify token was persisted to config
	loaded, err := LoadTokens()
	if err != nil {
		t.Fatalf("LoadTokens() error: %v", err)
	}
	if loaded.AccessToken != "new-access" {
		t.Errorf("persisted AccessToken = %q, want %q", loaded.AccessToken, "new-access")
	}
}
