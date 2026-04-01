// Package problems provides a robust implementation of RFC 9457 (Problem Details for HTTP APIs).
//
// The package is designed around a "Gatekeeper" architecture: the 'New' function 
// validates and sanitizes all inputs to return a 'problem' instance that is 
// guaranteed to be valid. All other package functions treat the 'problem' type 
// as a trusted, read-only internal state, ensuring high reliability and 
// consistent API responses.
package problems

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"unicode"
)

var (
	ErrInvalidStatus = errors.New("status must be in range between 100 and 599")

	ErrTypeAbsent                = errors.New("Type field is absent")
	ErrTitleAbsent               = errors.New("Title field is absent")
	ErrStatusAbsent              = errors.New("Status field is absent")
	ErrReservedField             = errors.New("extension member may not be one of the 5 reserved fields")
	ErrExtValidationShortKey     = errors.New("extension key should be 3 or more characters long")
	ErrExtValidationStartLetter  = errors.New("extension key should start with a letter")
	ErrExtValidationInvalidChars = errors.New("extension key may contain only digits, letters and '_'")
	ErrNilExtensions             = errors.New("Extensions field is nil")

	ErrStatusCodesDiffer = errors.New("problem and given explicitly status codes differ")
	ErrNilWriter         = errors.New("ResponseWriter is nil")
)

type ProblemParams struct {
	Type       string
	Title      string
	Status     int
	Detail     string
	Instance   string
	Extensions map[string]any
}

type problem struct {
	typ        string
	title      string
	status     int
	detail     string
	instance   string
	extensions map[string]any
}

func (p problem) Typ() string                { return p.typ }
func (p problem) Title() string              { return p.title }
func (p problem) Status() int                { return p.status }
func (p problem) Detail() string             { return p.detail }
func (p problem) Instance() string           { return p.instance }
func (p problem) Extensions() map[string]any { return p.extensions }

// New creates a problem struct that strictly adheres to RFC 9457.
// It prioritizes returning a usable problem over failing: if the input is 
// invalid, it applies sensible defaults (e.g., "about:blank" for missing types). 
// Any returned error indicates a validation failure in the provided ProblemParams, 
// though the resulting problem is always safe for use in WriteProblem.
func New(p ProblemParams) (pr problem, err error) {
	// The 5 reserved field validation
	e := validateProblemParams(p)
	if e != nil {
		if errors.Is(e, ErrTypeAbsent) {
			p.Type = "about:blank"
		}
		if errors.Is(e, ErrTitleAbsent) {
			if title := http.StatusText(p.Status); title != "" {
				p.Title = title
			} else {
				p.Title = http.StatusText(http.StatusInternalServerError)
			}
		}
		if errors.Is(e, ErrStatusAbsent) || errors.Is(e, ErrInvalidStatus) {
			p.Status = http.StatusInternalServerError
		}
		if errors.Is(e, ErrNilExtensions) {
			p.Extensions = make(map[string]any, 0)
		}
		err = errors.Join(err, e)
	}

	// Extension member key validation
	ext := make(map[string]any, len(p.Extensions))
	for k, v := range p.Extensions {
		if e := validateExtensionKey(k); e == nil {
			ext[k] = v
		} else {
			err = errors.Join(err, e)
		}
	}
	p.Extensions = ext

	return problem{
		typ:        p.Type,
		title:      p.Title,
		status:     p.Status,
		detail:     p.Detail,
		instance:   p.Instance,
		extensions: p.Extensions,
	}, err
}

// WriteProblem serializes the problem to JSON and writes it to the ResponseWriter.
// It sets the "Content-Type" header to "application/problem+json" and uses the 
// status code embedded in the problem. If the provided statusCode differs from 
// the problem's status, the problem's internal status takes precedence.
func WriteProblem(w http.ResponseWriter, statusCode int, p problem) (err error) {
	if w == nil {
		return ErrNilWriter
	}

	if p.status != statusCode {
		err = errors.Join(err, fmt.Errorf("%w: p.Status=%v, statusCode=%v", ErrStatusCodesDiffer, p.status, statusCode))
		statusCode = p.status
	}

	problem := map[string]any{
		"type":   p.typ,
		"title":  p.title,
		"status": p.status,
	}

	if p.detail != "" {
		problem["detail"] = p.detail
	}
	if p.instance != "" {
		problem["instance"] = p.instance
	}

	for k, v := range p.extensions {
		problem[k] = v
	}

	data, e := json.Marshal(problem)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		err = errors.Join(err, fmt.Errorf("marshal problem: %w", e))
		return err
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(statusCode)

	if _, e := w.Write(data); e != nil {
		err = errors.Join(err, fmt.Errorf("write response: %w", e))
	}

	return err
}

// validateProblemParams checks the core RFC 9457 fields of a ProblemParams.
// It does not validate extension members.
func validateProblemParams(p ProblemParams) (err error) {
	if p.Type == "" {
		err = errors.Join(err, ErrTypeAbsent)
	}
	if p.Title == "" {
		err = errors.Join(err, ErrTitleAbsent)
	}
	if p.Status < 100 || p.Status > 599 {
		if p.Status == 0 {
			err = errors.Join(err, ErrStatusAbsent)
		} else {
			err = errors.Join(err, ErrInvalidStatus)
		}
	}
	if p.Extensions == nil {
		err = errors.Join(err, ErrNilExtensions)
	}

	return err
}

// validateExtensionKey ensures a key follows RFC 9457 requirements: 
// it must be at least 3 characters, start with a letter, contain only 
// alphanumeric/underscore characters, and not conflict with reserved names.
func validateExtensionKey(k string) (err error) {
	switch k {
	case "type", "title", "status", "detail", "instance":
		err = errors.Join(err, ErrReservedField)
	}

	if len(k) < 3 {
		err = errors.Join(err, ErrExtValidationShortKey)
	}

	for i, r := range k {
		if i == 0 && !unicode.IsLetter(r) {
			err = errors.Join(err, ErrExtValidationStartLetter)
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			err = errors.Join(err, ErrExtValidationInvalidChars)
		}
	}
	return err
}
