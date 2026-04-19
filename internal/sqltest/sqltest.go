package sqltest

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const image = "postgres@sha256:52098013b4b64a746626437d38afc03cabff6cdeb4d3d92e2342aa95f0ce56ea"

type Container struct {
	URI string

	close    sync.Once
	postgres *postgres.PostgresContainer
}

func CreateContainer(ctx context.Context) (*Container, error) {
	provider := testcontainers.ProviderDefault

	if os.Getenv("TESTCONTAINERS_PROVIDER_PODMAN") != "" {
		provider = testcontainers.ProviderPodman
	}

	pg, err := postgres.Run(ctx, image,
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.BasicWaitStrategies(),
		testcontainers.WithProvider(provider),
	)

	if err != nil {
		return nil, fmt.Errorf("run container: %w", err)
	}

	uri, err := pg.ConnectionString(ctx)

	if err != nil {
		return nil, fmt.Errorf("construct connection string: %w", err)
	}

	uri = strings.ReplaceAll(uri, "localhost", "127.0.0.1")

	return &Container{URI: uri, postgres: pg}, nil
}

func (c *Container) Destroy(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var err error

	c.close.Do(func() {
		err = c.postgres.Terminate(ctx)
	})

	if err != nil {
		return fmt.Errorf("terminate container: %w", err)
	}

	return nil
}
