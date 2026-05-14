package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEvaluationFailed     = errors.New("evaluation failed")
	ErrInvalidConfigVersion = errors.New("invalid config version")
)

type Flag struct {
	ID          uuid.UUID
	Key         string
	Enabled     bool
	Description string
	Config      Config
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type EvaluationContext struct {
	// Any identifier to be used for bucket creation (e.g. UserID the most common, OrganizationID, SessionID)
	SubjectKey string
	FlagKey    string
	Attributes map[string]string
}

type FlagData struct {
	Key         string
	Enabled     bool
	Description string
	Config      Config
}

func NewFlag(params FlagData) (*Flag, error) {
	if params.Config.Version != CurrentConfigVersion {
		return nil, fmt.Errorf("%w: current %d, got %d",
			ErrInvalidConfigVersion, CurrentConfigVersion, params.Config.Version)
	}

	return &Flag{
		ID:          uuid.New(),
		Key:         params.Key,
		Enabled:     params.Enabled,
		Description: params.Description,
		Config:      params.Config,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

func (f *Flag) Evaluate(eval EvaluationContext) (bool, error) {
	if !f.Enabled {
		return f.Config.Default, nil
	}

	// After MVP: If f.Config.Version < CurrentConfigVersion,
	// call UpgradeConfigToLatest locally before evaluating.

	for _, r := range f.Config.Rules {
		result, match, err := r.Evaluate(eval)
		if err != nil {
			return false, fmt.Errorf("%w: %v", ErrEvaluationFailed, err)
		}
		if match {
			return result, nil
		}
	}

	return f.Config.Default, nil
}

func (f *Flag) Update(params FlagData) error {
	f.Key = params.Key
	f.Enabled = params.Enabled
	f.Description = params.Description
	f.Config = params.Config
	f.UpdatedAt = time.Now()
	return nil
}
