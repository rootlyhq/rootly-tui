package oauth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/rootlyhq/rootly-tui/internal/config"
	"github.com/rootlyhq/rootly-tui/internal/debug"
)

const (
	RedirectURL  = "http://localhost:19798/callback"
	CallbackPort = ":19798"
	ClientName   = "Rootly TUI"
)

var Scopes = []string{"openid", "profile", "email", "all"}

// NewConfig creates an OAuth2 config for the given auth base URL and client ID.
func NewConfig(authBaseURL, clientID string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:    clientID,
		RedirectURL: RedirectURL,
		Scopes:      Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:   authBaseURL + "/oauth/authorize",
			TokenURL:  authBaseURL + "/oauth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}

// registrationRequest is the body for POST /oauth/register.
type registrationRequest struct {
	ClientName              string   `json:"client_name"`
	RedirectURIs            []string `json:"redirect_uris"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
}

// registrationResponse is the response from POST /oauth/register.
type registrationResponse struct {
	ClientID string `json:"client_id"`
}

// RegisterClient dynamically registers an OAuth client and returns the client_id.
// The client_id is also saved to config for future use.
func RegisterClient(ctx context.Context, authBaseURL string) (string, error) {
	reqBody := registrationRequest{
		ClientName:              ClientName,
		RedirectURIs:            []string{RedirectURL},
		TokenEndpointAuthMethod: "none",
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal registration request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authBaseURL+"/oauth/register", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create registration request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not register OAuth client: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MB max

	if resp.StatusCode != http.StatusCreated {
		debug.Logger.Error("OAuth client registration failed",
			"status", resp.StatusCode,
			"body", string(respBody),
		)
		return "", fmt.Errorf("could not register OAuth client (HTTP %d)", resp.StatusCode)
	}

	var regResp registrationResponse
	if err := json.Unmarshal(respBody, &regResp); err != nil {
		return "", fmt.Errorf("failed to parse registration response: %w", err)
	}

	if regResp.ClientID == "" {
		return "", fmt.Errorf("registration response missing client_id")
	}

	debug.Logger.Info("OAuth client registered", "client_id", regResp.ClientID)

	if err := saveClientID(regResp.ClientID); err != nil {
		return "", fmt.Errorf("failed to save client_id: %w", err)
	}

	return regResp.ClientID, nil
}

// LoadClientID loads the cached OAuth client ID from config.
func LoadClientID() string {
	cfg, err := config.Load()
	if err != nil {
		return ""
	}
	return cfg.OAuthClientID
}

// ClearClientID removes the cached client ID from config.
func ClearClientID() error {
	cfg, err := config.Load()
	if err != nil {
		return nil
	}
	cfg.ClearOAuthClientID()
	return config.Save(cfg)
}

func saveClientID(clientID string) error {
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}
	cfg.OAuthClientID = clientID
	return config.Save(cfg)
}

// GenerateState generates a random state string for CSRF protection.
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// ExchangeCode exchanges an authorization code for tokens using PKCE.
func ExchangeCode(ctx context.Context, cfg *oauth2.Config, code, verifier string) (*oauth2.Token, error) {
	return cfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
}

// DeriveAuthBaseURL derives the OAuth authorization base URL from an API host.
// It strips any path (e.g. /api) and uses http:// for local hosts.
func DeriveAuthBaseURL(apiHost string) string {
	// If it already has a scheme, strip path and return scheme+host
	if strings.HasPrefix(apiHost, "http://") || strings.HasPrefix(apiHost, "https://") {
		return stripPath(apiHost)
	}

	// Extract host part (strip /api or other paths)
	host := apiHost
	if idx := strings.Index(host, "/"); idx != -1 {
		host = host[:idx]
	}

	if isLocalHost(host) {
		return "http://" + host
	}
	return "https://" + host
}

// isLocalHost returns true for localhost or 127.0.0.1.
func isLocalHost(host string) bool {
	// Strip port to check hostname
	hostname := host
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		hostname = host[:idx]
	}
	return hostname == "localhost" || hostname == "127.0.0.1"
}

// stripPath removes path from a URL, keeping only scheme + host.
func stripPath(u string) string {
	// Find scheme end
	schemeEnd := strings.Index(u, "://")
	if schemeEnd == -1 {
		return u
	}
	rest := u[schemeEnd+3:]
	// Find path start
	if idx := strings.Index(rest, "/"); idx != -1 {
		return u[:schemeEnd+3+idx]
	}
	return u
}
