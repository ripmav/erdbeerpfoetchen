package cli_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ripmav/erdbeerpfoetchen/internal/cli"
	"github.com/ripmav/erdbeerpfoetchen/internal/sqltest"
)

func TestMigrateCommand_Run(t *testing.T) {
	container, err := sqltest.CreateContainer(t.Context())
	require.NoError(t, err)
	defer container.Destroy(t.Context()) //nolint:errcheck

	var cfg cli.Config
	cfg.DB.URI = container.URI
	cfg.EnableDebug = true

	cmd := cli.MigrateCommand{}

	require.NoError(t, cmd.Run(t.Context(), &cfg), "first run must succeed")
	require.NoError(t, cmd.Run(t.Context(), &cfg), "second run must be idempotent")
}
