package application

import (
	"context"

	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/flag/domain"
)

//mockery:generate: true
type FlagRepository interface {
	FetchFlagByKey(ctx context.Context, flagKey string) (*domain.Flag, error)
}

//mockery:generate: true
type FlagService interface {
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
