package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/planitaicojp/houjin-cli/internal/config"
)

func TestEnvOr_set(t *testing.T) {
	t.Setenv("HOUJIN_APP_ID", "test-id")
	if v := config.EnvOr("HOUJIN_APP_ID", "fallback"); v != "test-id" {
		t.Errorf("expected test-id, got %s", v)
	}
}

func TestEnvOr_fallback(t *testing.T) {
	if v := config.EnvOr("HOUJIN_NONEXISTENT", "fallback"); v != "fallback" {
		t.Errorf("expected fallback, got %s", v)
	}
}

func TestLoad_noFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOUJIN_CONFIG_DIR", dir)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Format != "json" {
		t.Errorf("expected default format json, got %s", cfg.Format)
	}
}

func TestLoad_withFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOUJIN_CONFIG_DIR", dir)

	yamlContent := []byte("app_id: my-app-id\nformat: table\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), yamlContent, 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AppID != "my-app-id" {
		t.Errorf("expected my-app-id, got %s", cfg.AppID)
	}
	if cfg.Format != "table" {
		t.Errorf("expected table, got %s", cfg.Format)
	}
}

func TestGetAppID_envOverridesFile(t *testing.T) {
	t.Setenv("HOUJIN_APP_ID", "env-id")

	cfg := &config.Config{AppID: "file-id"}
	appID := config.GetAppID(cfg)
	if appID != "env-id" {
		t.Errorf("expected env-id, got %s", appID)
	}
}

func TestGetAppID_fromFile(t *testing.T) {
	os.Unsetenv("HOUJIN_APP_ID")

	cfg := &config.Config{AppID: "file-id"}
	appID := config.GetAppID(cfg)
	if appID != "file-id" {
		t.Errorf("expected file-id, got %s", appID)
	}
}
