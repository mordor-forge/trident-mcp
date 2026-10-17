// Package config loads trident-mcp's settings from the environment.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds everything the server needs to start.
type Config struct {
	// APIKey comes from TRIDENT_TRIPO_API_KEY (set by the Claude Code
	// plugin), falling back to TRIPO_API_KEY.
	APIKey string
	// BaseURL is TRIPO_BASE_URL; empty means the global Tripo host.
	BaseURL string
	// OutputDir is MODEL_OUTPUT_DIR, default ~/generated_models. It is created
	// if missing.
	OutputDir string
}

// Load reads the configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		APIKey:    firstEnv("TRIDENT_TRIPO_API_KEY", "TRIPO_API_KEY"),
		BaseURL:   strings.TrimSpace(os.Getenv("TRIPO_BASE_URL")),
		OutputDir: expandHome(firstEnv("MODEL_OUTPUT_DIR")),
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("no Tripo API key: set TRIPO_API_KEY (create one at https://platform.tripo3d.ai)")
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = defaultOutputDir()
	}
	abs, err := filepath.Abs(cfg.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("resolving output directory %s: %w", cfg.OutputDir, err)
	}
	cfg.OutputDir = abs
	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating output directory %s: %w", cfg.OutputDir, err)
	}
	return cfg, nil
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return ""
}

func expandHome(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~"))
		}
	}
	return path
}

func defaultOutputDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "generated_models"
	}
	return filepath.Join(home, "generated_models")
}
