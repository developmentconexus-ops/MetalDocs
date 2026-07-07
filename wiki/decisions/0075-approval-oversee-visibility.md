# ADR 0075 — `approval.oversee` Capability + Visibility Model

> **Status:** Accepted
> **Date:** 2026-07-07
> **Scope:** New capabilities `approval.review` (area-grade) and `approval.oversee` (tenant-grade);
> tier-1 explicit routing for approval runtime verbs; oversight read-path visibility model.
> **Milestone:** `docs/superpowers/milestones/approval-remediation/milestone-2b-approval-kernel-backend/f3-capabilities-review-oversee/`
> **Key files:**
> - `internal/modules/iam/domain/model.go` — capability consts
> - `internal/modules/iam/domain/capability_scope.go` — scope classification
> - `internal/modules/iam/domain/catalog.go` — pt-BR descriptions
> - `apps/api/cmd/metaldocs-api/permissions.go` — explicit tier-1 rows
> - `internal/modules/documents/approval/application/read_service.go` — two-capability oversight check
> - `db/reference-data/0001_product_reference_data.sql` — seed grants

## Context

Two gaps existed in the approval module's capability model:

1. **Tier-1 route→capability mapping used a generic prefix fallback.** `permissions.go` mapped every
   method under `/api/v1/approval/` via 4 prefix rows keyed only on HTTP verb (GET/POST/PUT/DELETE),
   all resolving to `CapDocumentView`/`CapDocumentSubmit`. This silently diverged from the real
   tier-2 checks inside individual handlers (e.g. `document.signoff` for stage signoff,
   `document.edit` for cancel) — the same class of cross-tier mismatch BE-9 already fixed for
   route-admin (2026-07-02, ADR 0018 §6). It also encoded 2 routes (`PUT`/`DELETE
   /instances/{id}`, `POST /instances/{id}/decisions`) that were never actually registered by the
   real router — stale test fixtures that only ever passed because the fallback silently caught
   any method/path under the prefix.
2. **No capability exists for "act on a review-kind approval stage"** (suggestions/parecer mode,
   coming in F4) or for **"oversee any approval instance in the tenant"** (worklist "all" view,
   cockpit observer mode, coming in F8). Both are read/act concerns distinct from the document
   lifecycle capabilities (`document.view`, `document.signoff`) that today gate all approval reads —
   conflating "can view a document" with "can act on / oversee an approval process" prevents
   granting the two independently in the future.

`document.review` (ADR 0069, eQMS periodic-review mark-reviewed) already exists and is unrelated —
naming the new review-stage capability `approval.review` avoids that collision, per the ratified
design spec (`docs/superpowers/specs/2026-07-07-approval-remediation-design.md` §6).

## Decision

### 1. Two new capabilities, classified

- `approval.review` (area-grade) — acting on a review-kind approval stage where the actor is in the
  stage pool. No live tier-1 route in F3 (F4 introduces `POST
  /approval-instances/{id}/stages/{stageId}/review-verdict`, the first consumer).
- `approval.oversee` (tenant-grade, read-only) — oversight of any approval instance in the tenant,
  regardless of area membership or process pool. First live consumer: the oversight-read alternative
  check below (F3). Tenant-grade because oversight cuts across area boundaries by definition —
  oversight is a capability, never a role (ADR 0022).

Both are registered as `Capability` consts, added to `validCapabilities`, classified in
`capabilityScopes`, and given pt-BR descriptions in `capabilityDescriptions`.
`TestCapabilityRegistrySize` 38 → 40; `TestAreaGradeCapabilitySet`'s locked area-grade set 11 → 12
(adds `CapApprovalReview` only — `CapApprovalOversee` is tenant-grade, not part of that set).

### 2. Tier-1 generic prefix fallback deleted; explicit rows per real runtime verb

The 4 generic `/api/v1/approval/` prefix rows are replaced with one explicit row per real registered
route (verified against `internal/modules/documents/approval/http/router.go`): signoff →
`CapDocumentSignoff`, cancel → `CapDocumentEdit`, get-instance → `CapDocumentView`, inbox →
`CapDocumentView`. This is a **routing-truth fix, not a capability-model change** — no tier-2 check
changes; every row now matches the capability its own handler already enforces at tier-2. Route-admin
rows (`/api/v1/approval/routes/*`) are untouched — that tier-1/tier-2 gap was already closed as BE-9.

