package application_test

import (
	"errors"
	"log"
	"testing"

	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/flag/application"
	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/flag/domain"
	appmock "github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/mocks/flag/application"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(repo *appmock.MockFlagRepository)
		req         application.EvaluateRequest
		expectedRes application.EvaluateResponse
		expectedErr error
	}{
		{
			name: "valid_input",
			req: application.EvaluateRequest{
				FlagKey:    "abc",
				SubjectKey: "abc",
			},
			setupMock: func(repo *appmock.MockFlagRepository) {
				repo.EXPECT().
					FetchFlagByKey(mock.Anything, mock.Anything).
					Return(newFlag(domain.FlagData{
						Key:         "abc",
						Enabled:     true,
						Description: "abc",
						Config: domain.Config{
							Default: false,
							Version: domain.CurrentConfigVersion,
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
					}), nil).
					Once()
			},
			expectedRes: application.EvaluateResponse{
				Enabled: true,
			},
			expectedErr: nil,
		},
		{
			name:      "no_flag_key",
			setupMock: nil,
			req: application.EvaluateRequest{
				SubjectKey: "abc",
			},
			expectedRes: application.EvaluateResponse{
				Enabled: false,
			},
			expectedErr: application.ErrNoFlagKey,
		},
		{
			name: "no_matching_flag_by_key",
			req: application.EvaluateRequest{
				FlagKey:    "abc",
				SubjectKey: "abc",
			},
			setupMock: func(repo *appmock.MockFlagRepository) {
				repo.EXPECT().
					FetchFlagByKey(mock.Anything, mock.Anything).
					Return(nil, application.ErrFlagNotFound).
					Once()
			},
			expectedRes: application.EvaluateResponse{
				Enabled: false,
			},
			expectedErr: application.ErrFlagNotFound,
		},
		{
			name: "evaluation_fail",
			req: application.EvaluateRequest{
				FlagKey:    "abc",
				SubjectKey: "abc",
				Attributes: map[string]string{
					"city": "London",
				},
			},
			setupMock: func(repo *appmock.MockFlagRepository) {
				repo.EXPECT().
					FetchFlagByKey(mock.Anything, mock.Anything).
					Return(newFlag(domain.FlagData{
						Key:         "abc",
						Enabled:     true,
						Description: "abc",
						Config: domain.Config{
							Default: false,
							Version: domain.CurrentConfigVersion,
							Rules: []domain.Rule{
								{
									Conditions: []domain.Condition{
										{
											Kind:      domain.ConditionKindAttribute,
											Attribute: "country",
											Operator:  domain.OperatorEquals,
											Value:     "UK",
										},
									},
									Result: true,
								},
							},
						},
					}), nil).
					Once()
			},
			expectedRes: application.EvaluateResponse{
				Enabled: false,
			},
			expectedErr: domain.ErrEvaluationFailed,
		},
	}

	assert := assert.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := appmock.NewMockFlagRepository(t)
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}
			svc := application.NewFlagService(repo)
			ctx := t.Context()

			res, err := svc.Evaluate(ctx, tt.req)
			assert.ErrorIs(err, tt.expectedErr)
			assert.Equal(tt.expectedRes, res)
		})
	}
}

var errFetch = errors.New("some error")

func newFlag(params domain.FlagData) *domain.Flag {
	f, err := domain.NewFlag(params)
	if err != nil {
		log.Fatalf("failed to create flag with params: %v", params)
	}

	return f
}
