package domain

import (
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/stretchr/testify/assert"
)

func TestEvaluate(t *testing.T) {
	assert := assert.New(t)
	tests := []struct{
		name   string
		config Config
		eval   EvaluationContext
		expectedResult bool
		expectErr      bool
	}{
		{
			name: "no_rules",
			config: Config{
				Default: true,
			},
			eval: EvaluationContext{},
			expectedResult: true,
			expectErr: false,
		},
		{
			name: "rollout_rule_100",
			config: Config{
				Default: false,
				Rules: []Rule{
					Rule{
						Action: Action{
							Rollout: ptr.Int(100),
						},
					},
				},
			},
			eval: EvaluationContext{
				RolloutKey: "abc",
				FlagKey: "def",
			},
			expectedResult: true,
			expectErr: false,
		},
		{
			name: "rollout_rule_0",
			config: Config{
				Default: true,
				Rules: []Rule{
					Rule{
						Action: Action{
							Rollout: ptr.Int(0),
						},
					},
				},
			},
			eval: EvaluationContext{
				RolloutKey: "abc",
				FlagKey: "def",
			},
			expectedResult: false,
			expectErr: false,
		},
		{
			name: "rollout_rule_0",
			config: Config{
				Default: true,
				Rules: []Rule{
					Rule{
						Action: Action{
							Rollout: ptr.Int(0),
						},
					},
				},
			},
			eval: EvaluationContext{
				RolloutKey: "abc",
				FlagKey: "def",
			},
			expectedResult: false,
			expectErr: false,
		},
		{
			name: "no_rollout_key",
			config: Config{
				Default: true,
				Rules: []Rule{
					Rule{
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
			expectErr: true,
		},
		{
			name: "no_flag_key",
			config: Config{
				Default: true,
				Rules: []Rule{
					Rule{
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
			expectErr: true,
		},
		{
			name: "no_flag_key_and_rollout_key",
			config: Config{
				Default: true,
				Rules: []Rule{
					Rule{
						Action: Action{
							Rollout: ptr.Int(0),
						},
					},
				},
			},
			eval: EvaluationContext{},
			expectedResult: false,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.config, tt.eval)
			assert.Equal(tt.expectedResult, result)
			assert.Equal(tt.expectErr, err != nil)
		})
	}
}