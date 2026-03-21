//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/database/migrator"
	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/flag/domain"
	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/testinfra"
	"github.com/aws/smithy-go/ptr"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestFetchFlagByKey(t *testing.T) {
	ctx := t.Context()
	require := require.New(t)

	connStr := testinfra.SetupPostgres(ctx, t)

	db, err := sql.Open("postgres", connStr)
	require.NoError(err)

	err = migrator.ApplyMigrations(connStr)
	require.NoError(err)

	err = seedDatabase(ctx, t, db)
	require.NoError(err)

	t.Run("existing_key", func(t *testing.T) {
		repo := New(db)
		f, err := repo.FetchFlagByKey(ctx, "available_key")
		require.NoError(err)
		require.NotEqual(f.Key, "")
	})

	t.Run("non_existing_key", func(t *testing.T) {
		repo := New(db)
		f, err := repo.FetchFlagByKey(ctx, "unavailable_key")
		require.ErrorIs(err, domain.ErrFlagNotFound)
		require.Equal(f.Key, "")
	})
}

func seedDatabase(ctx context.Context, t *testing.T, db *sql.DB) error {
	t.Helper()

	query := `
		INSERT INTO flags (id, flag_key, config, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	config := domain.Config{
		Default: ptr.Bool(false),
		Rules: []domain.Rule{
			{
				Action: domain.Action{
					Rollout: ptr.Int(100),
				},
			},
		},
	}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	now := time.Now()

	id := uuid.New()

	_, err = db.ExecContext(ctx, query, id, "available_key", configBytes, true, now, now)
	if err != nil {
		return fmt.Errorf("failed to insert seed data: %w", err)
	}

	return nil
}
