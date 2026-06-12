package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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
		var err error
		payload, err = json.Marshal(map[string]string{})
		if err != nil {
			return fmt.Errorf("marshal empty governance payload: %w", err)
		}
	}
	return a.writer.Record(ctx, auditdomain.Event{
		ID:           uuid.NewString(),
		OccurredAt:   time.Now().UTC(),
		ActorID:      event.ActorUserID,
		Action:       string(event.EventType),
		ResourceType: event.ResourceType,
		ResourceID:   event.ResourceID,
		PayloadJSON:  string(payload),
		TenantID:     event.TenantID,
	})
}

// LogTx writes the governance event inside tx so the audit record is
// atomically committed with the mutation that caused it (REQ-ASYNC-1, F-07).
// tx must not be nil; callers must ensure a real *sql.Tx is provided.
func (a *AuditGovernanceAdapter) LogTx(ctx context.Context, tx *sql.Tx, event domain.GovernanceEvent) error {
	if tx == nil {
		return fmt.Errorf("taxonomy: LogTx called with nil tx; adapter requires a real *sql.Tx (implement Unwrap on the tx type)")
	}
	payload := event.PayloadJSON
	if payload == nil {
		var err error
		payload, err = json.Marshal(map[string]string{})
		if err != nil {
			return fmt.Errorf("marshal empty governance payload: %w", err)
		}
	}
	auditEvent := auditdomain.Event{
		ID:           uuid.NewString(),
		OccurredAt:   time.Now().UTC(),
		ActorID:      event.ActorUserID,
		Action:       string(event.EventType),
		ResourceType: event.ResourceType,
		ResourceID:   event.ResourceID,
		PayloadJSON:  string(payload),
		TenantID:     event.TenantID,
	}
	return a.writer.RecordTx(ctx, tx, auditEvent)
}
