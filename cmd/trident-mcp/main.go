package main

import (
	"context"
	"log"

	"github.com/mordor-forge/trident-mcp/internal/config"
	"github.com/mordor-forge/trident-mcp/internal/provider/tripo"
	"github.com/mordor-forge/trident-mcp/internal/server"
)

var version = "dev"

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	p, err := tripo.NewFromConfig(cfg)
	if err != nil {
		log.Fatalf("creating provider: %v", err)
	}

	// TripoProvider implements all four interfaces.
	srv := server.NewWithOptions(p, p, p, p, server.Options{
		Backend:   cfg.Backend(),
		OutputDir: cfg.OutputDir,
		Version:   version,
	})

	if err := srv.Run(ctx); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
