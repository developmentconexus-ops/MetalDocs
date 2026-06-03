# ADR 0019 — `CapAuditRead` (Read-naming exception) and `CapSessionManage`

> **Status:** accepted 2026-06-02 (PR-2 of 12-PR IAM Admin Center rebuild)
> **Last verified:** 2026-06-02
> **Scope:** Registry additions ahead of PR-6 (audit trail handlers) and PR-7 (sessions & security handlers). Caps land **dormant** — no handler consumes them in this change.
> **Out of scope:** Tier-1 `(method,path)→cap` wiring (PR-6/PR-7); Tier-2 enforcement; Postgres tripwire attachment; FE consumers.
> **Key files:**
> - `internal/modules/iam/domain/model.go:42` — capability registry (`Capability` typed consts + `validCapabilities` map)
> - `internal/modules/iam/domain/model_test.go:36` — `TestCapabilityRegistrySize` (size-lock; bumped 27 → 28)
> - `db/migrations/0218_iam_caps_audit_session_pr2.sql` — `role_capabilities` grant migration
> - `db/reference-data/0001_product_reference_data.sql` — curated baseline mirror
> - `wiki/decisions/0016-view-grade-capabilities.md` — view/manage split precedent

## Context

The 12-PR IAM Admin Center rebuild needs two new capability constants up front so later PRs can wire handlers against a stable registry:

- **PR-6 — Audit trail.** Tenant-scoped admin needs to read the immutable governance event stream (`audit_events`). The registry already contains `CapAuditRead` (`audit.read`) seeded by migration 0189 for `system_admin` only. PR-6 will expose this surface to additional admin-ish roles, so the grant matrix must broaden first.
- **PR-7 — Sessions & security.** Tenant-scoped admin needs to force-logout / revoke active auth sessions. No capability exists today. Session **read** folds into `CapUserView` (session listing is a user-attribute query); only the write surface needs a new cap.

ADR 0016 established the View / Manage split as the default capability-naming pattern: every writable resource has a `*View` read-grade cap and a `*Manage` (or finer-grained `*Submit`, `*Approve`, …) write-grade cap. Audit events break that pattern — they are **read-only-by-design**. Governance records are immutable; there is no operator-facing "manage" surface, and there will not be one.

## Decision

1. **Keep `CapAuditRead` as named.** The `*Read` suffix (not `*View`) signals "this is the sole grade — no `*Manage` counterpart exists or will exist." This is the registry's naming convention for read-only-by-design resources. The const + string already exist; this ADR documents the exception so future authors do not "fix" the naming to match the View/Manage pattern.
2. **Add `CapSessionManage`** (`"session.manage"`). Standard View/Manage pattern would also imply a `CapSessionView`, but session read folds into `CapUserView` (session listing is a per-user query against the auth subsystem; no separate read surface is exposed). If a session-only read surface materializes later, a `CapSessionView` cap is additive and non-breaking.
3. **Broaden `audit.read` grants** beyond `system_admin` to roles with audit-of-the-trail needs. Migration 0210 adds the rows; reference-data baseline mirrors them. The grant set is policy, not derivable from code.

### Naming convention (registry-wide)

| Resource shape | Pattern | Examples |
|---|---|---|
| Writable resource with read + write grades | `*View` + `*Manage` (or finer verbs) | `CapUserView`/`CapUserManage`, `CapMembershipView`/`CapMembershipManage`, `CapTaxonomyView`/`CapTaxonomyManage` |
| Read-only-by-design resource (immutable, no operator manage surface) | `*Read` only | `CapAuditRead` |
| Write-only-by-design resource (no operator read surface, or read folds into another cap) | `*Manage` only | `CapSessionManage`, `CapRouteManage` |

## Role grant matrix (`role_capabilities`)

| Role | `audit.read` | `session.manage` |
|---|---|---|
| `system_admin` | ✓ (already seeded by 0189) | ✓ |
| `qms_admin` | ✓ (new) | — |
| `area_admin` | ✓ (new) | — |
| `approver` | ✓ (new) | — |
| `author` | — | — |
| `editor` | — | — |
| `signer` | — | — |
| `viewer` | — | — |

Net new rows: 4 (3× `audit.read` + 1× `session.manage`).

### Rationale per role

- **`system_admin`** — keeps both. Already holds `audit.read` (0189). Gains `session.manage` as the operator-of-last-resort for compromised sessions and forced logouts.
- **`qms_admin`** — the ISO/QMS compliance lead. Needs the audit trail to evidence controls during internal and external audits. Does **not** get `session.manage` (security-operations concern, not compliance).
- **`area_admin`** — needs audit visibility for events scoped to areas they administer (membership grants/revokes, document state transitions, route changes). Tenant-scoped admin observability is the whole point of this PR sequence.
- **`approver`** — needs audit on documents they approve to evidence signoff integrity (who signed what, when, with which cap asserted). Read-only; cannot mutate the trail.
- **`author` / `editor` / `signer` / `viewer`** — no operational need for audit; would create over-grant noise. They see their own document history via existing per-document timelines, not the raw governance stream.
- **`session.manage` restricted to `system_admin`** — force-logout / revoke is a security-sensitive write that should sit with the platform operator role, not the QMS/compliance/area-admin roles. Broaden later if product surfaces a tenant security-admin sub-role.

