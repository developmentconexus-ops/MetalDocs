// Package domain holds the tokens bounded-context entities and ports. Pure
// domain: no SQL, no HTTP. The token *grammar* is Node-owned
// (@metaldocs/shared-tokens); the name rule here is anti-corruption storage
// hygiene + membership only — see spec §7.
package domain

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

// ErrNotFound is returned when an entry does not exist for the tenant (or a
// cross-tenant id is requested — both resolve to 404, never 403).
var ErrNotFound = errors.New("tokens: entry not found")

// ErrImmutableName signals an attempt to change name on update. Maps to 422 with
// code immutable_field, distinct from generic validation.
var ErrImmutableName = errors.New("tokens: name is immutable")

// ValidationError is a friendly first-line field validation failure. The DB
// CHECK constraints are the enforcement; this is the app-side mirror.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string { return fmt.Sprintf("tokens: %s: %s", e.Field, e.Message) }

// nameRe is anti-corruption storage hygiene, NOT the token grammar. It mirrors
// the DB CHECK (name ~ '^[A-Za-z0-9_]+$'). The canonical charset (leading-char
// rule incl.) stays Node-owned and is enforced at the SP-3 UI edge.
var nameRe = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

const (
	maxName        = 64
	maxValue       = 4096
	maxLabel       = 256
	maxDescription = 1024
)

type Entry struct {
	ID          string
	TenantID    string
	Name        string
	Value       string
	Label       string
	Description *string
	CreatedBy   string
	UpdatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type NewEntryInput struct {
	TenantID    string
	ActorID     string
	Name        string
	Value       string
	Label       string
	Description *string
}

// NewEntry validates input and returns a populated Entry (ID/timestamps are set
// by the infrastructure layer / DB defaults). CreatedBy == UpdatedBy at create.
func NewEntry(in NewEntryInput) (*Entry, error) {
	if err := validateName(in.Name); err != nil {
		return nil, err
	}
	if err := validateValue(in.Value); err != nil {
		return nil, err
	}
	if err := validateLabel(in.Label); err != nil {
		return nil, err
	}
	if err := validateDescription(in.Description); err != nil {
		return nil, err
	}
	return &Entry{
		TenantID:    in.TenantID,
		Name:        in.Name,
		Value:       in.Value,
		Label:       in.Label,
		Description: in.Description,
		CreatedBy:   in.ActorID,
		UpdatedBy:   in.ActorID,
	}, nil
}

// ApplyUpdate mutates value/label/description on an existing entry after
// validation. name is immutable and is never touched here.
func (e *Entry) ApplyUpdate(value, label string, description *string, actorID string) error {
	if err := validateValue(value); err != nil {
		return err
	}
	if err := validateLabel(label); err != nil {
		return err
	}
	if err := validateDescription(description); err != nil {
		return err
	}
	e.Value = value
	e.Label = label
	e.Description = description
	e.UpdatedBy = actorID
	return nil
}

func validateName(name string) error {
	switch {
	case len(name) < 1 || len(name) > maxName:
		return &ValidationError{Field: "name", Message: "name must be 1-64 characters"}
	case !nameRe.MatchString(name):
		return &ValidationError{Field: "name", Message: "name must match [A-Za-z0-9_]"}
	}
	return nil
}

func validateValue(value string) error {
	if len(value) < 1 || len(value) > maxValue {
		return &ValidationError{Field: "value", Message: "value must be 1-4096 characters"}
	}
	return nil
}

func validateLabel(label string) error {
	if len(label) < 1 || len(label) > maxLabel {
		return &ValidationError{Field: "label", Message: "label must be 1-256 characters"}
	}
	return nil
}

func validateDescription(description *string) error {
	if description != nil && len(*description) > maxDescription {
		return &ValidationError{Field: "description", Message: "description must be <= 1024 characters"}
	}
	return nil
}
