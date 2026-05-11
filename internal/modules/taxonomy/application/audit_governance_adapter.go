package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	auditdomain "metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/modules/taxonomy/domain"
)

// AuditGovernanceAdapter implements domain.GovernanceLogger by routing events
// to the canonical metaldocs.audit_events sink.
type AuditGovernanceAdapter struct {
	writer auditdomain.Writer
}

func NewAuditGovernanceAdapter(w auditdomain.Writer) *AuditGovernanceAdapter {
	return &AuditGovernanceAdapter{writer: w}
}

func (a *AuditGovernanceAdapter) Log(ctx context.Context, event domain.GovernanceEvent) error {
	payload := event.PayloadJSON
	if payload == nil {
		payload, _ = json.Marshal(map[string]string{})
	}
	return a.writer.Record(ctx, auditdomain.Event{
		ID:           uuid.NewString(),
		OccurredAt:   time.Now().UTC(),
		ActorID:      event.ActorUserID,
		Action:       event.EventType,
		ResourceType: event.ResourceType,
		ResourceID:   event.ResourceID,
		PayloadJSON:  string(payload),
		TenantID:     event.TenantID,
	})
}
