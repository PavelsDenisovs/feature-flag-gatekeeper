package application

import (
	"errors"
	"testing"

	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/flag/domain"
	"github.com/stretchr/testify/assert"
)

func TestEvaluate(t *testing.T) {
	for _, tt := range evaluateTests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			repo := tt.setupRepo()
			svc := New(repo)
			ctx := t.Context()

			res, err := svc.Evaluate(ctx, tt.req)

			assert.Equal(tt.expectErr, err != nil)

			if err != nil {
				assert.False(res.Enabled)
			}

			if !tt.expectErr {
				assert.Equal(tt.expectedRes, res)
			}

			if tt.expectedErr != nil {
				assert.ErrorIs(err, tt.expectedErr)
			}
		})
	}
}

var errFetch = errors.New("some error")

var evaluateTests = []struct {
	name        string
	req         EvaluateRequest
	setupRepo   func() *mockFlagRepository
	expectedRes EvaluateResponse
	expectErr   bool
	expectedErr error
}{
	{
		name: "no_flag_key",
		req: EvaluateRequest{
			SubjectKey: "abc",
		},
		setupRepo: func() *mockFlagRepository {
			return &mockFlagRepository{}
		},
		expectedRes: EvaluateResponse{
			Enabled: false,
		},
		expectErr: true,
	},
	{
		name: "no_matching_flag_by_key",
		req: EvaluateRequest{
			FlagKey:    "abc",
			SubjectKey: "abc",
		},
		setupRepo: func() *mockFlagRepository {
			return &mockFlagRepository{
				flags: []domain.Flag{
					{
						Key:     "a",
						Config:  domain.Config{},
						Enabled: true,
					},
					{
						Key:     "b",
						Config:  domain.Config{},
						Enabled: true,
					},
				},
			}
		},
		expectedRes: EvaluateResponse{
			Enabled: false,
		},
		expectErr: true,
	},
	{
		name: "matching_flag_with_no_rules",
		req: EvaluateRequest{
			FlagKey:    "a",
			SubjectKey: "abc",
		},
		setupRepo: func() *mockFlagRepository {
			return &mockFlagRepository{
				flags: []domain.Flag{
					{
						Key: "a",
						Config: domain.Config{
							Default: true,
						},
						Enabled: true,
					},
					{
						Key: "b",
						Config: domain.Config{
							Default: true,
							Rules: []domain.Rule{
								{
									Conditions: []domain.Condition{
										{
											Kind:       domain.ConditionKindRollout,
											Percentage: 100,
										},
									},
									Result: true,
								},
							},
						},
						Enabled: true,
					},
				},
			}
		},
		expectedRes: EvaluateResponse{
			Enabled: true,
		},
		expectErr: false,
	},
	{
		name: "matching_flag_with_100_rollout",
		req: EvaluateRequest{
			FlagKey:    "b",
			SubjectKey: "abc",
		},
		setupRepo: func() *mockFlagRepository {
			return &mockFlagRepository{
				flags: []domain.Flag{
					{
						Key:     "a",
						Config:  domain.Config{},
						Enabled: true,
					},
					{
						Key: "b",
						Config: domain.Config{
							Default: false,
							Rules: []domain.Rule{
								{
									Conditions: []domain.Condition{
										{
											Kind:       domain.ConditionKindRollout,
											Percentage: 100,
										},
									},
									Result: true,
								},
							},
						},
						Enabled: true,
					},
				},
			}
		},
		expectedRes: EvaluateResponse{
			Enabled: true,
		},
		expectErr: false,
	},
	{
		name: "matching_flag_with_0_rollout",
		req: EvaluateRequest{
			FlagKey:    "b",
			SubjectKey: "abc",
		},
		setupRepo: func() *mockFlagRepository {
			return &mockFlagRepository{
				flags: []domain.Flag{
					{
						Key:     "a",
						Config:  domain.Config{},
						Enabled: true,
					},
					{
						Key: "b",
						Config: domain.Config{
							Default: false,
							Rules: []domain.Rule{
								{
									Conditions: []domain.Condition{
										{
											Kind:       domain.ConditionKindRollout,
											Percentage: 0,
										},
									},
									Result: true,
								},
							},
						},
						Enabled: true,
					},
				},
			}
		},
		expectedRes: EvaluateResponse{
			Enabled: false,
		},
		expectErr: false,
	},
	{
		name: "matching_flag_disabled",
		req: EvaluateRequest{
			FlagKey:    "b",
			SubjectKey: "abc",
		},
		setupRepo: func() *mockFlagRepository {
			return &mockFlagRepository{
				flags: []domain.Flag{
					{
						Key:     "a",
						Config:  domain.Config{},
						Enabled: true,
					},
					{
						Key: "b",
						Config: domain.Config{
							Default: false,
							Rules: []domain.Rule{
								{
									Conditions: []domain.Condition{
										{
											Kind:       domain.ConditionKindRollout,
											Percentage: 100,
										},
									},
									Result: true,
								},
							},
						},
						Enabled: false,
					},
				},
			}
		},
		expectedRes: EvaluateResponse{
			Enabled: false,
		},
		expectErr: false,
	},
	{
		name: "no_subject_key_for_disabled_rollout",
		req: EvaluateRequest{
			FlagKey: "b",
		},
		setupRepo: func() *mockFlagRepository {
			return &mockFlagRepository{
				flags: []domain.Flag{
					{
						Key:     "a",
						Config:  domain.Config{},
						Enabled: true,
					},
					{
						Key: "b",
						Config: domain.Config{
							Default: true,
							Rules: []domain.Rule{
								{
									Conditions: []domain.Condition{
										{
											Kind:       domain.ConditionKindRollout,
											Percentage: 100,
										},
									},
									Result: true,
								},
							},
						},
						Enabled: false,
					},
				},
			}
		},
		expectedRes: EvaluateResponse{
			Enabled: true,
		},
		expectErr: false,
	},
	{
		name: "injected_error_in_fetch_flag",
		req: EvaluateRequest{
			FlagKey:    "b",
			SubjectKey: "abc",
		},
		setupRepo: func() *mockFlagRepository {
			return &mockFlagRepository{
				flags: []domain.Flag{
					{
						Key:     "a",
						Config:  domain.Config{},
						Enabled: true,
					},
					{
						Key: "b",
						Config: domain.Config{
							Default: true,
							Rules: []domain.Rule{
								{
									Conditions: []domain.Condition{
										{
											Kind:       domain.ConditionKindRollout,
											Percentage: 100,
										},
									},
									Result: true,
								},
							},
						},
						Enabled: true,
					},
				},
				errFetch: errFetch,
			}
		},
		expectedRes: EvaluateResponse{
			Enabled: false,
		},
		expectErr:   true,
		expectedErr: errFetch,
	},
}
