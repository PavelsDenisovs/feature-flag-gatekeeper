package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/flag/domain"
)

var (
	ErrNoFlagKey = errors.New("FlagKey is absent")
)

type EvaluateRequest struct {
	SubjectKey string
	FlagKey    string
	Attributes map[string]string
}

type EvaluateResponse struct {
	Enabled bool
}

// Evaluate determines whether a flag is enabled for the given evaluation context.
//
// If err != nil, res.Enabled is false
func (s *flagService) Evaluate(ctx context.Context, req EvaluateRequest) (res EvaluateResponse, err error) {
	if req.FlagKey == "" {
		return EvaluateResponse{
			Enabled: false,
		}, ErrNoFlagKey
	}

	flag, err := s.repo.FetchFlagByKey(ctx, req.FlagKey)
	if err != nil {
		return EvaluateResponse{
			Enabled: false,
		}, fmt.Errorf("%w: %v", ErrFlagNotFound, err)
	}

	enabled, err := flag.Evaluate(domain.EvaluationContext{
		SubjectKey: req.SubjectKey,
		FlagKey:    req.FlagKey,
		Attributes: req.Attributes,
	})
	if err != nil {
		return EvaluateResponse{
			Enabled: false,
		}, fmt.Errorf("%w: %v", domain.ErrEvaluationFailed, err)
	}

	return EvaluateResponse{
		Enabled: enabled,
	}, nil
}
