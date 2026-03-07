package domain

type EvaluationContext struct {
	// Any identificator to be used for bucket creation (e.g. UserID the most common, OrganizationID, SessionID)
	RolloutKey string
	FlagKey    string
}

func Evaluate(config Config, eval EvaluationContext) (bool, error) {
	for _, rule := range config.Rules {
		result, match, err := rule.Evaluate(eval)
		if err != nil {
			return false, err
		}
		if match {
			return result, nil
		}
	}
	return config.Default, nil
}

