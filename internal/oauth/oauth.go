package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/oauth2"
)

const (
	ClientID     = "rootly-tui"
	RedirectURL  = "http://localhost:19798/callback"
	CallbackPort = ":19798"
)

var Scopes = []string{"openid", "profile", "email", "all"}

// NewConfig creates an OAuth2 config for the given auth base URL.
func NewConfig(authBaseURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:    ClientID,
		RedirectURL: RedirectURL,
		Scopes:      Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:   authBaseURL + "/oauth/authorize",
			TokenURL:  authBaseURL + "/oauth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
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
