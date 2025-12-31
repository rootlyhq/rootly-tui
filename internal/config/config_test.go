package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultEndpoint(t *testing.T) {
	if DefaultEndpoint != "api.rootly.com" {
		t.Errorf("expected default endpoint to be 'api.rootly.com', got '%s'", DefaultEndpoint)
	}
}

func TestPath(t *testing.T) {
	path := Path()
	if path == "" {
		t.Error("expected config path to be non-empty")
	}

	// Should end with config.yaml
	if filepath.Base(path) != "config.yaml" {
		t.Errorf("expected config path to end with 'config.yaml', got '%s'", filepath.Base(path))
	}

	// Should contain .rootly-tui directory
	dir := filepath.Dir(path)
	if filepath.Base(dir) != ".rootly-tui" {
		t.Errorf("expected config dir to be '.rootly-tui', got '%s'", filepath.Base(dir))
	}
}

func TestDir(t *testing.T) {
	dir := Dir()
	if dir == "" {
		t.Error("expected config dir to be non-empty")
	}

	if filepath.Base(dir) != ".rootly-tui" {
		t.Errorf("expected config dir to end with '.rootly-tui', got '%s'", filepath.Base(dir))
	}
}

func TestConfigIsValid(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected bool
	}{
		{
			name:     "valid config",
			config:   Config{APIKey: "test-key", Endpoint: "api.rootly.com"},
			expected: true,
		},
		{
			name:     "missing api key",
			config:   Config{APIKey: "", Endpoint: "api.rootly.com"},
			expected: false,
		},
		{
			name:     "missing endpoint",
			config:   Config{APIKey: "test-key", Endpoint: ""},
			expected: false,
		},
		{
			name:     "empty config",
			config:   Config{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.IsValid(); got != tt.expected {
				t.Errorf("Config.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "rootly-tui-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Test saving config
	cfg := &Config{
		APIKey:   "test-api-key-12345",
		Endpoint: "api.test.rootly.com",
	}

	err = Save(cfg)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify file was created
	if !Exists() {
		t.Error("expected config file to exist after save")
	}

	// Test loading config
	loaded, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded.APIKey != cfg.APIKey {
		t.Errorf("expected APIKey '%s', got '%s'", cfg.APIKey, loaded.APIKey)
	}

	if loaded.Endpoint != cfg.Endpoint {
		t.Errorf("expected Endpoint '%s', got '%s'", cfg.Endpoint, loaded.Endpoint)
	}
}

func TestLoadDefaultEndpoint(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "rootly-tui-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Save config without endpoint
	cfg := &Config{
		APIKey:   "test-key",
		Endpoint: "", // Empty endpoint
	}

	err = Save(cfg)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Load and verify default endpoint is set
	loaded, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded.Endpoint != DefaultEndpoint {
		t.Errorf("expected default endpoint '%s', got '%s'", DefaultEndpoint, loaded.Endpoint)
	}
}

func TestLoadNonExistent(t *testing.T) {
	// Create temp directory for test (empty, no config)
	tmpDir, err := os.MkdirTemp("", "rootly-tui-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Try to load non-existent config
	_, err = Load()
	if err == nil {
		t.Error("expected error when loading non-existent config")
	}
}

func TestExistsWhenNotExists(t *testing.T) {
	// Create temp directory for test (empty, no config)
	tmpDir, err := os.MkdirTemp("", "rootly-tui-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	if Exists() {
		t.Error("expected Exists() to return false for non-existent config")
	}
}
