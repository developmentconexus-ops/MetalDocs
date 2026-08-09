package domain

import (
	"errors"
	"strings"
	"time"
)

// Sentinel errors returned by NewDocument and NewQuery validation.
var (
	ErrDocumentIDRequired = errors.New("search: document id required")
	ErrDocumentTitleEmpty = errors.New("search: document title required")
	ErrQueryTenantEmpty   = errors.New("search: tenant id required")
)

// Classification is a document's sensitivity level.
type Classification string

// Classification values. Legacy field, no longer populated by the v2 reader
// (see SearchDocumentResponse doc comment in delivery/http/handler.go).
const (
	ClassificationPublic       Classification = "PUBLIC"
	ClassificationInternal     Classification = "INTERNAL"
	ClassificationConfidential Classification = "CONFIDENTIAL"
)

// Status is a document's lifecycle state.
type Status string

// Status values, matching the documents module's lifecycle states.
const (
	StatusDraft    Status = "DRAFT"
	StatusActive   Status = "ACTIVE"
	StatusArchived Status = "ARCHIVED"
	StatusObsolete Status = "OBSOLETE"
)

// Document is a single search result row.
type Document struct {
	ID               string
	Title            string
	DocumentType     string
	DocumentProfile  string
	DocumentFamily   string
	DocumentSequence int
	DocumentCode     string
	ProcessArea      string
	Subject          string
	OwnerID          string
	BusinessUnit     string
	Department       string
	Classification   Classification
	Status           Status
	Tags             []string
	EffectiveAt      *time.Time
	ExpiryAt         *time.Time
	CreatedAt        time.Time
}

// NewDocument validates doc, trimming ID and Title and rejecting either if
// empty.
func NewDocument(doc Document) (Document, error) {
	doc.ID = strings.TrimSpace(doc.ID)
	doc.Title = strings.TrimSpace(doc.Title)
	if doc.ID == "" {
		return Document{}, ErrDocumentIDRequired
	}
	if doc.Title == "" {
		return Document{}, ErrDocumentTitleEmpty
	}
	return doc, nil
}

// Query is the set of filters and pagination inputs for SearchDocuments.
type Query struct {
	TenantID string
	// ActorUserID is the authenticated caller. Per-document visibility is
	// enforced against this actor at the data layer (unified role/area model +
	// controlled_document_{area,user}_grants). Empty actor → no results.
	ActorUserID     string
	Text            string
	DocumentType    string
	DocumentProfile string
	DocumentFamily  string
	ProcessArea     string
	OwnerID         string
	Department      string
	Status          Status
	ExpiryBefore    *time.Time
	ExpiryAfter     *time.Time
	Limit           int
}

// NewQuery validates query, trimming TenantID and ActorUserID and rejecting
// an empty TenantID.
func NewQuery(query Query) (Query, error) {
	query.TenantID = strings.TrimSpace(query.TenantID)
	if query.TenantID == "" {
		return Query{}, ErrQueryTenantEmpty
	}
	query.ActorUserID = strings.TrimSpace(query.ActorUserID)
	return query, nil
}
