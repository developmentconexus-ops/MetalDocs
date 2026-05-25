package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrDocumentIDRequired = errors.New("search: document id required")
	ErrDocumentTitleEmpty = errors.New("search: document title required")
	ErrQueryTenantEmpty   = errors.New("search: tenant id required")
)

type Classification string

const (
	ClassificationPublic       Classification = "PUBLIC"
	ClassificationInternal     Classification = "INTERNAL"
	ClassificationConfidential Classification = "CONFIDENTIAL"
)

type Status string

const (
	StatusDraft    Status = "DRAFT"
	StatusActive   Status = "ACTIVE"
	StatusArchived Status = "ARCHIVED"
	StatusObsolete Status = "OBSOLETE"
)

type SubjectType string

const (
	SubjectTypeUser SubjectType = "user"
	SubjectTypeRole SubjectType = "role"
)

type Effect string

const (
	EffectAllow Effect = "allow"
	EffectDeny  Effect = "deny"
)

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

type Query struct {
	TenantID        string
	Text            string
	DocumentType    string
	DocumentProfile string
	DocumentFamily  string
	ProcessArea     string
	Subject         string
	OwnerID         string
	BusinessUnit    string
	Department      string
	Classification  Classification
	Status          Status
	Tag             string
	ExpiryBefore    *time.Time
	ExpiryAfter     *time.Time
	Limit           int
}

func NewQuery(query Query) (Query, error) {
	query.TenantID = strings.TrimSpace(query.TenantID)
	if query.TenantID == "" {
		return Query{}, ErrQueryTenantEmpty
	}
	return query, nil
}

type AccessPolicy struct {
	SubjectType   SubjectType
	SubjectID     string
	ResourceScope string
	ResourceID    string
	Capability    string
	Effect        Effect
}
