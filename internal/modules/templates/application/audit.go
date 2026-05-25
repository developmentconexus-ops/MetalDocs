package application

import (
	"time"

	"metaldocs/internal/modules/templates/domain"
)

func newAuditEvent(tenantID, templateID, actorID string, versionID *string, action domain.AuditAction, details map[string]any, occurredAt time.Time) (*domain.AuditEvent, error) {
	event, err := domain.NewAuditEvent(tenantID, templateID, actorID, versionID, action, details, occurredAt)
	if err != nil {
		return nil, err
	}
	return &event, nil
}
