package cli

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ripmav/erdbeerpfoetchen/internal/database"
)

type MigrateCommand struct{}

func (cmd *MigrateCommand) Run(ctx context.Context, cfg *Config) error {
	db, err := database.Connect(ctx, cfg.DB.URI, cfg.EnableDebug)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.ErrorContext(ctx, "failed to close database connection", "error", err)
		}
	}()

	if err := db.Migrate(ctx); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	slog.InfoContext(ctx, "migrations applied successfully")
	return nil
}
