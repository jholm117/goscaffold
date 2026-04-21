package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_FromFile(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(cfgFile, []byte("module_prefix: github.com/testuser\nhomebrew_tap_token: tok123\n"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadFrom(cfgFile)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if cfg.ModulePrefix != "github.com/testuser" {
		t.Errorf("ModulePrefix = %q, want %q", cfg.ModulePrefix, "github.com/testuser")
	}
	if cfg.HomebrewTapToken != "tok123" {
		t.Errorf("HomebrewTapToken = %q, want %q", cfg.HomebrewTapToken, "tok123")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	cfg, err := LoadFrom("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("LoadFrom should not error on missing file: %v", err)
	}
	if cfg.ModulePrefix != "" {
		t.Errorf("ModulePrefix = %q, want empty", cfg.ModulePrefix)
	}
}

func TestLoad_DefaultPath(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	_ = cfg
}
