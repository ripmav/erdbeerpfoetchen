package cli

import (
	"context"
	"log/slog"

	"github.com/ripmav/erdbeerpfoetchen/internal/database"
)

// MigrateCommand connects to the database and applies all pending migrations.
type MigrateCommand struct{}

func (cmd *MigrateCommand) Run(ctx context.Context, cfg *Config) error {
	db, err := database.ConnectAndMigrate(ctx, cfg.DB.URI, cfg.EnableDebug)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.ErrorContext(ctx, "failed to close database connection", "error", err)
		}
	}()

	slog.InfoContext(ctx, "migrations applied successfully")
	return nil
}
