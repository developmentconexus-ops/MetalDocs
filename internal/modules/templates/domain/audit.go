package domain

import (
	"errors"
	"time"
)

type AuditAction string

const (
	AuditCreated               AuditAction = "created"
	AuditSaved                 AuditAction = "saved"
	AuditSubmitted             AuditAction = "submitted"
	AuditReviewed              AuditAction = "reviewed"
	AuditApproved              AuditAction = "approved"
	AuditRejected              AuditAction = "rejected"
	AuditPublished             AuditAction = "published"
	AuditObsoleted             AuditAction = "obsoleted"
	AuditArchived              AuditAction = "archived"
	AuditRestored              AuditAction = "restored"
	AuditApprovalConfigUpdated AuditAction = "approval_config_updated"
)

type AuditEvent struct {
	TenantID   string
	TemplateID string
	VersionID  *string
	ActorID    string
	Action     AuditAction
	Details    map[string]any
	OccurredAt time.Time
}

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
