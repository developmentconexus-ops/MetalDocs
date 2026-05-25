package domain

import "time"

type Classification string

const (
	ClassificationPublic       Classification = "PUBLIC"
	ClassificationInternal     Classification = "INTERNAL"
	ClassificationConfidential Classification = "CONFIDENTIAL"
)

type DocStatus string

const (
	DocStatusDraft    DocStatus = "DRAFT"
	DocStatusActive   DocStatus = "ACTIVE"
	DocStatusArchived DocStatus = "ARCHIVED"
	DocStatusObsolete DocStatus = "OBSOLETE"
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
	Status           DocStatus
	Tags             []string
	EffectiveAt      *time.Time
	ExpiryAt         *time.Time
	CreatedAt        time.Time
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
	Status          DocStatus
	Tag             string
	ExpiryBefore    *time.Time
	ExpiryAfter     *time.Time
	Limit           int
}

type AccessPolicy struct {
	SubjectType   SubjectType
	SubjectID     string
	ResourceScope string
	ResourceID    string
	Capability    string
	Effect        Effect
}
