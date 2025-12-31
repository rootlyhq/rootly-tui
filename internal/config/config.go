package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DefaultEndpoint = "api.rootly.com"
	configDir       = ".rootly-tui"
	configFile      = "config.yaml"
)

type Config struct {
	APIKey   string `yaml:"api_key"`
	Endpoint string `yaml:"endpoint"`
}

func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, configDir)
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), configFile)
}

func Exists() bool {
	_, err := os.Stat(ConfigPath())
	return err == nil
}

func Load() (*Config, error) {
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("config file not found")
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Endpoint == "" {
		cfg.Endpoint = DefaultEndpoint
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	if cfg.Endpoint == "" {
		cfg.Endpoint = DefaultEndpoint
	}

	if err := os.MkdirAll(ConfigDir(), 0700); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(ConfigPath(), data, 0600)
}

func (c *Config) IsValid() bool {
	return c.APIKey != "" && c.Endpoint != ""
}
