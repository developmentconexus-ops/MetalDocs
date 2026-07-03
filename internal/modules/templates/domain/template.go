// Package domain holds the templates bounded context's core model: templates,
// their versioned lifecycle (draft → under_review → approved → published →
// obsolete), placeholder/metadata schemas, approval configuration with ISO
// segregation-of-duties rules, audit events, and the sentinel errors shared by
// the application and delivery layers.
package domain

import (
	"errors"
	"time"
)

// Template is the aggregate root for a document template: a tenant-scoped,
// named container identified by a doc-type code and key, tracking its latest
// and published version/revision pointers and archival state.
type Template struct {
	ID                     string
	TenantID               string
	DocTypeCode            string
	Key                    string
	Name                   string
	Description            string
	LatestVersion          int
	LatestRevisionNumber   int
	PublishedVersionID     *string
	PublishedVersionNumber *int
	CurrentRevisionNumber  *int
	CreatedBy              string
	SystemOwned            bool
	CreatedAt              time.Time
	UpdatedAt              time.Time
	ArchivedAt             *time.Time
}

// IsArchived reports whether the template has been archived.
func (t *Template) IsArchived() bool { return t.ArchivedAt != nil }

var (
	// ErrNotFound is returned when a template or version does not exist in
	// the caller's tenant scope.
	ErrNotFound = errors.New("templates: not_found")
	// ErrKeyConflict is returned when a template key already exists for the
	// tenant and doc type.
	ErrKeyConflict = errors.New("templates: key_conflict")
	// ErrArchived is returned when a write is attempted on an archived
	// template.
	ErrArchived = errors.New("templates: archived")
	// ErrSystemTemplateImmutable is returned when a mutation targets a
	// system-owned template.
	ErrSystemTemplateImmutable = errors.New("templates: system_template_immutable")
)
