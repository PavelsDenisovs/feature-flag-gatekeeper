package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/features/flag/domain"
	"github.com/google/uuid"
)

type repository struct {
	db *sql.DB
}

func New(db *sql.DB) *repository {
	return &repository{db: db}
}

type flagRow struct {
	ID          uuid.UUID       `db:"id"`
	FlagKey     string          `db:"flag_key"`
	Enabled     bool            `db:"enabled"`
	Description string          `db:"description"`
	Config      json.RawMessage `db:"config"`
	CreatedAt   time.Time       `db:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at"`
}

func (row *flagRow) fromDomain(f domain.Flag) error {
	configBytes, err := json.Marshal(f.Config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	row.ID = f.ID
	row.FlagKey = f.Key
	row.Enabled = f.Enabled
	row.Description = f.Description
	row.Config = configBytes
	row.CreatedAt = f.CreatedAt
	row.UpdatedAt = f.UpdatedAt

	return nil
}

func (row *flagRow) toDomain() (domain.Flag, error) {
	var f domain.Flag
	if err := json.Unmarshal(row.Config, &f.Config); err != nil {
		return domain.Flag{}, fmt.Errorf("unmarshal config: %w", err)
	}

	f.ID = row.ID
	f.Key = row.FlagKey
	f.Enabled = row.Enabled
	f.Description = row.Description
	f.CreatedAt = row.CreatedAt
	f.UpdatedAt = row.UpdatedAt

	return f, nil
}

func (r *repository) FetchFlagByKey(ctx context.Context, key string) (domain.Flag, error) {
	query := `
		SELECT id, flag_key, enabled, description, config, created_at, updated_at 
    FROM flags WHERE flag_key = $1`

	var row flagRow
	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&row.ID, &row.FlagKey, &row.Enabled, &row.Description, &row.Config, &row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Flag{}, domain.ErrFlagNotFound
		}
		return domain.Flag{}, err
	}
	return row.toDomain()
}