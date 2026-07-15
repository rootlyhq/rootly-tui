package oauth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rootlyhq/rootly-tui/internal/config"
	"github.com/rootlyhq/rootly-tui/internal/debug"
)

func TestBearerTokenNotLoggedAsPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := &config.Config{Endpoint: "api.rootly.com"}
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}

	debug.ClearLogs()
	debug.Enable()
	defer debug.Disable()

	accessToken := "super-secret-access-token-value-1234567890"

	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			t.Error("expected Bearer auth header")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer apiSrv.Close()

	td := &TokenData{
		AccessToken:  accessToken,
		RefreshToken: "test-refresh",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}

	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"new","token_type":"Bearer","expires_in":3600,"refresh_token":"r"}`))
	}))
	defer tokenSrv.Close()

	oauthCfg := NewConfig(tokenSrv.URL, "test-client", nil)
	client := NewHTTPClientWithTokens(oauthCfg, td, http.DefaultTransport, "rootly-tui/test")

	req, _ := http.NewRequestWithContext(t.Context(), "GET", apiSrv.URL, http.NoBody)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	defer resp.Body.Close()

	logs := debug.GetLogs()

	for _, entry := range logs {
		if strings.Contains(entry, accessToken) {
			t.Errorf("log entry contains full token: %s", entry)
		}
		if strings.Contains(entry, accessToken[:10]) {
			t.Errorf("log entry contains token prefix: %s", entry)
		}
		if strings.Contains(entry, accessToken[len(accessToken)-4:]) {
			t.Errorf("log entry contains token suffix: %s", entry)
		}
	}
}
