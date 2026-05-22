package problem

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
)

func TestProblem_MarshalJSON_OmitsOptionalEmptyFields(t *testing.T) {
	tests := []struct {
		name string
		p    Problem
	}{
		{
			name: "required fields only",
			p: Problem{
				Title:  "Validation failed",
				Status: 400,
				Code:   CodeValidationError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.p)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			var m map[string]any
			if err := json.Unmarshal(b, &m); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			if _, ok := m["type"]; ok {
				t.Fatalf("unexpected key type")
			}
			if _, ok := m["detail"]; ok {
				t.Fatalf("unexpected key detail")
			}
			if _, ok := m["instance"]; ok {
				t.Fatalf("unexpected key instance")
			}
			if _, ok := m["errors"]; ok {
				t.Fatalf("unexpected key errors")
			}
		})
	}
}

func TestProblem_MarshalJSON_IncludesAllFields(t *testing.T) {
	tests := []struct {
		name string
		p    Problem
	}{
		{
			name: "all fields",
			p: Problem{
				Type:     "https://errors.metaldocs.io/validation",
				Title:    "Validation failed",
				Status:   400,
				Detail:   "request body invalid",
				Instance: "/v1/docs/1",
				Code:     CodeValidationError,
				Errors: []FieldError{{
					Field:   "key",
					Code:    FieldCodeRequired,
					Message: "is required",
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.p)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			var m map[string]any
			if err := json.Unmarshal(b, &m); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			if m["type"] != tt.p.Type || m["title"] != tt.p.Title || m["detail"] != tt.p.Detail || m["instance"] != tt.p.Instance || m["code"] != string(tt.p.Code) {
				t.Fatalf("unexpected scalar values: %#v", m)
			}
			if m["status"] != float64(tt.p.Status) {
				t.Fatalf("unexpected status: got %v want %d", m["status"], tt.p.Status)
			}

			rawErrors, ok := m["errors"].([]any)
			if !ok || len(rawErrors) != 1 {
				t.Fatalf("unexpected errors array: %#v", m["errors"])
			}
			err0, ok := rawErrors[0].(map[string]any)
			if !ok {
				t.Fatalf("unexpected error item type: %#v", rawErrors[0])
			}
			if err0["field"] != tt.p.Errors[0].Field || err0["code"] != string(tt.p.Errors[0].Code) || err0["message"] != tt.p.Errors[0].Message {
				t.Fatalf("unexpected error item values: %#v", err0)
			}
		})
	}
}

func TestProblem_FluentChain(t *testing.T) {
	p := New(400, CodeValidationError, "Validation failed").
		WithDetail("d").
		WithInstance("/x").
		WithType("https://errors.metaldocs.io/validation").
		WithFieldError("key", FieldCodeRequired, "msg")

	if p.Status != 400 || p.Code != CodeValidationError || p.Title != "Validation failed" {
		t.Fatalf("unexpected base problem: %+v", p)
	}
	if p.Detail != "d" || p.Instance != "/x" || p.Type != "https://errors.metaldocs.io/validation" {
		t.Fatalf("unexpected fluent values: %+v", p)
	}
	if len(p.Errors) != 1 {
		t.Fatalf("unexpected errors length: got %d want 1", len(p.Errors))
	}
	if p.Errors[0] != (FieldError{Field: "key", Code: FieldCodeRequired, Message: "msg"}) {
		t.Fatalf("unexpected field error: %+v", p.Errors[0])
	}
}

func TestProblem_FromValidation(t *testing.T) {
	fields := []FieldError{
		{Field: "a", Code: FieldCodeRequired, Message: "required"},
		{Field: "b", Code: FieldCodeInvalidFormat, Message: "invalid"},
		{Field: "c", Code: FieldCodeOutOfRange, Message: "range"},
	}

	p := FromValidation(fields)

	if p.Status != 400 || p.Code != CodeValidationError || p.Title != "Validation failed" {
		t.Fatalf("unexpected problem: %+v", p)
	}
	if len(p.Errors) != len(fields) {
		t.Fatalf("unexpected errors length: got %d want %d", len(p.Errors), len(fields))
	}
	for i := range fields {
		if p.Errors[i] != fields[i] {
			t.Fatalf("unexpected error at %d: got %+v want %+v", i, p.Errors[i], fields[i])
		}
	}
}

func TestProblem_Write(t *testing.T) {
	p := New(422, CodeValidationError, "Validation failed").
		WithDetail("bad payload").
		WithInstance("/v1/docs/1").
		WithType("https://errors.metaldocs.io/validation").
		WithFieldError("key", FieldCodeRequired, "required")

	rr := httptest.NewRecorder()
	if err := Write(rr, p); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	if got := rr.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("unexpected content-type: %q", got)
	}
	if rr.Code != p.Status {
		t.Fatalf("unexpected status: got %d want %d", rr.Code, p.Status)
	}

	var got Problem
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal body failed: %v", err)
	}
	if got.Type != p.Type || got.Title != p.Title || got.Status != p.Status || got.Detail != p.Detail || got.Instance != p.Instance || got.Code != p.Code {
		t.Fatalf("unexpected body scalars: got %+v want %+v", got, *p)
	}
	if len(got.Errors) != len(p.Errors) {
		t.Fatalf("unexpected body errors length: got %d want %d", len(got.Errors), len(p.Errors))
	}
	for i := range p.Errors {
		if got.Errors[i] != p.Errors[i] {
			t.Fatalf("unexpected body error at %d: got %+v want %+v", i, got.Errors[i], p.Errors[i])
		}
	}
}

func TestProblem_ErrorInterface(t *testing.T) {
	p := New(404, CodeNotFound, "Not found")

	var err error = p
	expected := "problem: 404 Not found (NOT_FOUND)"
	if err.Error() != expected {
		t.Fatalf("unexpected error string: got %q want %q", err.Error(), expected)
	}

	var target *Problem
	if !errors.As(err, &target) {
		t.Fatalf("errors.As failed")
	}
	if target != p {
		t.Fatalf("errors.As returned wrong target")
	}
}

func TestProblem_Write_HeaderSetBeforeStatus(t *testing.T) {
	p := New(400, CodeValidationError, "Validation failed")
	rr := httptest.NewRecorder()

	if err := Write(rr, p); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	if _, ok := rr.HeaderMap["Content-Type"]; !ok {
		t.Fatalf("content-type not present in header map")
	}
}

func TestProblem_New_PanicsOnInvalidStatus(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"below range", 99},
		{"above range", 600},
		{"zero", 0},
		{"negative", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("expected panic for status %d", tt.status)
				}
			}()
			New(tt.status, CodeInternalError, "test")
		})
	}
}
