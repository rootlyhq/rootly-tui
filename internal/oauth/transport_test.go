package oauth

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

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

func TestNewHTTPClientWithTokensNoTokens(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := &config.Config{Endpoint: "api.rootly.com"}
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}

	oauthCfg := NewConfig("https://app.rootly.com", "test-client")
	client := NewHTTPClientWithTokens(oauthCfg, nil, http.DefaultTransport, "rootly-tui/test")

	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewHTTPClientWithTokensValid(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

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

	td := &TokenData{
		AccessToken:  "test-access",
		RefreshToken: "test-refresh",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}

	// Token server for potential refresh
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"refreshed","token_type":"Bearer","expires_in":3600,"refresh_token":"new-refresh"}`))
	}))
	defer tokenSrv.Close()

	oauthCfg := NewConfig(tokenSrv.URL)
	client := NewHTTPClientWithTokens(oauthCfg, td, http.DefaultTransport, "rootly-tui/test")

	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestRetryOn401(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := &config.Config{Endpoint: "api.rootly.com"}
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}

	var apiCalls atomic.Int32

	// API server that returns 401 on first call, 200 on second
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call := apiCalls.Add(1)
		if call == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// Verify the refreshed token is used
		auth := r.Header.Get("Authorization")
		if auth != "Bearer refreshed-token" {
			t.Errorf("expected refreshed token, got %q", auth)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer apiSrv.Close()

	// Token server that issues a new token
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"refreshed-token","token_type":"Bearer","expires_in":3600,"refresh_token":"new-refresh"}`))
	}))
	defer tokenSrv.Close()

	td := &TokenData{
		AccessToken:  "old-token",
		RefreshToken: "test-refresh",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}

	oauthCfg := NewConfig(tokenSrv.URL)
	client := NewHTTPClientWithTokens(oauthCfg, td, http.DefaultTransport, "rootly-tui/test")

	req, _ := http.NewRequestWithContext(t.Context(), "GET", apiSrv.URL, http.NoBody)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 after retry, got %d", resp.StatusCode)
	}
	if apiCalls.Load() != 2 {
		t.Errorf("expected 2 API calls (initial + retry), got %d", apiCalls.Load())
	}
}
