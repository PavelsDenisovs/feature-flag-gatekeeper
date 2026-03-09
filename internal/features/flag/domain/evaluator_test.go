package domain

import (
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/stretchr/testify/assert"
)

func TestEvaluate(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name           string
		enabled        bool
		config         Config
		eval           EvaluationContext
		expectedResult bool
		expectErr      bool
	}{
		{
			name:    "flag_disabled",
			enabled: false,
			config: Config{
				Default: false,
				Rules: []Rule{
					{
						Action: Action{
							Rollout: ptr.Int(100),
						},
					},
				},
			},
			eval: EvaluationContext{
				FlagKey:    "abc",
				RolloutKey: "def",
			},
			expectedResult: false,
			expectErr:      false,
		},
		{
			name:    "no_rules_default_true",
			enabled: true,
			config: Config{
				Default: true,
			},
			eval:           EvaluationContext{},
			expectedResult: true,
			expectErr:      false,
		},
		{
			name:    "no_rules_default_false",
			enabled: true,
			config: Config{
				Default: false,
			},
			eval: EvaluationContext{
				FlagKey: "abc",
			},
			expectedResult: false,
			expectErr:      false,
		},
		{
			name:    "rollout_rule_100",
			enabled: true,
			config: Config{
				Default: false,
				Rules: []Rule{
					{
						Action: Action{
							Rollout: ptr.Int(100),
						},
					},
				},
			},
			eval: EvaluationContext{
				FlagKey:    "abc",
				RolloutKey: "def",
			},
			expectedResult: true,
			expectErr:      false,
		},
		{
			name:    "rollout_rule_0",
			enabled: true,
			config: Config{
				Default: true,
				Rules: []Rule{
					{
						Action: Action{
							Rollout: ptr.Int(0),
						},
					},
				},
			},
			eval: EvaluationContext{
				FlagKey:    "abc",
				RolloutKey: "def",
			},
			expectedResult: false,
			expectErr:      false,
		},
		{
			name:    "no_rollout_key",
			enabled: true,
			config: Config{
				Default: true,
				Rules: []Rule{
					{
						Action: Action{
							Rollout: ptr.Int(0),
						},
					},
				},
			},
			eval: EvaluationContext{
				FlagKey: "def",
			},
			expectedResult: false,
			expectErr:      true,
		},
		{
			name:    "no_flag_key",
			enabled: true,
			config: Config{
				Default: true,
				Rules: []Rule{
					{
						Action: Action{
							Rollout: ptr.Int(0),
						},
					},
				},
			},
			eval: EvaluationContext{
				RolloutKey: "abc",
			},
			expectedResult: false,
			expectErr:      true,
		},
		{
			name:    "no_flag_key_and_rollout_key",
			enabled: true,
			config: Config{
				Default: true,
				Rules: []Rule{
					{
						Action: Action{
							Rollout: ptr.Int(0),
						},
					},
				},
			},
			eval:           EvaluationContext{},
			expectedResult: false,
			expectErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.enabled, tt.config, tt.eval)
			assert.Equal(tt.expectedResult, result, tt.name)
			assert.Equal(tt.expectErr, err != nil, tt.name)
		})
	}
}
