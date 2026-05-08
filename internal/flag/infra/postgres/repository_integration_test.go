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
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestFetchFlagByKey(t *testing.T) {
	ctx := t.Context()
	require := require.New(t)

	pq, err := testinfra.SetupPostgres(ctx)
	require.NoError(err)
	defer pq.Cleanup()

	db, err := sql.Open("postgres", pq.ConnectionString)
	require.NoError(err)
	defer db.Close()

	err = migrator.ApplyMigrations(pq.ConnectionString)
	require.NoError(err)

	err = seedDatabase(ctx, t, db)
	require.NoError(err)

	t.Run("existing_key", func(t *testing.T) {
		repo := NewFlagRepository(db)
		f, err := repo.FetchFlagByKey(ctx, "available_key")
		require.NoError(err)
		require.NotNil(f)
		require.NotEqual(f.Key, "")
	})

	t.Run("non_existing_key", func(t *testing.T) {
		repo := NewFlagRepository(db)
		f, err := repo.FetchFlagByKey(ctx, "unavailable_key")
		require.ErrorIs(err, domain.ErrFlagNotFound)
		require.Nil(f)
	})
}

func seedDatabase(ctx context.Context, t *testing.T, db *sql.DB) error {
	t.Helper()

	query := `
		INSERT INTO flags (id, flag_key, config, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	config := domain.Config{
		Default: false,
		Rules: []domain.Rule{
			{
				Conditions: []domain.Condition{
					{
						Kind:       domain.ConditionKindRollout,
						Percentage: 100,
					},
				},
				Result: true,
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
