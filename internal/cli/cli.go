package cli

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"time"
)

type Config struct {
	EnableDebug bool `name:"debug" env:"DEBUG" default:"false" help:"Enable debug mode"`

	Server struct {
		Listen       string `name:"listen" env:"LISTEN" default:":8080" help:"Listen address"`
		EnableUnsafe bool   `name:"unsafe" env:"UNSAFE" default:"false" help:"Allow insecure connections"`

		TLS struct {
			CertPath string `name:"cert" env:"CERT_PATH" help:"Path to TLS certificate file"`
			KeyPath  string `name:"key" env:"KEY_PATH" help:"Path to TLS key file"`
		} `embed:"" prefix:"tls." envprefix:"TLS_"`
	} `embed:"" prefix:"server." envprefix:"SERVER_"`

	DB struct {
		URI string `name:"uri" env:"URI" help:"Database URI"`
	} `embed:"" prefix:"database." envprefix:"DATABASE_"`
}

func (cfg *Config) SetDefaultLogger() {
	level := slog.LevelInfo

	if cfg.EnableDebug {
		level = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))
}

func (cfg *Config) PrintBuildInfo() {
	info, _ := debug.ReadBuildInfo()

	var (
		revision  string // Git revision
		timestamp string // Git revision timestamp
		modified  string // dirty local source tree
	)

	for _, kv := range info.Settings {
		switch kv.Key {
		case "vcs.revision":
			revision = kv.Value
		case "vcs.time":
			timestamp = kv.Value
		case "vcs.modified":
			modified = kv.Value
		}
	}

	slog.Info(
		"Build Info",
		"git.revision", revision,
		"git.timestamp", timestamp,
		"git.modified", modified,
		"go.version", info.GoVersion,
		"module", info.Path,
	)
}

func (cfg *Config) ListenAndServe(ctx context.Context, mux http.Handler) error {
	logger := slog.NewLogLogger(slog.Default().Handler(), slog.LevelInfo)
	srv := &http.Server{
		Addr:              cfg.Server.Listen,
		ErrorLog:          logger,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      25 * time.Second,
		IdleTimeout:       60 * time.Second,

		Handler: mux,
	}

	quit := make(chan error, 1)

	go func() {
		quit <- cfg.listenAndServe(srv)
	}()

	select {
	case <-ctx.Done():
		slog.Info("initiate server shutdown")
		return srv.Shutdown(context.Background())
	case err := <-quit:
		return err
	}
}

func (cfg *Config) listenAndServe(srv *http.Server) error {
	certPath := cfg.Server.TLS.CertPath
	keyPath := cfg.Server.TLS.KeyPath

	if certPath == "" && keyPath == "" {
		if cfg.Server.EnableUnsafe {
			return srv.ListenAndServe()
		}
	}

	return srv.ListenAndServeTLS(certPath, keyPath)
}
