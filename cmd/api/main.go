package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/alecthomas/kong"

	"github.com/ripmav/erdbeerpfoetchen/internal/cli"
	"github.com/ripmav/erdbeerpfoetchen/internal/config"
)

type application struct {
	cli.Config `envprefix:"PFOETCHEN_"`

	ConfigFile string             `name:"config" short:"c" env:"PFOETCHEN_CONFIG" help:"Path to YAML config file" type:"existingfile"`
	Api        cli.ApiCommand     `cmd:"" help:"Run the API server"`
	Migrate    cli.MigrateCommand `cmd:"" help:"Apply database migrations and exit"`
}

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		ch := make(chan os.Signal, 1)

		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(ch)

		select {
		case <-ch:
			cancel()
		case <-ctx.Done():
			return
		}
	}()

	options := []kong.Option{
		kong.Name("Pfötchen API"),
		kong.UsageOnError(),
	}

	if path := prescanConfigFlag(); path != "" {
		resolver, err := config.New(path)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		options = append(options, kong.Resolvers(resolver))
	}

	var app application

	k := kong.Parse(&app, options...)
	k.BindTo(ctx, (*context.Context)(nil))

	if err := k.Run(&app.Config); err != nil {
		return fmt.Errorf("exit: %w", err)
	}

	return nil
}

// prescanConfigFlag returns the config file path from PFOETCHEN_CONFIG env var
// or --config / -c CLI flag, whichever is found first. It must run before
// kong.Parse so the YAML resolver can be registered in time.
func prescanConfigFlag() string {
	if path := os.Getenv("PFOETCHEN_CONFIG"); path != "" {
		return path
	}
	args := os.Args[1:]
	for i, arg := range args {
		switch {
		case arg == "--config" || arg == "-c":
			if i+1 < len(args) {
				return args[i+1]
			}
		case strings.HasPrefix(arg, "--config="):
			return strings.TrimPrefix(arg, "--config=")
		}
	}
	return ""
}
