package main

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
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
	if err := srv.Run(ctx); err != nil && ctx.Err() == nil && !clientGone(err) {
		log.Fatalf("trident-mcp: %v", err)
	}
}

// clientGone reports the normal end of a stdio session: the client closed
// stdin, which the SDK surfaces as JSON-RPC "server is closing" (-32004).
func clientGone(err error) bool {
	var rpcErr *jsonrpc.Error
	return errors.Is(err, io.EOF) || (errors.As(err, &rpcErr) && rpcErr.Code == -32004)
}
