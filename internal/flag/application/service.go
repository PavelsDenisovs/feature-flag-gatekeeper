package application

import (
	"context"
	"errors"

	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/flag/domain"
	"github.com/google/uuid"
)

var (
	ErrFlagNotFound         = errors.New("flag not found")
)

//mockery:generate: true
type FlagRepository interface {
	CreateFlag(ctx context.Context, f *domain.Flag) error
	UpdateFlag(ctx context.Context, f *domain.Flag) error
	FetchFlag(ctx context.Context, id uuid.UUID) (*domain.Flag, error)
	FetchFlagByKey(ctx context.Context, key string) (*domain.Flag, error)
	DeleteFlag(ctx context.Context, id uuid.UUID) error
	ListFlags(ctx context.Context, limit, offset int) ([]*domain.Flag, error)
}

//mockery:generate: true
type FlagService interface {
	CreateFlag(ctx context.Context, params domain.FlagData) (uuid.UUID, error)
	GetFlag(ctx context.Context, id uuid.UUID) (domain.Flag, error)
	UpdateFlag(ctx context.Context, id uuid.UUID, params domain.FlagData) error
	DeleteFlag(ctx context.Context, id uuid.UUID) error
	ListFlags(ctx context.Context, limit, offset int) ([]domain.Flag, error)
	Evaluate(ctx context.Context, req EvaluateRequest) (res EvaluateResponse, err error)
}

type flagService struct {
	repo FlagRepository
}

func NewFlagService(repo FlagRepository) *flagService {
	if repo == nil {
		panic("nil repository")
	}
	return &flagService{repo: repo}
}
