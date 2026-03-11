package application

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/features/flag/domain"
	"github.com/aws/smithy-go/ptr"
	"github.com/stretchr/testify/assert"
)

func TestEvaluate(t *testing.T) {
	for _, tt := range evaluateTests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			repo := tt.setupRepo()
			svc := New(repo)

			res, err := svc.Evaluate(context.Background(), tt.req)
			
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

type mockFlagRepository struct{
	flags    []domain.Flag
	errFetch error
}

func (r *mockFlagRepository) FetchFlagByKey(ctx context.Context, flagKey string) (domain.Flag, error) {
	if r.errFetch != nil {
		return domain.Flag{}, r.errFetch
	}
	for _, f := range r.flags {
		if f.Key == flagKey {
			return f, nil
		}
	}
	return domain.Flag{}, fmt.Errorf("Flag with key %s is not found", flagKey)
}

var errFetch = errors.New("some error")

var evaluateTests = []struct{
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
				RolloutKey: "abc",
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
				FlagKey: "abc",
				RolloutKey: "abc",
			},
			setupRepo: func() *mockFlagRepository {
				return &mockFlagRepository{
					flags: []domain.Flag{
						{
							Key: "a",
							Config: domain.Config{},
							Enabled: true,
						},
						{
							Key: "b",
							Config: domain.Config{},
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
				FlagKey: "a",
				RolloutKey: "abc",
			},
			setupRepo: func() *mockFlagRepository {
				return &mockFlagRepository{
					flags: []domain.Flag{
						{
							Key: "a",
							Config: domain.Config{
								Default: ptr.Bool(true),
							},
							Enabled: true,
						},
						{
							Key: "b",
							Config: domain.Config{
								Default: ptr.Bool(true),
								Rules: []domain.Rule{
									{
										Action: domain.Action{
											Rollout: ptr.Int(100),
										},
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
				FlagKey: "b",
				RolloutKey: "abc",
			},
			setupRepo: func() *mockFlagRepository {
				return &mockFlagRepository{
					flags: []domain.Flag{
						{
							Key: "a",
							Config: domain.Config{},
							Enabled: true,
						},
						{
							Key: "b",
							Config: domain.Config{
								Default: ptr.Bool(false),
								Rules: []domain.Rule{
									{
										Action: domain.Action{
											Rollout: ptr.Int(100),
										},
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
				FlagKey: "b",
				RolloutKey: "abc",
			},
			setupRepo: func() *mockFlagRepository {
				return &mockFlagRepository{
					flags: []domain.Flag{
						{
							Key: "a",
							Config: domain.Config{},
							Enabled: true,
						},
						{
							Key: "b",
							Config: domain.Config{
								Default: ptr.Bool(true),
								Rules: []domain.Rule{
									{
										Action: domain.Action{
											Rollout: ptr.Int(0),
										},
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
				FlagKey: "b",
				RolloutKey: "abc",
			},
			setupRepo: func() *mockFlagRepository {
				return &mockFlagRepository{
					flags: []domain.Flag{
						{
							Key: "a",
							Config: domain.Config{},
							Enabled: true,
						},
						{
							Key: "b",
							Config: domain.Config{
								Default: ptr.Bool(true),
								Rules: []domain.Rule{
									{
										Action: domain.Action{
											Rollout: ptr.Int(100),
										},
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
			name: "no_rollout_key_for_disabled_rollout",
			req: EvaluateRequest{
				FlagKey: "b",
			},
			setupRepo: func() *mockFlagRepository {
				return &mockFlagRepository{
					flags: []domain.Flag{
						{
							Key: "a",
							Config: domain.Config{},
							Enabled: true,
						},
						{
							Key: "b",
							Config: domain.Config{
								Default: ptr.Bool(true),
								Rules: []domain.Rule{
									{
										Action: domain.Action{
											Rollout: ptr.Int(100),
										},
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
			name: "matching_flag_without_default",
			req: EvaluateRequest{
				FlagKey: "a",
				RolloutKey: "abc",
			},
			setupRepo: func() *mockFlagRepository {
				return &mockFlagRepository{
					flags: []domain.Flag{
						{
							Key: "a",
							Config: domain.Config{
								Rules: []domain.Rule{
									{
										Action: domain.Action{
											Rollout: ptr.Int(100),
										},
									},
								},
							},
							Enabled: true,
						},
						{
							Key: "b",
							Config: domain.Config{
								Default: ptr.Bool(true),
								Rules: []domain.Rule{
									{
										Action: domain.Action{
											Rollout: ptr.Int(100),
										},
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
			expectErr: true,
		},
		{
			name: "injected_error_in_fetch_flag",
			req: EvaluateRequest{
				FlagKey: "b",
				RolloutKey: "abc",
			},
			setupRepo: func() *mockFlagRepository {
				return &mockFlagRepository{
					flags: []domain.Flag{
						{
							Key: "a",
							Config: domain.Config{},
							Enabled: true,
						},
						{
							Key: "b",
							Config: domain.Config{
								Default: ptr.Bool(true),
								Rules: []domain.Rule{
									{
										Action: domain.Action{
											Rollout: ptr.Int(100),
										},
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
			expectErr: true,
			expectedErr: errFetch,
		},
	}