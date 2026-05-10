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
	Code     string       `json:"code"`
	Errors   []FieldError `json:"errors,omitempty"`
}

// FieldError describes a validation error on a specific input field.
type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// New creates a Problem with required fields.
func New(status int, code, title string) *Problem {
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
func (p *Problem) WithFieldError(field, code, message string) *Problem {
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