## Consequences

- **Positive** — PR-6 and PR-7 handler work can land against a stable registry; size-lock test (`TestCapabilityRegistrySize`) prevents accidental drift.
- **Positive** — naming convention (`*Read` vs `*View`/`*Manage`) is now documented; future authors can apply it consistently to other read-only-by-design resources (e.g. potential future `webhook.delivery.read`).
- **Positive** — grants land in both the forward migration and the curated baseline, so fresh-DB rebuilds match runtime.
- **Negative** — `validCapabilities` registry grows from 27 → 28. Trivial; test bump in same change.
- **Negative** — `qms_admin` is not in `domain.Role` (it is a DB-only role for `role_capabilities` purposes today). Granting `audit.read` to `qms_admin` is consistent with prior migrations (`0192_template_review_capability.sql` already seeds `qms_admin`). When `qms_admin` is promoted into `domain.Role`, this ADR's grant rows remain correct.
- **Dormant** — until PR-6/PR-7 land, no handler consults these caps. `CapabilityService.CanDo` will return true for the granted (role, cap) pairs, but no `(method, path)` Tier-1 row points at them yet. Verify only by SQL spot-check until PR-6.

## Alternatives Considered

| Option | Verdict | Reason |
|---|---|---|
| Rename `CapAuditRead` → `CapAuditView` for naming uniformity | Rejected | Audit is read-only-by-design; the `*View` suffix implies a `*Manage` counterpart that will never exist. Renaming would mislead future authors into thinking a `CapAuditManage` is missing. |
| Add `CapSessionView` alongside `CapSessionManage` | Rejected (for now) | Session listing folds into `CapUserView` (sessions are a per-user attribute). Adding `CapSessionView` would create two caps that grant identical access in PR-7. Additive later if a session-only read surface ships. |
| Grant `audit.read` to all roles by default | Rejected | Audit stream contains cross-tenant-admin actions (role grants, membership changes, force-logouts). `author`/`editor`/`signer`/`viewer` have no operational need and the noise would surface PII (actor identities) without a corresponding workflow. |
| Defer registry additions until PR-6/PR-7 actually need them | Rejected | The 12-PR sequence is staged precisely so registry, grants, and handlers can land independently. Bundling registry + handler in one PR causes large reverts when a handler review surfaces issues. |
| Single coarse `governance.manage` covering both audit-read and session-management | Rejected | Different audiences (compliance reads audit; security ops revokes sessions), different blast radius, different audit-log fidelity needs. |

## Verification

1. `go build ./...` — green.
2. `go test ./internal/modules/iam/domain/... -count=1` — green; `TestCapabilityRegistrySize` reflects 28.
3. Apply migration 0210 against local DB; SQL spot-check:
   ```sql
   SELECT role, capability
   FROM metaldocs.role_capabilities
   WHERE capability IN ('audit.read', 'session.manage')
   ORDER BY capability, role;
   ```
   Expect 5 rows: `audit.read` × {`approver`, `area_admin`, `qms_admin`, `system_admin`} + `session.manage` × {`system_admin`}.
4. `go test ./internal/modules/iam/... -count=1` — broader IAM suite still green.

## Rollback

Forward-only per `wiki/database/migration-policy.md`. To reverse:

- Revert `model.go` const + map entry (`CapSessionManage`) and bump `model_test.go` size back to 27.
- Author a new follow-on migration deleting the rows added by 0210 (data only, no schema change).
- Reference-data block in `db/reference-data/0001_product_reference_data.sql` removed in same change.

`CapAuditRead` is not rolled back here — it pre-exists this ADR (migration 0189).

## References

- `wiki/decisions/0016-view-grade-capabilities.md` — View/Manage split pattern this ADR carves an exception to.
- `wiki/decisions/0007-two-tier-authz.md` — two-tier authz model the caps will eventually be enforced under.
- `wiki/modules/iam.md` — module overview; §5.3 cap table updated in this PR.
- `migrations/0189_audit_capability_seed.sql` — original `audit.read` seed for `system_admin` (legacy historical chain).
- `db/migrations/0218_iam_caps_audit_session_pr2.sql` — PR-2 grant migration (this ADR; post-baseline forward tail).
- `db/reference-data/0001_product_reference_data.sql` — curated baseline; mirrors 0210.
