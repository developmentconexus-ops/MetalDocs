# ADR 0055 — Global auth identities, tenant-scoped sessions (binding rules)

- **Status:** Accepted
- **Last verified:** 2026-07-02
- **Date:** 2026-07-02
- **Scope:** Converts the narrative "by design" finding of ADR 0027 §1 (`auth_identities` has no `tenant_id`) into binding, numbered operational rules for how identity and session tenancy interact at runtime: session minting, session resolution, identity lookups, and cross-tenant probe behavior. Does not reopen or restate ADR 0027's RLS-sequencing content. Closes tech-debt item T-008 (identities portion) and grade-A register item SEC-10 / DEC-03.
- **Depends on:** ADR 0027 (`auth_identities` tenant-global by design; RLS sequencing) and the pooled multi-tenant invariant (every tenant table carries `tenant_id`; cross-tenant URL → 404).

---

## Context

ADR 0027 already decided and executed the schema question: `auth_identities` (`db/baseline/0001_current_schema.sql:966-982`) carries no `tenant_id` column, by deliberate design — one human, one credential set (`user_id`, `username`, `password_hash`, lockout/failure metadata), potentially many tenant memberships via `iam_users.tenant_id`. That decision is unchanged and is not relitigated here.

What ADR 0027 does not do is state, as binding rules, how tenancy is actually enforced given that identity is global. The grade-A register (SEC-10 / DEC-03, `wiki/reviews/grade-a-simplification-report-2026-07-01.md:58,197`) flagged this gap: "Decide: ADR affirming global identity + tenant-scoped sessions, or add column. Record either way." The register's own framing of ADR 0027 as an "anomaly" reflects that 0027's Decision section narrates the schema choice but does not bind the session/lookup/probe behavior that makes the global-identity model safe in a pooled multi-tenant deployment. This ADR supplies those binding rules, verified against current runtime code.

### Verified runtime facts

- **`auth_identities` has no `tenant_id`.** `db/baseline/0001_current_schema.sql:966-982` — columns are `user_id, username, email, password_hash, password_algo, must_change_password, last_login_at, failed_login_attempts, locked_until, created_at, updated_at, display_name, is_active, last_failed_login_at, last_failed_login_ip`. No tenant column. Unchanged since ADR 0027.
- **`auth_sessions` carries `tenant_id`.** `db/baseline/0001_current_schema.sql:989-996` (columns), `:2537-2538` (PK on `session_id`), `:4045-4049` (`fk_auth_sessions_tenant` FK to `metaldocs.tenants(id)`), `:3246-3249` (`idx_auth_sessions_tenant_user` on `(tenant_id, user_id)`). RLS is enabled and forced on `auth_sessions` (`:1001` `FORCE ROW LEVEL SECURITY`; `:4489` policy). Added by migration `0184_auth_sessions_tenant_id.sql` per the auth tech-debt register.
- **Tenant is fixed at login and never re-derived from identity afterward.** `internal/modules/auth/application/service.go:344-366` (`resolveLoginTenant`) resolves the tenant for a new session strictly from `iam_users`-derived membership (`s.repo.GetUserTenants`) — a claimed tenant must be in the caller's membership set (`ErrTenantNotPermitted` otherwise, line 356) or, absent a claim, the session gets the caller's sole tenant (line 359-361) or fails closed (`ErrTenantClaimRequired`, line 365). The resolved tenant is persisted onto the `auth_sessions` row at creation (`repository.go:77-80`, `INSERT ... tenant_id`).
- **Session resolution reads tenancy off the session row, not the identity.** `internal/modules/auth/application/service.go:368-400` (`ResolveSession`) loads the session via `FindSession` (`internal/modules/auth/infrastructure/postgres/repository.go:87-104`, `SELECT ... tenant_id ... FROM metaldocs.auth_sessions WHERE session_id = $1`) and threads `session.TenantID` into `buildCurrentUser` (service.go:399) for every subsequent request on that session. `auth_identities` is never queried for tenant context on the request path.

## Decision

**Auth identity records are deliberately tenant-global; tenancy is enforced at the session and membership layer, not the identity layer.** This is the standard multi-tenant SSO shape: one identity, N tenant memberships, one tenant per active session. The following rules are binding:

1. **Sessions MUST be tenant-scoped.** Every `auth_sessions` row carries a non-null `tenant_id` fixed at mint time by `resolveLoginTenant` (`service.go:344-366`) from the caller's verified `iam_users` membership set — never from a caller-supplied claim alone, never inherited implicitly. `ResolveSession` (`service.go:368-400`) and `FindSession` (`repository.go:87-104`) read tenancy from the session row for every request on that session; `auth_identities` MUST NOT be queried for tenant context on any request path.
2. **Identity lookups never bypass tenant checks on the resources they unlock.** A resolved `user_id` from `auth_identities` grants no tenant-scoped access by itself. Every downstream authorization decision (tier-1 route capability, tier-2 `authz.Require`, DB tripwire) operates on the session's `tenant_id`, matching ADR 0022's capability model. Code that joins `auth_identities` to a tenant-scoped table MUST do so via `iam_users`/`auth_sessions`, never by assuming a 1:1 identity-to-tenant relationship.
3. **Cross-tenant probes still 404.** A session minted for tenant A MUST NOT resolve, read, or act on tenant B's resources regardless of the fact that the underlying identity is shared infrastructure. This is unchanged from the pooled multi-tenant invariant; the global identity table does not create an exception to it, because no request path derives tenant scope from `auth_identities`.
4. **Any future per-tenant identity requirement is a new ADR.** If a product requirement ever needs one identity to carry per-tenant-distinct credentials (e.g. different passwords per tenant, tenant-specific lockout policy), that is a schema change to `auth_identities` and MUST be proposed as a new ADR, not patched in as a silent column addition. This ADR and ADR 0027 remain the record of the current by-design choice until superseded.

## Consequences

- T-008 (`wiki/modules/auth-tech-debt.md`) is closed as by-design for both the sessions portion (already closed by migration `0184` per the existing row) and the identities portion (closed by this ADR + ADR 0027).
- SEC-10 / DEC-03 (`wiki/reviews/grade-a-simplification-report-2026-07-01.md`) is resolved: the requested ADR now exists with binding rules, no schema change made.
- Reviewers checking auth code for tenancy violations should verify against rules 1-3 above (session-row tenancy, no identity-layer tenant bypass, cross-tenant 404) rather than expecting a `tenant_id` column on `auth_identities`.
- No migration, schema change, or code change is required by this ADR. It documents and binds existing, verified runtime behavior.

## References

- ADR [`0027-rls-adoption-sequencing.md`](0027-rls-adoption-sequencing.md) — origin decision that `auth_identities` has no `tenant_id`, by design; RLS sequencing.
- ADR [`0022-authz-capability-coherence.md`](0022-authz-capability-coherence.md) — capability-based authz model that all session-derived tenant context feeds into.
- `wiki/modules/auth-tech-debt.md` T-008 — tech-debt row closed by this ADR.
- `wiki/reviews/grade-a-simplification-report-2026-07-01.md` SEC-10 / DEC-03 — register item this ADR resolves.
- `wiki/modules/auth/_artifacts/05-industry.md` IP-008 — industry pattern citation (row-level `tenant_id` on multi-tenant tables) that originally flagged the gap.
