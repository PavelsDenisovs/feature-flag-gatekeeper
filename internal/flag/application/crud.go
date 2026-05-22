package application

import (
	"context"

	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/flag/domain"
	"github.com/google/uuid"
)

type CreateFlagRequest struct {
	Key         string
	Enabled     bool
	Description string
	Config      domain.Config
}

func (s *flagService) CreateFlag(ctx context.Context, params domain.FlagData) (domain.Flag, error) {
	return domain.Flag{}, nil
}

func (s *flagService) GetFlag(ctx context.Context, id uuid.UUID) (domain.Flag, error) {
	return domain.Flag{}, nil
}

func (s *flagService) UpdateFlag(ctx context.Context, id uuid.UUID, params domain.FlagData) error {
	return nil
}

func (s *flagService) DeleteFlag(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (s *flagService) ListFlags(ctx context.Context, limit, offset int) ([]domain.Flag, error) {
	return []domain.Flag{}, nil
}
