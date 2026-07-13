package domain

import (
	"errors"
	"time"
)

// AuditAction identifies the kind of template/template-version lifecycle
// event an AuditEvent records.
type AuditAction string

const (
	// AuditCreated marks that a template was created.
	AuditCreated AuditAction = "created"
	// AuditSaved marks that a template version draft was saved.
	AuditSaved AuditAction = "saved"
	// AuditSubmitted marks that a template version was submitted for review/approval.
	AuditSubmitted AuditAction = "submitted"
	// AuditReviewed marks that a template version was reviewed.
	AuditReviewed AuditAction = "reviewed"
	// AuditApproved marks that a template version was approved.
	AuditApproved AuditAction = "approved"
	// AuditRejected marks that a template version was rejected.
	AuditRejected AuditAction = "rejected"
	// AuditPublished marks that a template version was published.
	AuditPublished AuditAction = "published"
	// AuditObsoleted marks that a template version was made obsolete.
	AuditObsoleted AuditAction = "obsoleted"
	// AuditArchived marks that a template was archived.
	AuditArchived AuditAction = "archived"
	// AuditRestored marks that an archived template was restored.
	AuditRestored AuditAction = "restored"
)

// AuditEvent records a single lifecycle action taken against a template or
// template version, for the templates audit trail.
type AuditEvent struct {
	TenantID   string
	TemplateID string
	VersionID  *string
	ActorID    string
	Action     AuditAction
	Details    map[string]any
	OccurredAt time.Time
}

// NewAuditEvent creates a new AuditEvent, requiring non-empty tenantID,
// templateID, and actorID.
func NewAuditEvent(tenantID, templateID, actorID string, versionID *string, action AuditAction, details map[string]any, occurredAt time.Time) (AuditEvent, error) {
	if tenantID == "" {
		return AuditEvent{}, errors.New("audit event: tenantID required")
	}
	if templateID == "" {
		return AuditEvent{}, errors.New("audit event: templateID required")
	}
	if actorID == "" {
		return AuditEvent{}, errors.New("audit event: actorID required")
	}
	return AuditEvent{
		TenantID:   tenantID,
		TemplateID: templateID,
		VersionID:  versionID,
		ActorID:    actorID,
		Action:     action,
		Details:    details,
		OccurredAt: occurredAt,
	}, nil
}
