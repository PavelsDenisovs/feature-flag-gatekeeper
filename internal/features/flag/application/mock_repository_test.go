package application

import (
	"context"
	"fmt"

	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/features/flag/domain"
)

type mockFlagRepository struct{
	flags    []domain.Flag
	errFetch error
}

func (r *mockFlagRepository) FetchFlagByKey(ctx context.Context, flagKey string) (domain.Flag, error) {
	if r.errFetch != nil {
		return domain.Flag{}, r.errFetch
	}
	for _, f := range r.flags {
		if f.Key == flagKey {
			return f, nil
		}
	}
	return domain.Flag{}, fmt.Errorf("Flag with key %s is not found", flagKey)
}