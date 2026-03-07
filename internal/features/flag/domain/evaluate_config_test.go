package domain

import (
	"sort"
	"testing"

	"github.com/aws/smithy-go/ptr"
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

func TestRule_Evaluate(t *testing.T) {
	assert := assert.New(t)
	tests := []struct{
		name string
		rule Rule
		eval EvaluationContext
		expectedResult        bool
		expectedMatch         bool
		expectErr             bool
		expectedMissingFields []Field
	}{
		{
			name: "missing_rollout_key_and_flag_key",
			rule: Rule{
				Action: Action{
					Rollout: ptr.Int(50),
				},
			},
			eval: EvaluationContext{},
			expectedResult: false,
			expectedMatch: false,
			expectErr: true,
			expectedMissingFields: []Field{RolloutKeyField, FlagKeyField},
		},
		{
			name: "missing_flag_key_and_rollout_key",
			rule: Rule{
				Action: Action{
					Rollout: ptr.Int(50),
				},
			},
			eval: EvaluationContext{},
			expectedResult: false,
			expectedMatch: false,
			expectErr: true,
			expectedMissingFields: []Field{FlagKeyField, RolloutKeyField},
		},
		{
			name: "missing_rollout_key",
			rule: Rule{
				Action: Action{
					Rollout: ptr.Int(50),
				},
			},
			eval: EvaluationContext{
				FlagKey: "abc",
			},
			expectedResult: false,
			expectedMatch: false,
			expectErr: true,
			expectedMissingFields: []Field{RolloutKeyField},
		},
		{
			name: "missing_flag_key",
			rule: Rule{
				Action: Action{
					Rollout: ptr.Int(50),
				},
			},
			eval: EvaluationContext{
				RolloutKey: "abc",
			},
			expectedResult: false,
			expectedMatch: false,
			expectErr: true,
			expectedMissingFields: []Field{FlagKeyField},
		},
		{
			name: "all_fields_present",
			rule: Rule{
				Action: Action{
					Rollout: ptr.Int(100),
				},
			},
			eval: EvaluationContext{
				RolloutKey: "abc",
				FlagKey: "def",
			},
			expectedResult: true,
			expectedMatch: true,
			expectErr: false,
			expectedMissingFields: []Field{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, match, err := tt.rule.Evaluate(tt.eval)
			assert.Equal(tt.expectErr, err != nil)
			assert.Equal(tt.expectedMatch, match)
			assert.Equal(tt.expectedResult, result)

			if err != nil {
				missingFields := err.(MissingFieldError).Fields
				assert.Equal(len(tt.expectedMissingFields), len(missingFields))

				sortFieldSlices(tt.expectedMissingFields, missingFields)
				assert.Equal(tt.expectedMissingFields, missingFields)
			}
		})
	}
}

func sortFieldSlices(slices... []Field) {
	for _, sl := range slices {
		var temp []string
		for _, f := range sl {
			temp = append(temp, string(f))
		}
		sort.Strings(temp)
		for i, str := range temp {
			sl[i] = Field(str[i])
		}
	}
}