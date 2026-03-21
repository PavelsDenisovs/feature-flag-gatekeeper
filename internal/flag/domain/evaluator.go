package domain

import (
	"errors"
	"fmt"
)

var ErrFlagNotFound = errors.New("flag not found")

type EvaluationContext struct {
	// Any identificator to be used for bucket creation (e.g. UserID the most common, OrganizationID, SessionID)
	RolloutKey string
	FlagKey    string
}

func Evaluate(enabled bool, config Config, eval EvaluationContext) (bool, error) {
	if !enabled {
		return false, nil
	}
	if config.Default == nil {
		return false, fmt.Errorf("config.Default is missing")
	}
	for _, rule := range config.Rules {
		result, match, err := rule.Evaluate(eval)
		if err != nil {
			return false, err
		}
		if match {
			return result, nil
		}
	}
	return *config.Default, nil
}
