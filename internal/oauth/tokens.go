package oauth

import (
	"time"

	"golang.org/x/oauth2"

	"github.com/rootlyhq/rootly-tui/internal/config"
)

// TokenData represents OAuth2 tokens (used as an intermediary).
type TokenData struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresAt    time.Time
}

// LoadTokens loads OAuth2 tokens from the config file.
func LoadTokens() (*TokenData, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return TokenDataFromConfig(cfg), nil
}

// TokenDataFromConfig extracts OAuth2 tokens from an already-loaded config.
func TokenDataFromConfig(cfg *config.Config) *TokenData {
	return &TokenData{
		AccessToken:  cfg.OAuthAccessToken,
		RefreshToken: cfg.OAuthRefreshToken,
		TokenType:    cfg.OAuthTokenType,
		ExpiresAt:    cfg.OAuthExpiresAt,
	}
}

// SaveTokens writes OAuth2 tokens to the config file.
func SaveTokens(td *TokenData) error {
	cfg, err := config.Load()
	if err != nil {
		// If config doesn't exist yet, start with defaults
		cfg = &config.Config{}
	}
	cfg.OAuthAccessToken = td.AccessToken
	cfg.OAuthRefreshToken = td.RefreshToken
	cfg.OAuthTokenType = td.TokenType
	cfg.OAuthExpiresAt = td.ExpiresAt
	cfg.UseOAuth = true
	return config.Save(cfg)
}

// SaveOAuth2Token converts an oauth2.Token and saves it to config.
func SaveOAuth2Token(tok *oauth2.Token) error {
	td := TokenDataFromOAuth2(tok)
	return SaveTokens(td)
}

// ClearTokens removes OAuth tokens from the config file.
func ClearTokens() error {
	cfg, err := config.Load()
	if err != nil {
		return nil // Nothing to clear
	}
	cfg.ClearOAuthTokens()
	return config.Save(cfg)
}

// IsExpired returns true if the token is expired or will expire within 30 seconds.
func (td *TokenData) IsExpired() bool {
	return td.ExpiresAt.Before(time.Now().Add(30 * time.Second))
}

// HasValidTokens returns true if tokens exist and the access token is not empty.
func (td *TokenData) HasValidTokens() bool {
	return td.AccessToken != "" && td.RefreshToken != ""
}

// ToOAuth2Token converts TokenData to an oauth2.Token.
func (td *TokenData) ToOAuth2Token() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  td.AccessToken,
		RefreshToken: td.RefreshToken,
		TokenType:    td.TokenType,
		Expiry:       td.ExpiresAt,
	}
}

// TokenDataFromOAuth2 converts an oauth2.Token to TokenData.
func TokenDataFromOAuth2(tok *oauth2.Token) *TokenData {
	return &TokenData{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		TokenType:    tok.TokenType,
		ExpiresAt:    tok.Expiry,
	}
}
