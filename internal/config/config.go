package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	BackendUnknown = "unknown"
	BackendTripo   = "tripo"
)

// Config holds all configuration for the MCP server.
type Config struct {
	Provider  ProviderConfig
	OutputDir string
}

// ProviderConfig holds Tripo API authentication.
type ProviderConfig struct {
	APIKey string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		OutputDir: envOr("MODEL_OUTPUT_DIR", defaultOutputDir()),
		Provider: ProviderConfig{
			APIKey: os.Getenv("TRIPO_API_KEY"),
		},
	}

	if cfg.Provider.APIKey == "" {
		return nil, fmt.Errorf(
			"no credentials configured: set TRIPO_API_KEY environment variable",
		)
	}

	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating output directory %s: %w", cfg.OutputDir, err)
	}

	return cfg, nil
}

// Backend returns the effective backend based on configuration.
func (c *Config) Backend() string {
	if c == nil {
		return BackendUnknown
	}
	if c.Provider.APIKey != "" {
		return BackendTripo
	}
	return BackendUnknown
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func defaultOutputDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "generated_models"
	}
	return filepath.Join(home, "generated_models")
}
