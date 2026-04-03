package config

import "os"

const (
	EnvAppID     = "HOUJIN_APP_ID"
	EnvFormat    = "HOUJIN_FORMAT"
	EnvConfigDir = "HOUJIN_CONFIG_DIR"
)

// EnvOr returns the environment variable value if set, otherwise the fallback.
func EnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
