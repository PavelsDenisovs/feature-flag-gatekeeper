package problems

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		input       ProblemParams
		expected    problem
		expectedErr error
	}{
		{
			name:        "valid_problem_params",
			input:       ProblemParams(validProblemParams()),
			expected:    problem(validProblem()),
			expectedErr: nil,
		},
		{
			name: "type_absent",
			input: ProblemParams(validProblemParams().
				WithType(""),
			),
			expected: problem(validProblem().
				WithType("about:blank"),
			),
			expectedErr: ErrTypeAbsent,
		},
		{
			name: "title_absent",
			input: ProblemParams(validProblemParams().
				WithTitle(""),
			),
			expected: problem(validProblem().
				WithTitle(http.StatusText(validProblem().status)),
			),
			expectedErr: ErrTitleAbsent,
		},
		{
			name: "status_absent",
			input: ProblemParams(validProblemParams().
				WithStatus(0),
			),
			expected: problem(validProblem().
				WithStatus(http.StatusInternalServerError),
			),
			expectedErr: ErrStatusAbsent,
		},
		{
			name: "title_as_ext_member",
			input: ProblemParams(validProblemParams().
				SetExtension("title", "A Title"),
			),
			expected:    problem(validProblem()),
			expectedErr: ErrReservedField,
		},
		{
			name: "ext_key_too_short",
			input: ProblemParams(validProblemParams().
				SetExtension("tt", "test"),
			),
			expected:    problem(validProblem()),
			expectedErr: ErrExtValidationShortKey,
		},
		{
			name: "ext_key_start_with_not_letter",
			input: ProblemParams(validProblemParams().
				SetExtension("1test", "test"),
			),
			expected:    problem(validProblem()),
			expectedErr: ErrExtValidationStartLetter,
		},
		{
			name: "ext_key_with_special_characters",
			input: ProblemParams(validProblemParams().
				SetExtension("test$", "test"),
			),
			expected:    problem(validProblem()),
			expectedErr: ErrExtValidationInvalidChars,
		},
		{
			name: "invalid_status",
			input: ProblemParams(validProblemParams().
				WithStatus(999),
			),
			expected: problem(validProblem().
				WithStatus(http.StatusInternalServerError),
			),
			expectedErr: ErrInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New(tt.input)
			assert.ErrorIs(t, err, tt.expectedErr)
			assert.Equal(t, tt.expected, p)
		})
	}
}

func TestWriteProblem(t *testing.T) {
	tests := []struct {
		name           string
		w              http.ResponseWriter
		status         int
		problem        problem
		expectedStatus int
		expectedBody   string
		expectedErr    error
	}{
		{
			name:           "valid_input",
			w:              httptest.NewRecorder(),
			status:         http.StatusForbidden,
			problem:        newValidProblem(t, validProblemParams()),
			expectedStatus: http.StatusForbidden,
			expectedBody: `{
				"type": "https://example.com/out-of-credit",
  			"title": "You do not have enough credit.",
  			"status": 403,
  			"detail": "Your current balance is 30, but that costs 50.",
  			"instance": "/account/12345/msgs/abc",
  			"balance": 30,
  			"counts": 50,
  			"currency": "USD"
			}`,
			expectedErr: nil,
		},
		{
			name:   "mismatched_status_codes",
			w:      httptest.NewRecorder(),
			status: http.StatusBadRequest,
			problem: newValidProblem(t, validProblemParams().
				WithStatus(http.StatusForbidden),
			),
			expectedStatus: http.StatusForbidden,
			expectedBody: `{
				"type": "https://example.com/out-of-credit",
  			"title": "You do not have enough credit.",
  			"status": 403,
  			"detail": "Your current balance is 30, but that costs 50.",
  			"instance": "/account/12345/msgs/abc",
  			"balance": 30,
  			"counts": 50,
  			"currency": "USD"
			}`,
			expectedErr: ErrStatusCodesDiffer,
		},
	}

	assert := assert.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WriteProblem(tt.w, tt.status, tt.problem)
			assert.ErrorIs(err, tt.expectedErr)
			if tt.w != nil {
				if rec, ok := tt.w.(*httptest.ResponseRecorder); ok {
					assert.Equal(tt.expectedStatus, rec.Code)
					assert.Equal("application/problem+json", rec.Result().Header.Get("Content-Type"))
					assert.JSONEq(tt.expectedBody, rec.Body.String())
				}
			}
		})
	}
}

