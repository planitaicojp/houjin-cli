package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DefaultFormat = "json"
	configFile    = "config.yaml"
)

// Config represents the application configuration.
type Config struct {
	AppID  string `yaml:"app_id"`
	Format string `yaml:"format"`
}

// DefaultConfigDir returns the config directory path.
func DefaultConfigDir() string {
	if d := os.Getenv(EnvConfigDir); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "houjin")
}

// Load reads the config file. Returns default config if file doesn't exist.
func Load() (*Config, error) {
	dir := DefaultConfigDir()
	path := filepath.Join(dir, configFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultConfig(), nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.Format == "" {
		cfg.Format = DefaultFormat
	}
	return &cfg, nil
}

// GetAppID returns the app ID with env override.
func GetAppID(cfg *Config) string {
	if v := os.Getenv(EnvAppID); v != "" {
		return v
	}
	return cfg.AppID
}

func defaultConfig() *Config {
	return &Config{
		Format: DefaultFormat,
	}
}
