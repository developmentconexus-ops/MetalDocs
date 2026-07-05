# ADR 0069 — Document periodic review/expiry, the `document.review` capability, and structured reason-for-change

> **Status:** Accepted 2026-07-04
> **Module(s):** `documents` · `documents/approval` · `iam` (capability registry + tripwire arm) · `jobs` (River review surfacer) · `audit`
> **REQ IDs:** `backend-target-architecture.md` REQ-AUTHZ-5 (capability coherence), REQ-ASYNC-* (River periodic jobs), REQ-DB-* (DB-enforced invariants)
> **Supersedes / amends:** none. Extends ADR 0022 (capabilities-never-roles) with one new capability; extends the M2 (ADR-less) tripwire-generation mechanism (migrations 0269/0270/0271) with migration 0275; complements ADR 0067 (async on River).

## Context

M6 (eQMS periodic review/expiry + structured reason-for-change) adds a controlled-document
lifecycle that ISO 9001 / 21 CFR Part 11 regulated operations require but MetalDocs did not yet
model:

- **Periodic review / expiry.** A published controlled document must carry a *next-review-due* date
  and, optionally, an *expiry* date; a governance actor must be able to record that a review was
  completed (resetting the cycle) without transitioning the document's status.
- **Effective-date vs publish-date.** Runtime-truth census (M6 validation-contract §0.1) established
  that `public.documents.effective_from` is **already** the effective date, written at
  schedule/publish; and `effective_to` **already exists but is unwritten**. There is no separate
  publish-date concept to reconcile — the two dates coincide today.
- **Structured reason-for-change (F6.3).** Revision creation captured only a free-text
  `revision_title`; a regulated attributable *reason for change* (with an optional category) was not
  a first-class field.

The capability-registry bump convention (`iam/domain/model_test.go:96` — "bump only via ADR")
makes this ADR a gated deliverable of F6.2.

## Decision

1. **Reuse `effective_from`/`effective_to`; add `review_due_at` + `last_reviewed_at`.** The
   review/expiry model is wired on `public.documents` **without a duplicate column family**:
   `effective_from` (effective date, reused), `effective_to` (expiry, newly wired),
   `review_due_at`, `last_reviewed_at` (new). All nullable (expand-only; legacy rows keep NULL = no
   cycle; no backfill). DB CHECKs enforce the invariants (expiry strictly after effective date;
   `review_due_at >= effective_from`) — the DB is the last line, the app the friendly first line
   (migration `0274_document_review_and_reason.sql`).

2. **Effective date = publish date (recorded, not changed).** M6 does not introduce a distinct
   publish-date; `effective_from` remains the single effective/target date set at schedule/publish.
   A future decoupling, if ever needed, is a separate ADR, not smuggled into M6.

3. **New capability `document.review` (ScopeTenant).** A dedicated capability gates the
   *mark-reviewed* workflow — recording a review completion on a live published revision (sets
   `last_reviewed_at` + the next `review_due_at`; **not** a status transition, status stays
   `published`). It is a **new** capability, not a reuse of `document.edit`/`document.publish`,
   because the authorization boundary is genuinely distinct: an actor authorized to *attest a
   periodic review* is not necessarily authorized to *edit content* or *publish*. It is classified
   **`ScopeTenant`** — the mark-reviewed act is a tenant-wide governance attestation, not an
   area-scoped content WRITE (tier-2 passes the `"tenant"` sentinel). Held by the governance actors
   `area_admin` + `qms_admin` (mirroring `document.publish`/`document.obsolete`); `system_admin`
   reaches it via the tier-2 bypass (not seeded explicitly). Wired across all capability touchpoints
   per REQ-AUTHZ-5 (const + `validCapabilities`, scope classify, catalog description, tier-1
   route→cap for `POST /documents/{documentId}/review`, seed grants, guard tests, registry size
   34→35).

4. **Tripwire-arm widening (migration 0275).** The mark-reviewed workflow asserts **only**
   `document.review` (tier-2 `authz.Require`) then UPDATEs `public.documents` — so the DB tripwire's
   `documents/UPDATE` arm must accept `document.review`, or every mark-reviewed UPDATE fails-closed
   `P0001` for every actor (the exact 0269/0270/0271 defect class). The arm is **additively** widened
   to `{document.edit, document.obsolete, membership.manage, document.review}`, generated from the
   `internal/platform/tripwire` Go source of truth (M2's `TripwireArms` + `RenderMigration`) into
   `db/migrations/0275_documents_update_tripwire_review_cap.sql`. This is the intended
   registry-driven growth path, not drift; a dated forward-erratum was appended to M2's
   validation-contract §1.2 recording the extension (HS-7 transparency).

5. **Surface review-due via River (not a hand-rolled scheduler).** A River periodic job in
   `metaldocs-jobs` flags documents due for review, reading through a **new documents published
   read-port** (`jobs → documents` via interface, never raw SQL on another module's table) —
   idempotent, tenant-seeded (M3 backstop), consistent with ADR 0067 (River is the single async
   primitive) and the transactional-outbox invariant.

6. **Structured reason-for-change (F6.3).** `reason_for_change` (+ optional `reason_category`) is a
   distinct contract field on the submit-revision request (not `revision_title`), persisted on
   `public.documents` in the submit business tx and carried into the audit trail via the existing
   `approval_submitted` governance-event payload (no new inline network call). Required at the API
   for REV≥1 (RFC 9457 `problem+json` 422 on omission), nullable in DB for legacy rows.

## Consequences

- Capability registry grows to **35**; `TestCapabilityRegistrySize` want bumped 34→35 with this ADR
  cited. Seed grant count 103→105 (area_admin + qms_admin).
- The `documents/UPDATE` tripwire arm now has four accepted caps; migration 0275 is the latest
  definition of `public.enforce_capability_asserted()` (supersedes 0271). The M2 golden test and
  api-lint `TRIPWIRE-ARM-PARITY` target advance from 0271 to 0275.
- An arm-level negative integration proof
  (`tests/integration/documents/tripwire_documents_test.go`) pins that the review UPDATE succeeds
  under `document.review` and is rejected `P0001` with no cap asserted — independent of the T5
  handler.
- No advisory lock is taken on the mark-reviewed path (H-PRE-1 satisfied by construction).
- The tier-2 `authz.Require(ctx, tx, CapDocumentReview, "tenant")` call and the mark-reviewed
  service/handler land in a later M6 feature (T5); the tier-1 gate, seed, arm, registry, and ADR
  are in place first so that path is fail-open-free the moment the handler ships.

## Future work (explicitly out of scope)

- Notification/escalation on review-overdue beyond surfacing (M8 ops-readiness or a product
  decision).
- Backfilling historical `reason_for_change` onto legacy rows (never — intent cannot be
  reconstructed).
- Training-acknowledgment / obligated-reader review-attestation (`distribution` module owns it).
