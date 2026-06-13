package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	auditdomain "metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/requesttrace"
)

// sessionRevoker is the narrow port satisfied structurally by
// *authpg.Repository. It is defined locally to avoid an import cycle between
// the iam/application and auth/infrastructure/postgres packages (H-3b Site 1).
// The auth postgres repo's RevokeSessionTx already preserves 0-rows →
// auth ErrSessionNotFound semantics.
type sessionRevoker interface {
	RevokeSessionTx(ctx context.Context, tx *sql.Tx, sessionID string, revokedAt time.Time) error
}

// SessionService owns the atomic revoke + audit write for admin-initiated
// session revocation (H-3b Site 1). The lookup (FindSession) remains in the
// delivery handler via the SessionAdmin port so the tenant guard runs before
// this service is called.
type SessionService struct {
	runner  db.TxRunner
	audit   auditdomain.Writer
	revoker sessionRevoker
}

// NewSessionService constructs a SessionService. All three arguments are
// required; callers MUST supply non-nil values.
func NewSessionService(runner db.TxRunner, audit auditdomain.Writer, revoker sessionRevoker) *SessionService {
	return &SessionService{runner: runner, audit: audit, revoker: revoker}
}

// RevokeSessionInfo carries the minimal session data the handler looked up
// before delegating the atomic revoke to this service.
type RevokeSessionInfo struct {
	SessionID string
	UserID    string
	TenantID  string
	Reason    string
}

// RevokeSession atomically revokes the session and appends the audit row in the
// same transaction. If the session no longer exists the underlying repo returns
// auth.ErrSessionNotFound (sentinel error), which propagates unchanged so the
// handler can return 404. A RecordTx failure causes the tx to roll back — the
// session is NOT revoked if the audit write fails.
//
// actorID is the authenticated caller who initiated the revocation.
func (s *SessionService) RevokeSession(ctx context.Context, info RevokeSessionInfo, actorID string) error {
	now := time.Now().UTC()
	payload := map[string]any{
		"session_id":    info.SessionID,
		"targetUserId": info.UserID,
	}
	if info.Reason != "" {
		payload["reason"] = info.Reason
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	ev := auditdomain.Event{
		ID:           "evt_" + uuid.NewString(),
		OccurredAt:   now,
		ActorID:      actorID,
		Action:       "auth.session.revoked",
		ResourceType: "session",
		ResourceID:   info.SessionID,
		PayloadJSON:  string(payloadJSON),
		TraceID:      requesttrace.Resolve(ctx),
		TenantID:     info.TenantID,
	}
	return s.runner.Do(ctx, func(tx *sql.Tx) error {
		if err := s.revoker.RevokeSessionTx(ctx, tx, info.SessionID, now); err != nil {
			return err
		}
		return s.audit.RecordTx(ctx, tx, ev)
	})
}
