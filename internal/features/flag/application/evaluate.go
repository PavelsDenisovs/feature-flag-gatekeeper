package application

import (
	"context"
	"fmt"

	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/features/flag/domain"
)

type EvaluateRequest struct {
	RolloutKey string
	FlagKey    string
}

type EvaluateResponse struct {
	Enabled bool
}

// Evaluate determines whether a flag is enabled for the given evaluation context.
//
// If err != nil, res.Enabled is false
func (s *service) Evaluate(ctx context.Context, req EvaluateRequest) (res EvaluateResponse, err error) {
	if req.FlagKey == "" {
		return EvaluateResponse{
			Enabled: false,
		}, fmt.Errorf("FlagKey is missing")
	}

	flag, err := s.repo.FetchFlagByKey(ctx, req.FlagKey)
	if err != nil {
		return EvaluateResponse{
			Enabled: false,
		}, fmt.Errorf("failed to fetch flag by key %s: %w", req.FlagKey, err)
	}

	enabled, err := domain.Evaluate(flag.Enabled, flag.Config, domain.EvaluationContext{
		RolloutKey: req.RolloutKey,
		FlagKey:    req.FlagKey,
	})

	if err != nil {
		return EvaluateResponse{
			Enabled: enabled,
		}, fmt.Errorf("evaluation failed: %w", err)
	}
	
	return EvaluateResponse{
		Enabled: enabled,
	}, nil
}