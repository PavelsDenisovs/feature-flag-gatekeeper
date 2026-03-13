//go:build integration

package testinfra

import (
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func SetupPostgres(t *testing.T) (connStr string) {
	t.Helper()

	ctx := t.Context()

	pqContainer, err := postgres.Run(
		ctx, "postgres:16-alpine",
		postgres.WithDatabase("test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("Failed to start container: %w", err)
		return ""
	}

	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(pqContainer.Container); err != nil {
			t.Logf("Failed to terminate container: %w", err)
		}
	})

	connStr, err = pqContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("Failed to get connection string from container: %w", err)
		return ""
	}

	return connStr
}