# ADR 0016 — View-Grade Capabilities for IAM, Membership, Taxonomy, Metrics

> **Status:** accepted 2026-06-01 (security review HIGH downgraded to LOW after handler scope verification — see §Security Boundary Notes; grant matrix product-confirmed 2026-06-01)
> **Last verified:** 2026-06-01
> **Scope:** Capability registry gap surfaced by F-001 audit; introduces four View-grade caps so Tier-1 declarative authz can split read vs write on writable prefixes.
> **Out of scope:** Tier-1 rule rewrite itself (lands in follow-up F-001 PR); Tier-2 enforcement; Postgres tripwire; codegen-from-OpenAPI (rejected per ADR 0007).
> **Key files:**
> - `internal/modules/iam/domain/model.go:42` — capability registry (`Capability` typed consts + `validCapabilities` map)
> - `apps/api/cmd/metaldocs-api/permissions.go:78,86,95,159,185` — Tier-1 rule rows blocked on missing read caps
> - `db/reference-data/0001_product_reference_data.sql:18` — `role_capabilities` seed shape
> - `wiki/decisions/0007-two-tier-authz.md` — two-tier model

## Context

F-001 audit (`wiki/references/qa-runs/approval-full-flow-20260601.md`, `wiki/references/qa-runs/plans-f001-f002.md`) found that the Tier-1 declarative authz table at `apps/api/cmd/metaldocs-api/permissions.go` conflates read and write capabilities on 13 rows. The correct fix is the surgical row split documented in Plan F-001: every writable prefix declares an explicit GET row with a View-grade cap plus per-verb rows with Manage/Submit caps (precedent at `permissions.go:104-117`).

The fix is blocked. The capability registry (`internal/modules/iam/domain/model.go:42-67`) contains exactly one read-grade cap (`CapDocumentView`) and one read-ish cap (`CapAuditRead`). It has no View cap for the user, membership, taxonomy, or metrics domains. Authors writing rule rows had no correct read cap to point at, so they fell back to Manage. Over time this hardened into convention.

Commit `8e1518c85` (`fix(taxonomy): downgrade read-path authz caps from Manage to View`) papered over the gap for taxonomy reads by reusing `CapDocumentView` for `FamilyRepository.GetByCode/List` and `ProfileRepository`/`AreaRepository` reads. The fix worked but crossed domain boundaries — taxonomy reads should not require a document capability — and signalled that the registry gap is the real problem.

## Decision

Introduce four new View-grade capability constants in the registry. Seed `role_capabilities` for existing roles. Do NOT alter `permissions.go` in this change — that is the follow-up F-001 PR.

New capabilities:

| Constant | String value | Domain | Purpose |
|---|---|---|---|
| `CapMetricsView` | `metrics.view` | Observability | Read `/api/v1/metrics` Prometheus endpoint |
| `CapMembershipView` | `membership.view` | IAM | Read access policies, area memberships |
| `CapUserView` | `user.view` | IAM | Read user lists, admin overview |
| `CapTaxonomyView` | `taxonomy.view` | Taxonomy | Read profiles/areas/families (replaces `CapDocumentView` workaround) |

Role grant matrix (proposed; product owner confirms before merge):

| Role | `metrics.view` | `membership.view` | `user.view` | `taxonomy.view` |
|---|---|---|---|---|
| `system_admin` | ✓ | ✓ | ✓ | ✓ |
| `area_admin` | — | ✓ | ✓ | ✓ |
| `approver` | — | ✓ | — | ✓ |
| `author` | — | ✓ | — | ✓ |
| `editor` | — | ✓ | — | ✓ |
| `viewer` | — | ✓ | — | ✓ |

Rationale per cap:

- **`metrics.view`** — Prometheus scrape endpoint. Operational/observability scope only. Granted to `system_admin` only; not a viewer-facing surface. Product may later expose to an `operator` role.
- **`membership.view`** — Every role needs to see who can sign for an area (UI shows approver lists, area pickers). Granted to all roles.
- **`user.view`** — IAM admin surface. Granted to admin-ish roles (`system_admin`, `area_admin`). `approver`/`author`/`editor`/`viewer` do not need user directory access.
- **`taxonomy.view`** — Every role browses documents and thus needs taxonomy (profile/area/family) lookups. Granted to all roles. Replaces the `CapDocumentView` workaround introduced by commit `8e1518c85`.

## Consequences

- **Positive:** unblocks F-001 surgical fix without cross-domain cap leak. Restores principled read/write separation in the registry. Future Tier-1 rule authors have correct caps to point at.
- **Positive:** matches AWS IAM / GCP IAM / Kubernetes RBAC pattern — atomic verbs, no implicit hierarchy.
- **Positive:** caps grant additively. Holding View does not imply Manage and vice versa. Audit log records exact cap exercised.
- **Negative:** `validCapabilities` registry grows from 20 to 24 entries. Trivial.
- **Negative:** taxonomy Tier-2 calls (`internal/modules/taxonomy/infrastructure/family_repository.go:39`, `:71`; `repository.go`) currently pass `CapDocumentView` (post-commit `8e1518c85`). Should migrate to `CapTaxonomyView` in the F-001 follow-up to remove the cross-domain leak. Until then, both caps point at the same set of roles, so behavior is identical.
- **Negative:** product owner must sign off on the role grant matrix before the migration ships. Default-grant rules are not derivable from code; they are policy.
- **Open:** future `operator` role for `metrics.view` if/when product exposes the scrape endpoint outside ops.

