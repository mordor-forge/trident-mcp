package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func clearEnv(t *testing.T) {
	for _, key := range []string{"TRIDENT_TRIPO_API_KEY", "TRIPO_API_KEY", "TRIPO_BASE_URL", "MODEL_OUTPUT_DIR"} {
		t.Setenv(key, "")
	}
}

func TestLoad_ReadsEnvironment(t *testing.T) {
	clearEnv(t)
	dir := t.TempDir()
	t.Setenv("TRIPO_API_KEY", "tsk_env")
	t.Setenv("TRIPO_BASE_URL", "https://openapi.tripo3d.com/v3")
	t.Setenv("MODEL_OUTPUT_DIR", dir)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIKey != "tsk_env" || cfg.BaseURL != "https://openapi.tripo3d.com/v3" || cfg.OutputDir != dir {
		t.Fatalf("cfg = %+v", cfg)
	}
}

func TestLoad_PluginKeyWins(t *testing.T) {
	clearEnv(t)
	t.Setenv("TRIPO_API_KEY", "tsk_env")
	t.Setenv("TRIDENT_TRIPO_API_KEY", "tsk_plugin")
	t.Setenv("MODEL_OUTPUT_DIR", t.TempDir())

	cfg, err := Load()
	if err != nil || cfg.APIKey != "tsk_plugin" {
		t.Fatalf("cfg = %+v, err = %v", cfg, err)
	}
}

func TestLoad_RequiresAKey(t *testing.T) {
	clearEnv(t)
	t.Setenv("MODEL_OUTPUT_DIR", t.TempDir())
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "TRIPO_API_KEY") {
		t.Fatalf("err = %v", err)
	}
}

func TestLoad_OutputDirDefaultsAndExpands(t *testing.T) {
	clearEnv(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("TRIPO_API_KEY", "tsk_env")

	cfg, err := Load()
	if err != nil || cfg.OutputDir != filepath.Join(home, "generated_models") {
		t.Fatalf("default: cfg = %+v, err = %v", cfg, err)
	}

	t.Setenv("MODEL_OUTPUT_DIR", "~/assets/3d")
	cfg, err = Load()
	if err != nil || cfg.OutputDir != filepath.Join(home, "assets", "3d") {
		t.Fatalf("tilde: cfg = %+v, err = %v", cfg, err)
	}
	if info, err := os.Stat(cfg.OutputDir); err != nil || !info.IsDir() {
		t.Fatalf("output dir was not created: %v", err)
	}
}
