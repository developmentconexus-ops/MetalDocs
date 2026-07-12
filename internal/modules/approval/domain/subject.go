package domain

import (
	"errors"
	"fmt"
)

// SubjectKind identifies which bounded-context entity an approval route or
// instance governs. Mirrors the DB CHECK constraint added by migration 0296
// on approval_routes.subject_kind / approval_instances.subject_kind exactly —
// only "document" and "template" are valid.
type SubjectKind string

// SubjectKind values understood by Subject.Validate.
const (
	SubjectKindDocument SubjectKind = "document"
	SubjectKindTemplate SubjectKind = "template"
)

// ErrInvalidSubjectKind is returned by Subject.Validate for any Kind other
// than SubjectKindDocument or SubjectKindTemplate.
var ErrInvalidSubjectKind = errors.New("approval: invalid subject kind")

// ErrEmptySubjectKey is returned by Subject.Validate when Key is empty.
var ErrEmptySubjectKey = errors.New("approval: subject key must not be empty")

// Subject generalizes what an approval route/instance governs beyond the
// legacy document-only model (M3 kernel extraction, ADR 0082). Kind names the
// governed entity type; Key is that entity's identifier within its own
// bounded context (e.g. a document_id or, in a future slice, a template id).
//
// This slice (P2.S2) only ever constructs document subjects — Key mirrors the
// legacy ProfileCode (Route) / DocumentID (Instance) column by construction,
// so every existing consumer of those legacy fields keeps working unchanged.
type Subject struct {
	Kind SubjectKind
	Key  string
}

// NewDocumentSubject builds a Subject for the document bounded context, where
// key is the route's profile_code or the instance's document_id depending on
// the caller.
func NewDocumentSubject(key string) Subject {
	return Subject{Kind: SubjectKindDocument, Key: key}
}

// Validate returns ErrInvalidSubjectKind unless Kind is exactly
// SubjectKindDocument or SubjectKindTemplate, and ErrEmptySubjectKey if Key
// is empty.
func (s Subject) Validate() error {
	switch s.Kind {
	case SubjectKindDocument, SubjectKindTemplate:
	default:
		return ErrInvalidSubjectKind
	}
	if s.Key == "" {
		return ErrEmptySubjectKey
	}
	return nil
}

// String renders Subject as "<kind>:<key>", useful for logging and error
// messages.
func (s Subject) String() string {
	return fmt.Sprintf("%s:%s", s.Kind, s.Key)
}

// Equal reports whether s and other have identical Kind and Key.
func (s Subject) Equal(other Subject) bool {
	return s.Kind == other.Kind && s.Key == other.Key
}
