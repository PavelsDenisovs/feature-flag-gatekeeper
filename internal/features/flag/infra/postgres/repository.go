package postgres

import (
	"context"
	"database/sql"

	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/features/flag/domain"
)

type repository struct {
	db *sql.DB
}

func New(db *sql.DB) *repository {
	return &repository{db: db}
}

func (r repository) FetchFlagByKey(ctx context.Context, key string) (domain.Flag, error) {
	return domain.Flag{}, nil
}