func TestValidateProblemParams(t *testing.T) {
	tests := []struct {
		name        string
		params      ProblemParams
		expectedErr error
	}{
		{
			name: "extensions_absent",
			params: ProblemParams(validProblemParams().
				WithExtensions(nil),
			),
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProblemParams(tt.params)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func newValidProblem(t *testing.T, params testProblemParams) problem {
	t.Helper()

	p, err := New(ProblemParams(params))
	if err != nil {
		t.Fatalf("expected valid problem params, got error: %v", err)
	}
	return p
}

func validProblemParams() testProblemParams {
	return testProblemParams{
		Type:     "https://example.com/out-of-credit",
		Status:   http.StatusForbidden,
		Title:    "You do not have enough credit.",
		Detail:   "Your current balance is 30, but that costs 50.",
		Instance: "/account/12345/msgs/abc",
		Extensions: map[string]any{
			"balance":  30,
			"counts":   50,
			"currency": "USD",
		},
	}
}

func validProblem() testProblem {
	return testProblem{
		typ:      "https://example.com/out-of-credit",
		status:   http.StatusForbidden,
		title:    "You do not have enough credit.",
		detail:   "Your current balance is 30, but that costs 50.",
		instance: "/account/12345/msgs/abc",
		extensions: map[string]any{
			"balance":  30,
			"counts":   50,
			"currency": "USD",
		},
	}
}

type testProblemParams struct {
	Type       string
	Title      string
	Status     int
	Detail     string
	Instance   string
	Extensions map[string]any
}

type testProblem struct {
	typ        string
	title      string
	status     int
	detail     string
	instance   string
	extensions map[string]any
}

func (tpp testProblemParams) WithType(v string) testProblemParams {
	tpp.Type = v
	return tpp
}
func (tpp testProblemParams) WithTitle(v string) testProblemParams {
	tpp.Title = v
	return tpp
}
func (tpp testProblemParams) WithStatus(v int) testProblemParams {
	tpp.Status = v
	return tpp
}
func (tpp testProblemParams) WithDetail(v string) testProblemParams {
	tpp.Detail = v
	return tpp
}
func (tpp testProblemParams) WithInstance(v string) testProblemParams {
	tpp.Instance = v
	return tpp
}
func (tpp testProblemParams) WithExtensions(v map[string]any) testProblemParams {
	tpp.Extensions = v
	return tpp
}

func (tpp testProblemParams) SetExtension(k string, v any) testProblemParams {
	m := make(map[string]any, len(tpp.Extensions))
	for key, val := range tpp.Extensions {
		m[key] = val
	}
	m[k] = v
	tpp.Extensions = m
	return tpp
}

func (tpp testProblemParams) DeleteExtension(k string) testProblemParams {
	m := make(map[string]any, len(tpp.Extensions))
	for key, val := range tpp.Extensions {
		if key != k {
			m[key] = val
		}
	}
	return tpp
}

func (tp testProblem) WithType(v string) testProblem {
	tp.typ = v
	return tp
}
func (tp testProblem) WithTitle(v string) testProblem {
	tp.title = v
	return tp
}
func (tp testProblem) WithStatus(v int) testProblem {
	tp.status = v
	return tp
}
func (tp testProblem) WithDetail(v string) testProblem {
	tp.detail = v
	return tp
}
func (tp testProblem) WithInstance(v string) testProblem {
	tp.instance = v
	return tp
}
func (tp testProblem) WithExtensions(v map[string]any) testProblem {
	tp.extensions = v
	return tp
}

func (tp testProblem) SetExtension(k string, v any) testProblem {
	m := make(map[string]any, len(tp.extensions))
	for key, val := range tp.extensions {
		m[key] = val
	}
	m[k] = v
	tp.extensions = m
	return tp
}

func (tp testProblem) DeleteExtension(k string) testProblem {
	m := make(map[string]any, len(tp.extensions))
	for key, val := range tp.extensions {
		if key != k {
			m[key] = val
		}
	}
	return tp
}
