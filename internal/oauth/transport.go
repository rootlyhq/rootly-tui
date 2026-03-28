package oauth

import (
	"context"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"

	"github.com/rootlyhq/rootly-tui/internal/debug"
)

// NewHTTPClientWithTokens creates an http.Client that automatically injects Bearer tokens,
// refreshes expired tokens, retries on 401, and persists refreshed tokens to disk.
func NewHTTPClientWithTokens(cfg *oauth2.Config, td *TokenData, baseTransport http.RoundTripper, userAgent string) *http.Client {
	if td == nil || !td.HasValidTokens() {
		return &http.Client{Transport: baseTransport}
	}

	tok := td.ToOAuth2Token()
	uaBase := &userAgentTransport{base: baseTransport, userAgent: userAgent}

	return &http.Client{
		Transport: &retryOn401Transport{
			cfg:  cfg,
			tok:  tok,
			base: uaBase,
		},
	}
}

// retryOn401Transport handles Bearer token injection, auto-refresh on expiry,
// and forced refresh + retry when the server returns 401.
type retryOn401Transport struct {
	cfg  *oauth2.Config
	tok  *oauth2.Token
	base http.RoundTripper
	mu   sync.Mutex
}

func (t *retryOn401Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	tok := t.tok
	t.mu.Unlock()

	debug.Logger.Debug("OAuth transport",
		"url", req.URL.String(),
		"token_valid", tok.Valid(),
		"token_expiry", tok.Expiry.Format(time.RFC3339),
		"token_type", tok.TokenType,
		"access_token_len", len(tok.AccessToken),
		"refresh_token_len", len(tok.RefreshToken),
		"has_access_token", tok.AccessToken != "",
		"has_refresh_token", tok.RefreshToken != "",
	)

	// Refresh proactively if expired
	if !tok.Valid() {
		debug.Logger.Info("OAuth token expired, refreshing proactively")
		newTok, err := t.refresh(tok)
		if err != nil {
			debug.Logger.Error("OAuth proactive refresh failed", "error", err)
			return nil, err
		}
		debug.Logger.Info("OAuth proactive refresh succeeded",
			"new_expiry", newTok.Expiry.Format(time.RFC3339),
			"new_access_token_len", len(newTok.AccessToken),
		)
		tok = newTok
	}

	// Set auth header and execute
	req2 := req.Clone(req.Context())
	tok.SetAuthHeader(req2)
	debug.Logger.Debug("OAuth request", "auth_header", req2.Header.Get("Authorization")[:min(30, len(req2.Header.Get("Authorization")))]+"...")
	resp, err := t.base.RoundTrip(req2)
	if err != nil {
		debug.Logger.Error("OAuth request failed", "error", err)
		return nil, err
	}

	debug.Logger.Debug("OAuth response", "status", resp.StatusCode, "url", req.URL.String())

	// On 401, force refresh and retry once
	if resp.StatusCode == http.StatusUnauthorized {
		debug.Logger.Warn("OAuth got 401, force-refreshing token",
			"token_endpoint", t.cfg.Endpoint.TokenURL,
			"refresh_token_len", len(tok.RefreshToken),
		)
		_ = resp.Body.Close()
		newTok, err := t.forceRefresh(tok)
		if err != nil {
			debug.Logger.Error("OAuth force refresh failed", "error", err)
			// Can't refresh — re-execute with same token to return 401 to caller
			req2b := req.Clone(req.Context())
			tok.SetAuthHeader(req2b)
			return t.base.RoundTrip(req2b)
		}
		debug.Logger.Info("OAuth force refresh succeeded, retrying request",
			"new_expiry", newTok.Expiry.Format(time.RFC3339),
		)
		req3 := req.Clone(req.Context())
		newTok.SetAuthHeader(req3)
		return t.base.RoundTrip(req3)
	}

	return resp, nil
}

func (t *retryOn401Transport) refresh(tok *oauth2.Token) (*oauth2.Token, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Double-check: another goroutine may have refreshed already
	if t.tok.AccessToken != tok.AccessToken {
		return t.tok, nil
	}

	src := t.cfg.TokenSource(context.Background(), tok)
	newTok, err := src.Token()
	if err != nil {
		return nil, err
	}

	t.tok = newTok
	_ = SaveOAuth2Token(newTok)
	return newTok, nil
}

func (t *retryOn401Transport) forceRefresh(tok *oauth2.Token) (*oauth2.Token, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// If another goroutine already refreshed past this token, use that
	if t.tok.AccessToken != tok.AccessToken {
		return t.tok, nil
	}

	// Force refresh by creating a token source with an expired token
	expired := &oauth2.Token{
		RefreshToken: tok.RefreshToken,
		TokenType:    tok.TokenType,
	}
	src := t.cfg.TokenSource(context.Background(), expired)
	newTok, err := src.Token()
	if err != nil {
		return nil, err
	}

	t.tok = newTok
	_ = SaveOAuth2Token(newTok)
	return newTok, nil
}

// userAgentTransport sets User-Agent on all requests.
type userAgentTransport struct {
	base      http.RoundTripper
	userAgent string
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.Header.Set("User-Agent", t.userAgent)
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req2)
}
