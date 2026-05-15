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

var DefaultScopes = []string{"openid", "profile", "email", "all"}

// NewConfig creates an OAuth2 config for the given auth base URL, client ID, and scopes.
func NewConfig(authBaseURL, clientID string, scopes []string) *oauth2.Config {
	if len(scopes) == 0 {
		scopes = DefaultScopes
	}
	return &oauth2.Config{
		ClientID:    clientID,
		RedirectURL: RedirectURL,
		Scopes:      scopes,
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
	Scope    string `json:"scope"`
}

// ClientRegistration holds the cached result of dynamic client registration.
type ClientRegistration struct {
	ClientID string
	Scopes   []string
}

// RegisterClient dynamically registers an OAuth client and returns the registration.
// The client_id and scopes are saved to config for future use.
func RegisterClient(ctx context.Context, authBaseURL string) (*ClientRegistration, error) {
	reqBody := registrationRequest{
		ClientName:              ClientName,
		RedirectURIs:            []string{RedirectURL},
		TokenEndpointAuthMethod: "none",
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal registration request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authBaseURL+"/oauth/register", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create registration request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not register OAuth client: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode != http.StatusCreated {
		debug.Logger.Error("OAuth client registration failed",
			"status", resp.StatusCode,
			"body", string(respBody),
		)
		return nil, fmt.Errorf("could not register OAuth client (HTTP %d)", resp.StatusCode)
	}

	var regResp registrationResponse
	if err := json.Unmarshal(respBody, &regResp); err != nil {
		return nil, fmt.Errorf("failed to parse registration response: %w", err)
	}

	if regResp.ClientID == "" {
		return nil, fmt.Errorf("registration response missing client_id")
	}

	scopes := strings.Fields(regResp.Scope)
	if len(scopes) == 0 {
		scopes = DefaultScopes
	}

	debug.Logger.Info("OAuth client registered", "client_id", regResp.ClientID, "scopes", regResp.Scope)

	if err := saveClientRegistration(regResp.ClientID, scopes); err != nil {
		return nil, fmt.Errorf("failed to save client registration: %w", err)
	}

	return &ClientRegistration{ClientID: regResp.ClientID, Scopes: scopes}, nil
}

// LoadClientRegistration loads the cached OAuth client registration from config.
func LoadClientRegistration() *ClientRegistration {
	cfg, err := config.Load()
	if err != nil || cfg.OAuthClientID == "" {
		return nil
	}
	scopes := strings.Fields(cfg.OAuthScopes)
	if len(scopes) == 0 {
		scopes = DefaultScopes
	}
	return &ClientRegistration{ClientID: cfg.OAuthClientID, Scopes: scopes}
}

// ClearClientRegistration removes the cached client registration from config.
func ClearClientRegistration() error {
	cfg, err := config.Load()
	if err != nil {
		return nil
	}
	cfg.ClearOAuthClientID()
	return config.Save(cfg)
}

func saveClientRegistration(clientID string, scopes []string) error {
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}
	cfg.OAuthClientID = clientID
	cfg.OAuthScopes = strings.Join(scopes, " ")
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

// DeriveAPIBaseURL derives the API base URL from an API host config value.
// Used for server-to-server calls (e.g. POST /oauth/register).
func DeriveAPIBaseURL(apiHost string) string {
	if strings.HasPrefix(apiHost, "http://") || strings.HasPrefix(apiHost, "https://") {
		return stripPath(apiHost)
	}

	host := apiHost
	if idx := strings.Index(host, "/"); idx != -1 {
		host = host[:idx]
	}

	if isLocalHost(host) {
		return "http://" + host
	}
	return "https://" + host
}

// DeriveAuthBaseURL derives the OAuth authorization base URL for browser-facing URLs.
// Strips "api." prefix from the host (e.g. api.rootly.com → rootly.com).
// Localhost is unchanged.
func DeriveAuthBaseURL(apiHost string) string {
	apiBase := DeriveAPIBaseURL(apiHost)

	schemeEnd := strings.Index(apiBase, "://")
	if schemeEnd == -1 {
		return apiBase
	}
	scheme := apiBase[:schemeEnd+3]
	host := apiBase[schemeEnd+3:]

	if isLocalHost(host) {
		return apiBase
	}

	host = strings.TrimPrefix(host, "api.")
	return scheme + host
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