## Implementation

1. **`internal/modules/iam/domain/model.go`** — add four `Capability` consts (see table above), add four entries to `validCapabilities` map. No other code change.
2. **`db/migrations/0217_view_grade_capabilities.sql`** — `INSERT … ON CONFLICT DO NOTHING` rows into `metaldocs.role_capabilities` per the grant matrix. Wrap in `BEGIN`/`COMMIT`. Append `schema_migrations` row.
3. **`db/reference-data/0001_product_reference_data.sql`** — append the same rows so curated baseline rebuilds include them. Reference-data file is idempotent (uses `ON CONFLICT … DO UPDATE`).
4. **`wiki/modules/iam.md`** — extend the cap table; bump `Last verified`.
5. **`wiki/concepts/authz-tiers.md`** — note the four new caps in the "Modules with tier-2 coverage" table footnote; bump `Last verified`.

## Security Boundary Notes

`membership.view` is granted broadly (all non-system_admin roles). Security review flagged this as potential roster over-grant. Verified during review:

- **`GET /api/v1/iam/area-memberships`** — handler `MembershipHandler.listMemberships` at `internal/modules/iam/delivery/http/routes_memberships.go:37` defaults `userId` to the authenticated actor and runs `canManageMembershipTarget` (`:172`). That predicate returns true only when `actor == target` or actor holds `RoleSystemAdmin`. Non-admin roles receive 403 when querying any user other than self. Tier-1 `membership.view` grant therefore exposes self-membership read only; full cross-area roster remains system_admin-gated at the handler. No data minimization violation.
- **`/api/v1/access-policies` (GET/PUT)** — no handler currently registered in `internal/modules/iam/delivery/http/`. The Tier-1 rule rows at `apps/api/cmd/metaldocs-api/permissions.go:86-87` are placeholders for a future feature. Grant has no effect today. F-001 follow-up rewrite must coordinate with whichever handler ships first.
- **`/api/v1/iam/users` (GET)** — admin surface. `user.view` granted only to `system_admin` and `area_admin` per matrix. Not exposed to viewer/author/editor/approver.
- **`/api/v1/metrics`** — Prometheus scrape. `metrics.view` granted only to `system_admin`.
- **`/api/v1/taxonomy/{profiles,areas,families}` (GET)** — read-only taxonomy metadata. Already implicitly readable via `CapDocumentView` workaround (commit `8e1518c85`). Granting `taxonomy.view` broadly removes the cross-domain leak without expanding effective access.

Conclusion: the four new caps gate only routes that either (a) already have handler-level self-scope filters, (b) are admin-only by grant matrix, or (c) have no handler yet. No new data-exposure surface is created by this prereq alone. F-001 follow-up PR adds a (method, path) → cap matrix test that locks these boundaries.

## Alternatives Considered

| Option | Verdict | Reason |
|---|---|---|
| Reuse `CapDocumentView` for all read rows | Rejected | Cross-domain cap leak. Audit log loses fidelity ("user exercised document.view" on IAM admin overview page is misleading). Locks in the taxonomy precedent as a permanent pattern. |
| Codegen caps from OpenAPI `security:` blocks | Rejected | ADR 0007 already rejected this for Tier-1 (no `*sql.Tx` pre-handler). Same constraint applies to cap definition. |
| Hierarchical "View implied by Manage" rule | Rejected | Adds implicit grant logic to `CapabilityService.CanDo`. Breaks the additive grant model, complicates audit, breaks future per-cap revocation. AWS/GCP/K8s precedent is explicit grants only. |
| Single coarse `iam.view` cap covering both user and membership reads | Rejected | Conflates two surfaces with different role audiences. Membership is broadly readable; user directory is admin-only. |

## Rollback

- Revert `model.go` const + map entries.
- New down migration removing the four cap rows (no schema change, only data).
- F-001 follow-up PR is blocked until prereq lands or is re-merged, so revert is contained.

## References

- `wiki/references/qa-runs/plans-f001-f002.md` — F-001 plan, audit truth table
- `wiki/references/qa-runs/approval-full-flow-20260601.md` — browser-confirmed F-001 evidence
- `wiki/decisions/0007-two-tier-authz.md` — two-tier model + codegen rejection
- Commit `8e1518c85` — taxonomy read cap workaround (precedent this ADR replaces)
- `internal/modules/iam/domain/model.go:42-67` — capability registry
- `db/reference-data/0001_product_reference_data.sql` — `role_capabilities` seed
