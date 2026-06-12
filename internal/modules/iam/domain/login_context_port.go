package domain

import "context"

// LoginContextPort is the narrow write surface the auth module calls after a
// successful login to record governance metadata on the iam_users row. Keeping
// this in iam/domain preserves the bounded-context boundary: auth drives the
// login flow, IAM owns iam_users, and the two modules meet at this interface
// rather than at a shared SQL dependency (F-06c, REQ-TOP-1).
//
// RecordLoginContext is best-effort — callers MUST swallow the error when login
// must not be blocked by a governance hint write. The implementation MUST
// treat a missing iam_users row as a no-op (no tenant binding yet).
type LoginContextPort interface {
	RecordLoginContext(ctx context.Context, userID, tenantID, ip, userAgent, deviceLabel string) error
}
