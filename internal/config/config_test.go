package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_TripoAPIKey(t *testing.T) {
	t.Setenv("TRIPO_API_KEY", "tsk_test-key")
	t.Setenv("MODEL_OUTPUT_DIR", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Provider.APIKey != "tsk_test-key" {
		t.Errorf("APIKey = %q, want %q", cfg.Provider.APIKey, "tsk_test-key")
	}
	if cfg.Backend() != BackendTripo {
		t.Errorf("Backend() = %q, want %q", cfg.Backend(), BackendTripo)
	}
}

func TestLoad_NoCredentials(t *testing.T) {
	t.Setenv("TRIPO_API_KEY", "")
	t.Setenv("MODEL_OUTPUT_DIR", t.TempDir())

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing credentials, got nil")
	}
	if !strings.Contains(err.Error(), "TRIPO_API_KEY") {
		t.Fatalf("error %q does not mention TRIPO_API_KEY", err)
	}
}

func TestConfigBackend_NilConfig(t *testing.T) {
	var cfg *Config
	if cfg.Backend() != BackendUnknown {
		t.Fatalf("Backend() = %q, want %q", cfg.Backend(), BackendUnknown)
	}
}

func TestLoad_DefaultOutputDir(t *testing.T) {
	t.Setenv("TRIPO_API_KEY", "tsk_test-key")
	t.Setenv("MODEL_OUTPUT_DIR", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	expected := filepath.Join(homeDir, "generated_models")
	if cfg.OutputDir != expected {
		t.Errorf("OutputDir = %q, want %q", cfg.OutputDir, expected)
	}
}

func TestLoad_CustomOutputDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TRIPO_API_KEY", "tsk_test-key")
	t.Setenv("MODEL_OUTPUT_DIR", dir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OutputDir != dir {
		t.Errorf("OutputDir = %q, want %q", cfg.OutputDir, dir)
	}
}

func TestLoad_CreatesOutputDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "output")
	t.Setenv("TRIPO_API_KEY", "tsk_test-key")
	t.Setenv("MODEL_OUTPUT_DIR", dir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OutputDir != dir {
		t.Errorf("OutputDir = %q, want %q", cfg.OutputDir, dir)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("output dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("output path is not a directory")
	}
}
