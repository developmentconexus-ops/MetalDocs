// Package problem defines RFC 9457 "Problem Details" payloads used for
// consistent HTTP API error responses.
package problem

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Problem is an RFC 9457 problem details response body.
type Problem struct {
	Type     string       `json:"type,omitempty"`
	Title    string       `json:"title"`
	Status   int          `json:"status"`
	Detail   string       `json:"detail,omitempty"`
	Instance string       `json:"instance,omitempty"`
	Code     Code         `json:"code"`
	Errors   []FieldError `json:"errors,omitempty"`
}

// FieldError describes a validation error on a specific input field.
type FieldError struct {
	Field   string `json:"field"`
	Code    Code   `json:"code"`
	Message string `json:"message"`
}

// NewFor creates a Problem whose status comes from the code's registration.
// This is the DEFAULT constructor: ADR 0089 decision 3 binds the status to the
// code, so the same condition cannot be 409 in one module and 412 in another.
//
// It panics on an unregistered (including zero) Code — a Problem with no status
// binding has no defensible status to emit.
func NewFor(code Code, title string) *Problem {
	status, ok := StatusFor(code)
	if !ok {
		panic(fmt.Sprintf("problem.NewFor: code %q is not registered", code.String()))
	}
	return &Problem{
		Title:  title,
		Status: status,
		Code:   code,
	}
}

// New creates a Problem with an EXPLICIT status, overriding the code's
// registered default.
//
// The status stays a required argument precisely so the override is visible at
// the call site rather than an implicit fallback: reviewers and the api-lint
// drift rule can see, at the line that emits it, that this response deviates
// from the registered binding. Prefer NewFor; reach for New only for a
// documented exception, and annotate it.
//
// Panics if status is outside the valid HTTP range [100, 599].
func New(status int, code Code, title string) *Problem {
	if status < 100 || status > 599 {
		panic(fmt.Sprintf("problem.New: invalid HTTP status %d", status))
	}
	return &Problem{
		Title:  title,
		Status: status,
		Code:   code,
	}
}

// WithDetail sets detail and returns the same Problem.
func (p *Problem) WithDetail(detail string) *Problem {
	p.Detail = detail
	return p
}

// WithInstance sets instance and returns the same Problem.
func (p *Problem) WithInstance(instance string) *Problem {
	p.Instance = instance
	return p
}

// WithFieldError appends a field error and returns the same Problem.
func (p *Problem) WithFieldError(field string, code Code, message string) *Problem {
	p.Errors = append(p.Errors, FieldError{Field: field, Code: code, Message: message})
	return p
}

// WithType sets the RFC 9457 type URI and returns the same Problem.
func (p *Problem) WithType(typeURI string) *Problem {
	p.Type = typeURI
	return p
}

// FromValidation creates a canonical validation problem.
func FromValidation(fields []FieldError) *Problem {
	return &Problem{
		Title:  "Validation failed",
		Status: http.StatusBadRequest,
		Code:   CodeValidationError,
		Errors: fields,
	}
}

// Write writes a problem details response as application/problem+json.
func Write(w http.ResponseWriter, p *Problem) error {
	body, err := json.Marshal(p)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	_, _ = w.Write(body)
	return nil
}

// Error returns a short error-string representation of the problem.
func (p *Problem) Error() string {
	return fmt.Sprintf("problem: %d %s (%s)", p.Status, p.Title, p.Code)
}
