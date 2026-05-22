package cli

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ripmav/erdbeerpfoetchen/internal/collection"
	"github.com/ripmav/erdbeerpfoetchen/internal/database"
	"github.com/ripmav/erdbeerpfoetchen/internal/handle"
	"github.com/ripmav/erdbeerpfoetchen/internal/middleware"
	"github.com/ripmav/erdbeerpfoetchen/internal/user"
)

type ApiCommand struct {
	Now time.Time `name:"now"`
}

func (cmd *ApiCommand) Run(ctx context.Context, cfg *Config) error {
	mux := http.NewServeMux()

	db, err := database.ConnectAndMigrate(ctx, cfg.DB.URI, cfg.EnableDebug)

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("failed to close database connection", "error", err)
		}
	}()

	collectionRepo := database.NewCollectionRepository(db, cfg.EnableDebug)
	userRepo := database.NewUserRepository(db, cfg.EnableDebug)

	collectionService := collection.New(collectionRepo)
	userService := user.New(userRepo)

	rateLimiter := middleware.NewRateLimiter(userService)

	collectionHandler := handle.NewCollectionHandler(collectionService, userService, rateLimiter)
	collectionHandler.Install(mux)

	authHandler := handle.NewAuthHandler(
		cfg.Twitch.ClientID,
		cfg.Twitch.ClientSecret,
		cfg.Twitch.RedirectURL,
		cfg.Twitch.AdminIDs,
		userService,
	)
	authHandler.Install(mux)

	if err := cfg.ListenAndServe(ctx, mux); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}
