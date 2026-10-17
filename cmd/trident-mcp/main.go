package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mordor-forge/trident-mcp/internal/config"
	"github.com/mordor-forge/trident-mcp/internal/server"
	"github.com/mordor-forge/trident-mcp/internal/tripo"
)

// version is set by GoReleaser via -ldflags.
var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("trident-mcp: %v", err)
	}
	client, err := tripo.New(tripo.Options{APIKey: cfg.APIKey, BaseURL: cfg.BaseURL})
	if err != nil {
		log.Fatalf("trident-mcp: %v", err)
	}
	srv := server.New(client, server.Options{OutputDir: cfg.OutputDir, Version: version})
	if err := srv.Run(ctx); err != nil && ctx.Err() == nil {
		log.Fatalf("trident-mcp: %v", err)
	}
}
