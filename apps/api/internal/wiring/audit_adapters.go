package wiring

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"

	auditdomain "metaldocs/internal/modules/audit/domain"
	docapp "metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/requesttrace"
)

// traceIDFromContext resolves a trace ID from the request context.
func traceIDFromContext(ctx context.Context) string {
	return requesttrace.Resolve(ctx)
}

// bypassAuditAdapter adapts the audit Writer to authz.BypassAuditSink so the
// low-level authz package can record tier-2 bypasses without importing the audit
// module (ADR 0022 Phase 11, F8). It writes in the caller's tx (RecordTx) at the
// same fidelity/atomicity as the in-tx normal-grant audit.
type bypassAuditAdapter struct {
	writer auditdomain.Writer
}

// NewBypassAuditSink constructs an authz.BypassAuditSink backed by the given audit writer.
func NewBypassAuditSink(writer auditdomain.Writer) authz.BypassAuditSink {
	if writer == nil {
		panic("bypass audit writer is nil")
	}
	return &bypassAuditAdapter{writer: writer}
}

func (a *bypassAuditAdapter) RecordBypass(ctx context.Context, tx *sql.Tx, ev authz.BypassEvent) error {
	payload, err := json.Marshal(map[string]any{
		"kind":       string(ev.Kind),
		"capability": ev.Capability,
		"area_code":  ev.AreaCode,
	})
	if err != nil {
		payload = []byte("{}")
	}
	actor := ev.ActorID
	if actor == "" {
		actor = "system"
	}
	resourceID := ev.Capability
	if resourceID == "" {
		resourceID = string(ev.Kind)
	}
	return a.writer.RecordTx(ctx, tx, auditdomain.Event{
		ID:           "evt_" + uuid.NewString(),
		OccurredAt:   time.Now().UTC(),
		ActorID:      actor,
		Action:       "authz.bypass." + string(ev.Kind),
		ResourceType: "authz_bypass",
		ResourceID:   resourceID,
		PayloadJSON:  string(payload),
		TraceID:      traceIDFromContext(ctx),
		TenantID:     ev.TenantID, // "" allowed for cross-tenant background sweeps
	})
}

// documentsAuditAdapter adapts the audit Writer to the docapp.Audit interface.
type documentsAuditAdapter struct {
	writer auditdomain.Writer
}

// NewDocumentsAuditSink constructs a docapp.Audit backed by the given audit writer.
func NewDocumentsAuditSink(writer auditdomain.Writer) docapp.Audit {
	if writer == nil {
		panic("documents audit writer is nil")
	}
	return &documentsAuditAdapter{writer: writer}
}

func (a *documentsAuditAdapter) WriteTx(ctx context.Context, tx db.Tx, tenantID, actorID, action, docID string, meta any) error {
	payload := map[string]any{"tenant_id": tenantID}
	if meta != nil {
		payload["meta"] = meta
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		raw = []byte("{}")
	}
	return a.writer.RecordTx(ctx, tx, auditdomain.Event{
		ID:           uuid.NewString(),
		OccurredAt:   time.Now().UTC(),
		ActorID:      actorID,
		Action:       action,
		ResourceType: "document",
		ResourceID:   docID,
		PayloadJSON:  string(raw),
		TraceID:      traceIDFromContext(ctx),
		TenantID:     tenantID,
	})
}

func (a *documentsAuditAdapter) Write(ctx context.Context, tenantID, actorID, action, docID string, meta any) {
	payload := map[string]any{"tenant_id": tenantID}
	if meta != nil {
		payload["meta"] = meta
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		raw = []byte("{}")
	}

	if err := a.writer.Record(ctx, auditdomain.Event{
		ID:           uuid.NewString(),
		OccurredAt:   time.Now().UTC(),
		ActorID:      actorID,
		Action:       action,
		ResourceType: "document",
		ResourceID:   docID,
		PayloadJSON:  string(raw),
		TraceID:      traceIDFromContext(ctx),
		TenantID:     tenantID,
	}); err != nil {
		slog.Error("documents audit write failed", "err", err)
	}
}
