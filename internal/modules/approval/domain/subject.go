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

// NewTemplateSubject builds a Subject for the template bounded context (M3
// P3.S2b-3b-ii, ADR 0082 template extension). key is the template version id
// for an Instance (the artifact under approval) — symmetric with
// NewDocumentSubject, whose key is the analogous per-call identifier. A
// template ROUTE's subject key is the template id (the governance selector),
// set directly by the caller since route creation for templates is a later
// slice (#19); this constructor is only used for the two-level keying's
// instance/artifact side.
func NewTemplateSubject(key string) Subject {
	return Subject{Kind: SubjectKindTemplate, Key: key}
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

// TemplateInboxMeta (Unit 4.2) carries the display metadata the worklist/inbox
// read-model needs for a template-subject approval instance: the owning
// template's id (TemplateID — the FE navigation target, NOT the version id)
// and its human name (Title). It lives in domain (a leaf package) so the
// approval-owned TemplateVersionReader port can return it while the
// templates-side adapter implements it, without templates/infrastructure
// having to import approval/application — that edge would cycle with the
// existing approval test → templates/infrastructure import.
type TemplateInboxMeta struct {
	TemplateID string
	Title      string
}
