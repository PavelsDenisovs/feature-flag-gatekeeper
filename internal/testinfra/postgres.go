//go:build integration

package testinfra

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	ErrContainerStartFailed = errors.New("failed to start container")
	ErrGetConnectionString  = errors.New("failed to get pq connection string from container")
)

type PostgresContainer struct {
	ConnectionString string
	Cleanup          func()
}

func SetupPostgres(ctx context.Context) (*PostgresContainer, error) {
	pqContainer, err := postgres.Run(
		ctx, "postgres:16-alpine",
		postgres.WithDatabase("test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrContainerStartFailed, err)
	}

	cleanup := func() {
		if err := testcontainers.TerminateContainer(pqContainer.Container); err != nil {
			log.Printf("Failed to terminate container: %v", err)
		}
	}

	connStr, err := pqContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("%w: %v", ErrGetConnectionString, err)
	}

	return &PostgresContainer{
		ConnectionString: connStr,
		Cleanup:          cleanup,
	}, nil
}
