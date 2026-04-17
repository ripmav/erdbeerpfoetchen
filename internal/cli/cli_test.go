package cli_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ripmav/erdbeerpfoetchen/internal/cli"
)

func TestConfig_SetDefaultLogger(t *testing.T) {
	cfg := &cli.Config{}
	cfg.SetDefaultLogger() // info level — must not panic

	cfg.EnableDebug = true
	cfg.SetDefaultLogger() // debug level — must not panic
}

func TestConfig_PrintBuildInfo(t *testing.T) {
	cfg := &cli.Config{}
	cfg.SetDefaultLogger()
	cfg.PrintBuildInfo() // must not panic; VCS fields may be empty in test binaries
}

func TestConfig_ListenAndServe_ContextCancellation(t *testing.T) {
	cfg := &cli.Config{}
	cfg.Server.Listen = "127.0.0.1:0"
	cfg.Server.EnableUnsafe = true

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := cfg.ListenAndServe(ctx, http.NewServeMux())
	// Graceful shutdown after context timeout must return nil.
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestConfig_ListenAndServe_InvalidAddress(t *testing.T) {
	cfg := &cli.Config{}
	cfg.Server.Listen = "invalid-address"
	cfg.Server.EnableUnsafe = true

	err := cfg.ListenAndServe(context.Background(), http.NewServeMux())
	if err == nil {
		t.Error("expected error for invalid listen address, got nil")
	}
}
