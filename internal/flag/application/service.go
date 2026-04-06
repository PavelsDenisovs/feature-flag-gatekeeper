package application

import (
	"context"

	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/flag/domain"
)

//mockery:generate: true
type Repository interface {
	FetchFlagByKey(ctx context.Context, flagKey string) (domain.Flag, error)
}

//mockery:generate: true
type Service interface {
	Evaluate(ctx context.Context, req EvaluateRequest) (res EvaluateResponse, err error)
}

type service struct {
	repo Repository
}

func New(repo Repository) *service {
	if repo == nil {
		panic("nil repository")
	}
	return &service{repo: repo}
}
