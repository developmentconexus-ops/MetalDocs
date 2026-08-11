package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

// IdentityStatus is the pg_roles posture of the identity a *sql.DB connected
// as (current_user), as of the moment it was queried.
type IdentityStatus struct {
	RoleName  string
	Superuser bool
	BypassRLS bool
}

// Unsafe reports whether this identity defeats RLS policies and REVOKE-based
// hardening (e.g. the audit_events UPDATE/DELETE/TRUNCATE revoke in
// db/grants/0001_role_grants.sql). Postgres evaluates neither RLS nor GRANT/
// REVOKE against a role with rolsuper or rolbypassrls -- both make every
// FORCE ROW LEVEL SECURITY policy and every REVOKE in this schema a no-op for
// that connection (issue #88 / axis A6.1).
func (s IdentityStatus) Unsafe() bool {
	return s.Superuser || s.BypassRLS
}

// QueryIdentityStatus reads pg_roles for db's connected identity
// (current_user). It is the read half of the A6.1 boot assertion, kept
// separate from AssertSafeIdentity so callers that need the identity for a
// readiness probe (a DependencyCheck, not a boot-fatal) can reuse the same
// query instead of re-deriving pass/fail logic.
func QueryIdentityStatus(ctx context.Context, db *sql.DB) (IdentityStatus, error) {
	var status IdentityStatus
	err := db.QueryRowContext(ctx,
		`SELECT rolname, rolsuper, rolbypassrls FROM pg_roles WHERE rolname = current_user`,
	).Scan(&status.RoleName, &status.Superuser, &status.BypassRLS)
	if err != nil {
		return IdentityStatus{}, fmt.Errorf("query pg_roles for current_user: %w", err)
	}
	return status, nil
}

// AssertSafeIdentity is the A6.1 boot-fatal gate: it queries pg_roles for
// db's connected identity and returns a non-nil error when that identity is
// SUPERUSER or BYPASSRLS. The process must never reach a state where it can
// serve or process work under an identity that makes RLS and REVOKE-based
// hardening inert.
//
// Every composition root that opens the runtime database connection
// (metaldocs-api, metaldocs-worker, metaldocs-jobs; see
// internal/platform/bootstrap/{api,worker,jobs}.go) gets this enforced by
// calling OpenServing instead of Open -- OpenServing calls AssertSafeIdentity
// internally, so the gate is structural, not a second call a composition
// root can forget. AssertSafeIdentity itself stays exported because
// DBIdentityHealthCheck (bootstrap/api.go) and its tests re-run the same
// check post-boot, on a connection OpenServing already validated once, to
// catch a role attribute changed live (e.g. ALTER ROLE ... SUPERUSER) after
// the process started.
func AssertSafeIdentity(ctx context.Context, db *sql.DB) error {
	status, err := QueryIdentityStatus(ctx, db)
	if err != nil {
		return err
	}
	if status.Unsafe() {
		return fmt.Errorf(
			"refusing to boot: db identity %q is unsafe (rolsuper=%t rolbypassrls=%t) -- "+
				"RLS policies and REVOKE-based hardening (e.g. audit_events) are inert under "+
				"this connection; connect as a dedicated non-superuser, non-bypassrls role (A6.1)",
			status.RoleName, status.Superuser, status.BypassRLS,
		)
	}
	return nil
}
