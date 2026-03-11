package application

import (
	"context"

	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/features/flag/domain"
)

type Repository interface {
	FetchFlagByKey(ctx context.Context, flagKey string) (domain.Flag, error)
}

type Service interface {}

type service struct {
	repo Repository
}

func New(repo Repository) *service {
	if repo == nil {
    panic("nil repository")
  }
	return &service{repo: repo}
}