This targets only the runtime verbs found still diverging; it does not touch submit/publish/etc.
sites already coherent (per the ratified design's explicit scope note).

### 3. Oversight read paths accept `approval.oversee` as an explicit alternative capability

`LoadInstance` and `LoadActiveInstanceByDocument` (`read_service.go`) each try
`authz.Require(CapDocumentView, "tenant")` first; on failure, try
`authz.Require(CapApprovalOversee, "tenant")`; only if **both** fail is the first error returned.
This is an explicit two-capability check — never a role check (ADR 0022) — and lets a future
`approval.oversee`-only actor (e.g. a quality-manager profile with no direct document.view grant)
read any instance without weakening the existing `document.view` path.

### 4. Seed grants mirror real actor pools, not literal "quality-manager"/"reviewer" role names

The design's language ("reviewer pools get approval.review", "quality-manager profile gets oversee")
describes conceptual actors, not literal role strings — MetalDocs has 8 canonical roles and no
`reviewer`/`quality_manager` role (`reviewer` is explicitly decommissioned,
`internal/modules/iam/domain/model_test.go` `TestIsAreaRole`). Grants map to the real roles that
already act on approval stages and are closest to those personas:

- `approval.review` → `approver`, `area_admin`, `qms_admin`, `signer`, `system_admin` — the exact
  same pool that already holds `document.signoff` (the real actors who act on a stage today).
- `approval.oversee` → `qms_admin`, `system_admin` — `qms_admin` is the closest existing role to
  the design's "quality-manager profile".

Seeded only in `db/reference-data/0001_product_reference_data.sql` (the canonical seed-grant file) —
**not** as a `db/migrations/*.sql` file. Verified: no `role_capabilities` INSERT exists anywhere in
`db/migrations/` for any capability, ever; seed grants are reference-data, not schema migration.

### 5. No new tripwire arm

Tripwire arms only ever gate INSERT/UPDATE, never reads. `approval.oversee` is read-only.
`approval.review` has no live write surface until F4 creates the review-verdict route — its arm (if
any is needed) is F4's concern, proven at that point via the same drift/parity lints. Both
`TRIPWIRE-ARM-DRIFT` and `TRIPWIRE-ARM-PARITY` pass unchanged with zero `internal/platform/tripwire/arms.go`
edits in this feature — confirming by construction, not assertion, that F3 needs none.

## Consequences

- **Positive:** tier-1/tier-2 coherence restored for all real approval runtime routes; the stale,
  never-registered synthetic routes are removed from both the resolver's test cases and the
  route-coverage fixture.
- **Positive:** oversight and review-stage participation are now independently grantable
  capabilities, unblocking F4 (review verdicts) and F8 (oversight worklist/cockpit) without further
  capability-model changes.
- **Neutral:** `CapApprovalReview` has zero live routes until F4 — an intentional, documented defer
  (spec.md), not a gap; `TestEveryCapSeededOrDeferred` passes because it is seeded, independent of
  route existence.
- **Negative:** the seed-grant role mapping is a judgment call (BE-9-adjacent) — mirroring the
  existing `document.signoff` pool for `approval.review` — since the design's persona language
  doesn't map 1:1 to canonical role names; documented here rather than left implicit.

## Alternatives Considered

| Option | Verdict | Reason |
|---|---|---|
| Name the new review-stage cap `document.review` | Rejected | Already claimed by ADR 0069's eQMS periodic-review mark-reviewed workflow — a genuinely different concern; reusing the name would collide two unrelated tier-2 gates under one string. |
| Grant `approval.oversee` via a role check ("if role == quality_manager") | Rejected | Violates ADR 0022's core invariant — capabilities, never roles. `quality_manager` isn't even a canonical role. |
| Seed grants via a new numbered `db/migrations/NNNN` file | Rejected | No precedent — every existing capability's grants live solely in `db/reference-data/0001_product_reference_data.sql`; migrations only ever carry schema. |
| Preserve the stale PUT/DELETE/`/decisions` test fixtures to avoid touching test files | Rejected | They asserted routes that were never registered by the real router — preserving fictitious coverage is worse than correcting it to runtime truth. |

## Rollback

Additive only: two new capability consts + classifications + descriptions, 7 new
`role_capabilities` seed rows, and a tier-1 routing correction (no schema/migration change).
Reversible by removing the consts/seed rows and restoring the generic prefix fallback rows, provided
no `approval.review`/`approval.oversee` grants are relied upon by then.

## References

- `docs/superpowers/specs/2026-07-07-approval-remediation-design.md` §6
- `wiki/decisions/0069-*.md` (document.review — the capability this ADR's naming avoids colliding with)
- `wiki/decisions/0018-approval-route-lifecycle.md` §6 (BE-9, route-admin tier-1/tier-2 fix — the
  precedent this ADR's tier-1 fix follows for runtime verbs)
- `docs/superpowers/milestones/approval-remediation/milestone-2b-approval-kernel-backend/f3-capabilities-review-oversee/spec.md`
