package oauth

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
)

// NewHTTPClient creates an http.Client that automatically injects Bearer tokens
// and refreshes expired tokens, persisting refreshed tokens to disk.
func NewHTTPClient(cfg *oauth2.Config, baseTransport http.RoundTripper, userAgent string) *http.Client {
	td, err := LoadTokens()
	if err != nil || !td.HasValidTokens() {
		return &http.Client{Transport: baseTransport}
	}

	src := cfg.TokenSource(context.TODO(), td.ToOAuth2Token())
	persistSrc := &persistingTokenSource{src: src}

	transport := &oauth2.Transport{
		Source: persistSrc,
		Base:   &userAgentTransport{base: baseTransport, userAgent: userAgent},
	}

	return &http.Client{Transport: transport}
}

// persistingTokenSource wraps a TokenSource and saves refreshed tokens to disk.
type persistingTokenSource struct {
	src oauth2.TokenSource
}

func (p *persistingTokenSource) Token() (*oauth2.Token, error) {
	tok, err := p.src.Token()
	if err != nil {
		return nil, err
	}
	// Best-effort save of refreshed tokens
	_ = SaveOAuth2Token(tok)
	return tok, nil
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
