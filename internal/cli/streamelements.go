package cli

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ripmav/erdbeerpfoetchen/internal/handle"
)

type ApiCommand struct {
	Now time.Time `name:"now"`
}

func (cmd *ApiCommand) Run(ctx context.Context, cfg *Config) error {
	mux := http.NewServeMux()

	handle.NewLamimiHandler(nil)

	if err := cfg.ListenAndServe(ctx, mux); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}
