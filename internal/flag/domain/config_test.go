package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRolloutAccept(t *testing.T) {
	assert := assert.New(t)
	t.Run("Bucket smaller than rollout -> accept", func(t *testing.T) {
		assert.True(rolloutAccept(10, 80))
	})
	t.Run("Bucket bigger than rollout -> not accept", func(t *testing.T) {
		assert.False(rolloutAccept(90, 70))
	})
	t.Run("Bucket is equal to rollout -> not accept", func(t *testing.T) {
		assert.False(rolloutAccept(50, 50))
	})
}

func TestBucket(t *testing.T) {
	t.Run("Two bucket calls with the same input -> equal results", func(t *testing.T) {
		fk := "abc"
		rk := "def"
		assert.Equal(t, bucket(fk, rk), bucket(fk, rk))
	})
}

func TestEvaluateCondition(t *testing.T) {
	tests := []struct {
		name          string
		cond          Condition
		eval          EvaluationContext
		expectedMatch bool
		expectedErr   error
	}{
		{
			name: "valid_input_successful_evaluation",
			cond: Condition{
				Kind:       ConditionKindRollout,
				Percentage: 100,
			},
			eval: EvaluationContext{
				SubjectKey: "abc",
				FlagKey:    "abc",
			},
			expectedMatch: true,
			expectedErr:   nil,
		},
		{
			name: "valid_input_unsuccessful_evaluation",
			cond: Condition{
				Kind:       ConditionKindRollout,
				Percentage: 0,
			},
			eval: EvaluationContext{
				SubjectKey: "abc",
				FlagKey:    "abc",
			},
			expectedMatch: false,
			expectedErr:   nil,
		},
		{
			name: "unknown_condition_kind",
			cond: Condition{
				Kind: ConditionKind("abcabc"),
			},
			eval:          EvaluationContext{},
			expectedMatch: false,
			expectedErr:   ErrUnknownConditionKind,
		},
		{
			name: "unknown_operator",
			cond: Condition{
				Kind:      ConditionKindAttribute,
				Attribute: "country",
				Operator:  Operator("abcabc"),
				Value:     "UK",
			},
			eval: EvaluationContext{
				Attributes: map[string]string{
					"country": "UK",
				},
			},
			expectedMatch: false,
			expectedErr:   ErrUnknownOperator,
		},
		{
			name: "missing_subject_key_for_rollout",
			cond: Condition{
				Kind:       ConditionKindRollout,
				Percentage: 100,
			},
			eval: EvaluationContext{
				FlagKey: "abc",
			},
			expectedMatch: false,
			expectedErr:   ErrIncompleteContext,
		},
		{
			name: "missing_flag_key_for_rollout",
			cond: Condition{
				Kind:       ConditionKindRollout,
				Percentage: 100,
			},
			eval: EvaluationContext{
				SubjectKey: "abc",
			},
			expectedMatch: false,
			expectedErr:   ErrIncompleteContext,
		},
		{
			name: "missing_attribute",
			cond: Condition{
				Kind:      ConditionKindAttribute,
				Attribute: "country",
				Operator:  OperatorEquals,
				Value:     "UK",
			},
			eval: EvaluationContext{
				Attributes: map[string]string{
					"city": "London",
				},
			},
			expectedMatch: false,
			expectedErr:   ErrIncompleteContext,
		},
		{
			name: "unmatched_value_of_valid_attribute",
			cond: Condition{
				Kind:      ConditionKindAttribute,
				Attribute: "country",
				Operator:  OperatorEquals,
				Value:     "UK",
			},
			eval: EvaluationContext{
				Attributes: map[string]string{
					"country": "GE",
				},
			},
			expectedMatch: false,
			expectedErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := evaluateCondition(tt.cond, tt.eval)
			assert.ErrorIs(t, err, tt.expectedErr)
			assert.Equal(t, tt.expectedMatch, match)
		})
	}
}

func TestRule_Evaluate(t *testing.T) {
	tests := []struct {
		name           string
		rule           Rule
		eval           EvaluationContext
		expectedResult bool
		expectedMatch  bool
		expectedErr    error
	}{
		{
			name: "rule_match_result_false",
			rule: Rule{
				Conditions: []Condition{
					{
						Kind:       ConditionKindRollout,
						Percentage: 100,
					},
				},
				Result: false,
			},
			eval: EvaluationContext{
				SubjectKey: "abc",
				FlagKey:    "abc",
			},
			expectedResult: false,
			expectedMatch:  true,
			expectedErr:    nil,
		},
		{
			name: "one_of_two_conditions_mismatch",
			rule: Rule{
				Conditions: []Condition{
					{
						Kind:      ConditionKindAttribute,
						Attribute: "country",
						Operator:  OperatorEquals,
						Value:     "UK",
					},
					{
						Kind:      ConditionKindAttribute,
						Attribute: "city",
						Operator:  OperatorEquals,
						Value:     "London",
					},
				},
				Result: true,
			},
			eval: EvaluationContext{
				Attributes: map[string]string{
					"country": "UK",
					"city":    "Liverpool",
				},
			},
			expectedResult: false,
			expectedMatch:  false,
			expectedErr:    nil,
		},
		{
			name: "all_conditions_match",
			rule: Rule{
				Conditions: []Condition{
					{
						Kind:      ConditionKindAttribute,
						Attribute: "country",
						Operator:  OperatorEquals,
						Value:     "UK",
					},
					{
						Kind:      ConditionKindAttribute,
						Attribute: "city",
						Operator:  OperatorEquals,
						Value:     "London",
					},
				},
				Result: true,
			},
			eval: EvaluationContext{
				Attributes: map[string]string{
					"country": "UK",
					"city":    "London",
				},
			},
			expectedResult: true,
			expectedMatch:  true,
			expectedErr:    nil,
		},
		{
			name: "one_condition_fail_one_match",
			rule: Rule{
				Conditions: []Condition{
					{
						Kind:      ConditionKindAttribute,
						Attribute: "country",
						Operator:  OperatorEquals,
						Value:     "UK",
					},
					{
						Kind:      ConditionKindAttribute,
						Attribute: "city",
						Operator:  OperatorEquals,
						Value:     "London",
					},
				},
				Result: true,
			},
			eval: EvaluationContext{
				Attributes: map[string]string{
					"continent": "Europe",
					"city":      "London",
				},
			},
			expectedResult: false,
			expectedMatch:  false,
			expectedErr:    ErrIncompleteContext,
		},
	}

	assert := assert.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, match, err := tt.rule.Evaluate(tt.eval)
			assert.ErrorIs(err, tt.expectedErr)
			assert.Equal(tt.expectedMatch, match)
			assert.Equal(tt.expectedResult, result)
		})
	}
}
