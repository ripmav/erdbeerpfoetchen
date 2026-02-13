package cli

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ripmav/erdbeerpfoetchen/internal/database"
	"github.com/ripmav/erdbeerpfoetchen/internal/handle"
	"github.com/ripmav/erdbeerpfoetchen/internal/lamimi"
)

type ApiCommand struct {
	Now time.Time `name:"now"`
}

func (cmd *ApiCommand) Run(ctx context.Context, cfg *Config) error {
	mux := http.NewServeMux()

	db, err := database.Connect(ctx, cfg.DB.URI)

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("failed to close database connection", "error", err)
		}
	}()

	lamimiService := lamimi.NewService(database.NewLamimiRepository(db))
	lamimiHandler := handle.NewLamimiHandler(lamimiService)
	lamimiHandler.Install(mux)

	if err := cfg.ListenAndServe(ctx, mux); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}
