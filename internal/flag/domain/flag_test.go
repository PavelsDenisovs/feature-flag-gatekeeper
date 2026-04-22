package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFlag(t *testing.T) {
	tests := []struct {
		name        string
		params      FlagData
		expectedErr error
	}{
		{
			name: "valid_input",
			params: FlagData{
				Key:         "abc",
				Enabled:     true,
				Description: "abc",
				Config: Config{
					Default: false,
					Version: CurrentConfigVersion,
					Rules: []Rule{
						{
							Conditions: []Condition{
								{
									Kind:       ConditionKindRollout,
									Percentage: 100,
								},
							},
							Result: true,
						},
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "invalid_config_version",
			params: FlagData{
				Key:         "abc",
				Enabled:     true,
				Description: "abc",
				Config: Config{
					Default: false,
					Version: CurrentConfigVersion + 1,
					Rules: []Rule{
						{
							Conditions: []Condition{
								{
									Kind:       ConditionKindRollout,
									Percentage: 100,
								},
							},
							Result: true,
						},
					},
				},
			},
			expectedErr: ErrInvalidConfigVersion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFlag(tt.params)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestFlag_Evaluate(t *testing.T) {
	tests := []struct {
		name           string
		newFlagParams  FlagData
		eval           EvaluationContext
		expectedResult bool
		expectedErr    error
	}{
		{
			name: "valid_input_successful_evaluation",
			newFlagParams: FlagData{
				Key:         "abc",
				Enabled:     true,
				Description: "abc",
				Config: Config{
					Default: false,
					Version: CurrentConfigVersion,
					Rules: []Rule{
						{
							Conditions: []Condition{
								{
									Kind:       ConditionKindRollout,
									Percentage: 100,
								},
							},
							Result: true,
						},
					},
				},
			},
			eval: EvaluationContext{
				SubjectKey: "abc",
				FlagKey:    "abc",
			},
			expectedResult: true,
			expectedErr:    nil,
		},
		{
			name: "valid_input_unsuccessful_evaluation",
			newFlagParams: FlagData{
				Key:         "abc",
				Enabled:     true,
				Description: "abc",
				Config: Config{
					Default: false,
					Version: CurrentConfigVersion,
					Rules: []Rule{
						{
							Conditions: []Condition{
								{
									Kind:       ConditionKindRollout,
									Percentage: 0,
								},
							},
							Result: true,
						},
					},
				},
			},
			eval: EvaluationContext{
				SubjectKey: "abc",
				FlagKey:    "abc",
			},
			expectedResult: false,
			expectedErr:    nil,
		},
		{
			name: "disabled_flag_fallbacks_to_default",
			newFlagParams: FlagData{
				Key:         "abc",
				Enabled:     false,
				Description: "abc",
				Config: Config{
					Default: true,
					Version: CurrentConfigVersion,
					Rules: []Rule{
						{
							Conditions: []Condition{
								{
									Kind:       ConditionKindRollout,
									Percentage: 0,
								},
							},
							Result: false,
						},
					},
				},
			},
			eval: EvaluationContext{
				SubjectKey: "abc",
				FlagKey:    "abc",
			},
			expectedResult: true,
			expectedErr:    nil,
		},
		{
			name: "no_matched_rules_fallbacks_to_default",
			newFlagParams: FlagData{
				Key:         "abc",
				Enabled:     true,
				Description: "abc",
				Config: Config{
					Default: true,
					Version: CurrentConfigVersion,
					Rules: []Rule{
						{
							Conditions: []Condition{
								{
									Kind:       ConditionKindRollout,
									Percentage: 0,
								},
								{
									Kind:      ConditionKindAttribute,
									Attribute: "country",
									Operator:  OperatorEquals,
									Value:     "UK",
								},
							},
							Result: false,
						},
					},
				},
			},
			eval: EvaluationContext{
				SubjectKey: "abc",
				FlagKey:    "abc",
				Attributes: map[string]string{
					"country": "GE",
				},
			},
			expectedResult: true,
			expectedErr:    nil,
		},
		{
			name: "evaluation_fail",
			newFlagParams: FlagData{
				Key:         "abc",
				Enabled:     true,
				Description: "abc",
				Config: Config{
					Default: true,
					Version: CurrentConfigVersion,
					Rules: []Rule{
						{
							Conditions: []Condition{
								{
									Kind:       ConditionKindRollout,
									Percentage: 0,
								},
							},
							Result: false,
						},
					},
				},
			},
			eval: EvaluationContext{
				SubjectKey: "abc",
			},
			expectedResult: false,
			expectedErr:    ErrEvaluationFailed,
		},
		{
			name: "first_rule_mismatch_second_rule_matches",
			newFlagParams: FlagData{
				Key:         "abc",
				Enabled:     true,
				Description: "abc",
				Config: Config{
					Default: false,
					Version: CurrentConfigVersion,
					Rules: []Rule{
						{
							Conditions: []Condition{
								{
									Kind:       ConditionKindRollout,
									Percentage: 0,
								},
							},
							Result: true,
						},
						{
							Conditions: []Condition{
								{
									Kind:       ConditionKindRollout,
									Percentage: 100,
								},
							},
							Result: true,
						},
					},
				},
			},
			eval: EvaluationContext{
				SubjectKey: "abc",
				FlagKey:    "abc",
			},
			expectedResult: true,
			expectedErr:    nil,
		},
	}

	require := require.New(t)

	for _, tt := range tests {
		f, err := NewFlag(tt.newFlagParams)
		require.NoError(err)
		res, err := f.Evaluate(tt.eval)
		require.ErrorIs(err, tt.expectedErr)
		require.Equal(tt.expectedResult, res)
	}
}
