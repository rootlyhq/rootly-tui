package oauth

import (
	"testing"
)

func TestDeriveAuthBaseURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"api.rootly.com", "https://api.rootly.com"},
		{"api.rootly.com/api", "https://api.rootly.com"},
		{"localhost:22166", "http://localhost:22166"},
		{"localhost:22166/api", "http://localhost:22166"},
		{"127.0.0.1:22166", "http://127.0.0.1:22166"},
		{"127.0.0.1:22166/api/v1", "http://127.0.0.1:22166"},
		{"http://localhost:22166", "http://localhost:22166"},
		{"http://localhost:22166/api", "http://localhost:22166"},
		{"https://api.example.com", "https://api.example.com"},
		{"https://api.example.com/v1", "https://api.example.com"},
		{"custom.rootly.io", "https://custom.rootly.io"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := DeriveAuthBaseURL(tt.input)
			if result != tt.expected {
				t.Errorf("DeriveAuthBaseURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateState(t *testing.T) {
	s1, err := GenerateState()
	if err != nil {
		t.Fatalf("GenerateState() error: %v", err)
	}
	if s1 == "" {
		t.Error("GenerateState() returned empty string")
	}

	s2, err := GenerateState()
	if err != nil {
		t.Fatalf("GenerateState() error: %v", err)
	}
	if s1 == s2 {
		t.Error("GenerateState() returned same value twice")
	}
}

func TestNewConfig(t *testing.T) {
	cfg := NewConfig("https://api.rootly.com")
	if cfg.ClientID != ClientID {
		t.Errorf("ClientID = %q, want %q", cfg.ClientID, ClientID)
	}
	if cfg.RedirectURL != RedirectURL {
		t.Errorf("RedirectURL = %q, want %q", cfg.RedirectURL, RedirectURL)
	}
	if cfg.Endpoint.AuthURL != "https://api.rootly.com/oauth/authorize" {
		t.Errorf("AuthURL = %q", cfg.Endpoint.AuthURL)
	}
	if cfg.Endpoint.TokenURL != "https://api.rootly.com/oauth/token" {
		t.Errorf("TokenURL = %q", cfg.Endpoint.TokenURL)
	}
}
