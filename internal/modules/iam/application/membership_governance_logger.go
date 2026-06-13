package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	auditdomain "metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/requesttrace"
)

// AuditMembershipLogger implements MembershipGovernanceLogger by routing
// area-membership grant/revoke events into the canonical metaldocs.audit_events
// sink, inside the mutation transaction (REQ-ASYNC-1, T-007). It is the single
// writer of area_membership audit rows: the delivery layer no longer emits a
// post-commit duplicate (H-3a). Mirrors taxonomy's AuditGovernanceAdapter.
type AuditMembershipLogger struct {
	writer auditdomain.Writer
}

// NewAuditMembershipLogger constructs the adapter. A nil writer is a wiring
// bug, not a runtime condition — fail loud at startup.
func NewAuditMembershipLogger(writer auditdomain.Writer) *AuditMembershipLogger {
	if writer == nil {
		panic("AuditMembershipLogger: audit writer is required")
	}
	return &AuditMembershipLogger{writer: writer}
}

// LogTx writes the governance event inside tx so the audit record commits
// atomically with the membership mutation. tx must not be nil.
func (l *AuditMembershipLogger) LogTx(ctx context.Context, tx db.Tx, action string, membership domain.UserProcessArea) error {
	if tx == nil {
		return fmt.Errorf("iam: LogTx called with nil tx")
	}
	return l.writer.RecordTx(ctx, tx, l.buildEvent(ctx, action, membership))
}

func (l *AuditMembershipLogger) buildEvent(ctx context.Context, action string, membership domain.UserProcessArea) auditdomain.Event {
	payload, err := json.Marshal(map[string]any{
		"area_code": membership.AreaCode,
		"role":      string(membership.Role),
	})
	if err != nil {
		payload = []byte("{}")
	}
	grantedBy := ""
	if membership.GrantedBy != nil {
		grantedBy = *membership.GrantedBy
	}
	eventAction := "iam.area_membership.granted"
	if action == "role.revoke" {
		eventAction = "iam.area_membership.revoked"
	}
	return auditdomain.Event{
		ID:           uuid.NewString(),
		OccurredAt:   time.Now().UTC(),
		ActorID:      grantedBy,
		Action:       eventAction,
		ResourceType: "area_membership",
		ResourceID:   membership.UserID,
		PayloadJSON:  string(payload),
		// Recover the request trace the deleted post-commit handler used to
		// carry: the Recovery middleware seeds it into ctx from X-Trace-Id, so
		// the in-tx write loses no trace correlation.
		TraceID:  requesttrace.Resolve(ctx),
		TenantID: membership.TenantID,
	}
}
