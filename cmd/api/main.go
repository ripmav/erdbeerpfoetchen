package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"

	"github.com/ripmav/erdbeerpfoetchen/internal/cli"
)

type application struct {
	cli.Config `envprefix:"PFOETCHEN_"`

	Api cli.ApiCommand `cmd:"" help:"Run the API server"`
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

	var app application

	options := []kong.Option{
		kong.Name("Pfötchen API"),
		kong.UsageOnError(),
	}

	k := kong.Parse(&app, options...)
	k.BindTo(ctx, (*context.Context)(nil))

	if err := k.Run(&app.Config); err != nil {
		return fmt.Errorf("exit: %w", err)
	}

	return nil
}
