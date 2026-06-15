# ADR 0029 — `UserDisplayNameReader`: cross-module display-name reads go through an iam-owned port

> **Status:** Accepted 2026-06-15
> **Last verified:** 2026-06-15
> **Scope:** How modules other than iam obtain `metaldocs.iam_users.display_name`. The owning module (iam), the port shape (single + batch by `(tenantID, userID)`), and the reads-live / off-tx (H-PRE-1) constraint.
> **Out of scope:** iam's own intra-module reads of `iam_users` (presence, auth); the `password_hash` reauth read (already behind `IamUserReader`); the `security` tenant-scope JOINs on `iam_users` (a different concern — bounded defer, see Consequences); the two-tier authz model (ADR 0022).
> **Key files:**
> - `internal/modules/iam/domain/user_display_name_port.go` — the owned port (`DisplayName` + `DisplayNames`, `NoopUserDisplayNameReader` null-object)
> - `internal/modules/iam/infrastructure/postgres/user_display_name_repository.go` — pool-backed impl (the only `iam_users` display-name SQL outside iam consumers)
> - `internal/modules/documents/repository/repository.go` — created-by display-name consumer (was raw `iam_users` SQL at :134)
> - `internal/modules/documents/approval/http/get_instance_handler.go` — batch display-name consumer (was raw `COALESCE(display_name,user_id)` `ANY($2)`)
> - `internal/modules/documents/approval/repository/postgres_approval_repository.go` — F1.3's contained `LoadActorDisplayName`, generalized onto the shared port
> - `internal/modules/iam/delivery/http/sessions_handler.go` — session-list display-name consumer (M4/F4.4; auth's `ListActiveSessions` `iam_users` JOIN removed, names resolved here via `DisplayNames`)
> - `internal/modules/security/infrastructure/postgres/repository.go` — lockout / failed-login / new-device display-name consumer (M4/F4.6; the 3 security `iam_users.display_name` JOINs removed, names resolved via `DisplayNames` + `missing→user_id` fallback; tenant scope via `TenantUserReader`, ADR 0031)

## Context

`metaldocs.iam_users` is owned by the iam module. Three cross-module sites read its `display_name`
column with raw SQL — `documents/repository/repository.go:134` (created-by), `approval/http/get_instance_handler.go:127`
(batch, `COALESCE(display_name,user_id)`), and F1.3's contained `ApprovalRepository.LoadActorDisplayName`
(`decision_service.go:163`). This is the **H-G class** defect: *cross-module reach-without-a-port* — a
consumer module issuing SQL against another module's owned table, coupling to its physical schema and
bypassing the owning module's invariants. F1.3 fixed the signoff instance in a contained way; M4/F4.1
generalizes the class.

The signoff display-name read additionally sits near a lock-holding signoff transaction. **H-PRE-1**
(advisory-lock deadlock constraint) forbids an authz-recording read on a fresh connection inside a
lock-holding atomic tx; F1.3 already hoisted that read off-tx and it must not regress.

## Decision

Cross-module display-name access flows through a single **iam-owned domain port**,
`iam/domain.UserDisplayNameReader`, with a single-read `DisplayName(ctx, tenantID, userID) (string, error)`
and a batch `DisplayNames(ctx, tenantID, userIDs) (map[string]string, error)`. The implementation
(`iam/infrastructure/postgres.userDisplayNameRepository`) is the **single owning adapter** over
`iam_users` display-name reads and runs on the connection pool — never on a caller's lock-holding tx
(H-PRE-1). The three named consumers depend on the interface, never on `iam_users`. Semantics are read
from the existing consumers (consumer-driven): single-read returns `("", nil)` when absent (best-effort
snapshot); batch omits absent/empty users so the caller reproduces its own
`COALESCE(NULLIF(display_name,''), user_id)` fallback. A `NoopUserDisplayNameReader` null-object serves
paths that need no resolution.

Reads stay **live** — no snapshot/denormalization of the name into consumer tables (design
D4/Approach-3; "freeze actor name" / Approach 2 was explicitly rejected absent a separate audit/legal
product decision).

## Consequences

- **0 `iam_users` display-name SQL outside iam/** in the swept surface — the H-G reach is closed at the
  *class* level (grep-verified), not one instance patched.
- Consumers couple to a stable Go interface; iam can change `iam_users` physical layout without breaking
  documents/approval.
- Display-name truth is always current (live read); no staleness, no backfill migration, no second
  source to keep in sync.
- The signoff display-name read stays off the lock-holding tx — H-PRE-1 preserved (runtime/`pg_locks`
  evidence in F4.1).
- **⚠ Census correction (2026-06-15, post-F4.1):** the original claim here that
  `security/infrastructure/postgres/repository.go` JOINs `iam_users` *only* for tenant-scoping was
  **wrong** — `ListLockouts`, `CountRecentFailedLoginsByUser`, and `ListNewDeviceLogins` DO read
  `display_name`. Under operator Option-2 (full close) these were migrated in **M4/F4.6** onto this
  port (display names) + the new iam tenant-scope/membership port (the `auth_identities`-coupled tenant
  scope) — **DONE**, not deferred. See `f4.6-security-display-name-port/` and `f4.5-iam-tenant-membership-port/`.
  The membership port deferred here is now **built** — see [`0031-tenant-user-reader-port.md`](0031-tenant-user-reader-port.md) (`TenantUserReader`, M4/F4.5), which supersedes this note.
- **Class closed at true zero (post-F4.6):** **0** cross-module `iam_users.display_name` reads remain
  outside `iam/` anywhere in the tree (grep-verified across `auth` + `security`). F4.6 also dropped
  `CountRecentLockouts`' tenant-scope `iam_users` JOIN via the membership port. The **only** remaining
  cross-module `iam_users` read is security's `MfaCoverage` aggregate (`iam_users.mfa_enabled` +
  `iam_user_roles`) — a true iam-owned *aggregate over its own data*, not a display-name reach; it
  stays a bounded defer (*Trigger:* M5 re-audit or next structural touch; *Owner:* backend). Intra-module
  reads (`iam/presence/*`) and the `password_hash` reauth read (existing `IamUserReader`) are out of class.

## References
- Feature F4.1 — `docs/superpowers/milestones/grade-a-architecture-remediation/milestone-4-systemic-ports/f4.1-user-display-name-reader/spec.md`
- Governing spec — `docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md` §5.2 (H-G class), §M4
- Sibling port ADR [`0030-template-version-state-port.md`](0030-template-version-state-port.md)
- H-PRE-1 — advisory-lock deadlock constraint (never authz-recording read inside a lock-holding tx)
