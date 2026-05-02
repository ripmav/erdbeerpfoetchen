package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
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

// prescanConfigFlag discovers the config file path before the full kong.Parse
// so the YAML resolver can be registered in time. It parses os.Args using the
// full application struct (so no flags are unknown), suppressing all output and
// exit calls since errors are irrelevant at this stage.
func prescanConfigFlag() string {
	var app application
	p, err := kong.New(&app,
		kong.Exit(func(int) {}),
		kong.Writers(io.Discard, io.Discard),
	)
	if err != nil {
		return ""
	}
	_, _ = p.Parse(os.Args[1:])
	return app.ConfigFile
}
