package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/flag/application"
	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/flag/domain"
	appmock "github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/mocks/flag/application"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name            string
		flagKey         string
		body            string
		setupMockSvc    func(svc *appmock.MockFlagService)
		expectedStatus  int
		expectedEnabled bool
	}{
		{
			name:    "success",
			flagKey: "abc",
			body: `{
				"subject_key": "123",
				"attributes": {
					"country": "UK"
				}
			}`,
			setupMockSvc: func(svc *appmock.MockFlagService) {
				svc.EXPECT().
					Evaluate(mock.Anything, application.EvaluateRequest{
						SubjectKey: "123",
						FlagKey:    "abc",
						Attributes: map[string]string{
							"country": "UK",
						},
					}).
					Return(application.EvaluateResponse{
						Enabled: true,
					}, nil).
					Once()
			},
			expectedStatus:  http.StatusOK,
			expectedEnabled: true,
		},
		{
			name:           "too_large_body",
			flagKey:        "abc",
			body:           strings.Repeat("x", (1024*5)+1),
			expectedStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name:           "body_absent",
			flagKey:        "abc",
			body:           "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid_request",
			flagKey:        "abc",
			body:           `{"subject_key": `,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "flag_not_found",
			flagKey: "abc",
			body:    `{"subject_key": "123"}`,
			setupMockSvc: func(svc *appmock.MockFlagService) {
				svc.EXPECT().
					Evaluate(mock.Anything, application.EvaluateRequest{
						SubjectKey: "123",
						FlagKey:    "abc",
					}).
					Return(application.EvaluateResponse{}, domain.ErrFlagNotFound).
					Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "evaluation_failed",
			flagKey: "abc",
			body:    `{"subject_key": "123"}`,
			setupMockSvc: func(svc *appmock.MockFlagService) {
				svc.EXPECT().
					Evaluate(mock.Anything, application.EvaluateRequest{
						SubjectKey: "123",
						FlagKey:    "abc",
					}).
					Return(application.EvaluateResponse{}, domain.ErrEvaluationFailed).
					Once()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "unexpected_service_error",
			flagKey: "abc",
			body:    `{"subject_key": "123"}`,
			setupMockSvc: func(svc *appmock.MockFlagService) {
				svc.EXPECT().
					Evaluate(mock.Anything, application.EvaluateRequest{
						SubjectKey: "123",
						FlagKey:    "abc",
					}).
					Return(application.EvaluateResponse{}, errors.New("unexpected error")).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mSvc := new(appmock.MockFlagService)
			if tt.setupMockSvc != nil {
				tt.setupMockSvc(mSvc)
			}

			mux := http.NewServeMux()
			RegisterEndpoints(mux, mSvc)

			req := httptest.NewRequest(http.MethodPost, "/flags/"+tt.flagKey+"/evaluate", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if rr.Code == http.StatusOK {
				var resp map[string]any
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(t, err, "Success response must be valid JSON")

				enabled, ok := resp["enabled"]
				assert.True(t, ok, "Success response must contain 'enabled' key")
				assert.Equal(t, tt.expectedEnabled, enabled)
			}
		})
	}
}
