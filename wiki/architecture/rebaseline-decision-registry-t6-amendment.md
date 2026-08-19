# Rebaseline Decision Registry — T6 Closure Amendment

> **Status:** ACTIVE / OPERATOR-RATIFIED REGISTRY RECONCILIATION  
> **Ratified:** 2026-08-18  
> **Parent registry:** `wiki/architecture/rebaseline-decision-registry.md`  
> **Detailed T6 authority:** `wiki/architecture/r10-t6-canonical-api-frontend-journeys.md`  
> **Implementation:** BLOCKED

This bounded amendment reconciles the closed T6 REOPEN set into the Rebaseline Decision Registry without rewriting unrelated parent-registry dispositions. Until the next full registry consolidation, read this file immediately after the parent registry and the earlier D4 amendment.

Everything in the parent registry not named below remains unchanged.

---

## 1. T6 stage status

```text
T6 — Canonical API / Frontend Journeys = CLOSED / OPERATOR-RATIFIED
```

The former T6 REOPEN set is no longer open.

Durable authority:

`wiki/architecture/r10-t6-canonical-api-frontend-journeys.md`

---

## 2. Numbering / creation

Former `DOC-12` REOPEN is closed as:

```text
numbering_scope = DOCUMENT_TYPE | DOCUMENT_TYPE_AREA
separator = '-'
sequence = decimal, minimum display width 3, natural expansion after 999
normalized DocumentType.code unique within Company
normalized Area.code unique within Company
committed Document.code unique and never reused
preview reserves nothing
no custom format/year/reset grammar Launch baseline
```

Area code is immutable after creation. DocumentType code + numbering scope become immutable after first committed use.

Creation/reference selectors use purpose-built least-privilege `document-creation/options`; they do not depend on admin/PII APIs.

---

## 3. Working Content / editor

Former `CNT-03` REOPEN is closed as:

```text
EditorSession/lease = NOT Launch correctness baseline
WorkingContent generation/OCC remains DRAFT correctness authority
```

One strong DRAFT ETag/If-Match token protects both Revision title and WorkingContent source mutation. Stale DRAFT mutation is 412 with zero mutation.

Exactly one interactive DOCX editor/viewer adapter is selected after representative fidelity proof. A separate T5 OfficialRendition renderer may use a different product.

---

## 4. Upload / content UX

T4 admission is exposed through:

```text
OPEN upload_id
→ direct create-only provider upload
→ server exact verification/descriptor derivation
→ READY
→ DRAFT OCC attach
```

The client never owns the ExactContentDescriptor. Provider upload success is not READY, and READY is not semantic WorkingContent/Submission truth.

Exact byte resources remain semantic MetalDocs URLs; provider bucket/key/version/managed-content handle never becomes public product identity.

Range reads are optional mechanism, not semantic baseline.

---

## 5. Search/read/history/audit

T6 proves no materialized Search consumer for Launch:

```text
canonical PostgreSQL query/view = Launch Search baseline
materialized Search projection = OFF
search_refresh = OFF
external Search engine = OFF
```

Free text is current EFFECTIVE code/title with deterministic rank; typed filters include Document Type, Area, responsible owner and lens-valid derived status.

No persisted `Document.currentStatus` is introduced.

Ordinary Library never presents DRAFT/SUBMITTED as official.

Domain History and Audit remain distinct read surfaces. Audit never reconstructs lifecycle state.

---

## 6. Admin / Authorization-facing UX

Admin Center sections are:

```text
Organization
Access
Document Governance
```

Permission ownership remains T3; UI co-location does not merge permissions.

Launch exposes no custom Role/Permission editor, generic workflow designer, session admin platform or dormant Launch+/Future administration.

GroupMembership current list is available under `access.manage` through a bounded cursor-paginated UserReference surface.

Template administration uses a bounded `template_use.manage` metadata projection and grants no implicit effective-content/history read.

Responsible-owner target eligibility is governed by the ratified D4 amendment:

```text
existing User + same Company + ENABLED
```

The relationship grants no Role/Permission.

---

## 7. Authentication / browser contract

Keycloak remains selected provider evidence behind the Authentication anti-corruption boundary.

Launch browser AuthN is Authorization Code + PKCE → MetalDocs ApplicationSession.

```text
no local password API
no Direct Grant/ROPC
no JIT User creation
session-bound CSRF on unsafe same-origin application requests
provider roles/groups never T3 authority
```

`/api/v1` application contract, `/auth` browser integration and operations/readiness surfaces are separate classes.

---

## 8. Errors / pagination / read models

Public application failures use RFC 9457 Problem Details with one canonical MetalDocs `code`; `type` is mechanically derived.

Potentially unbounded lists use opaque cursor pagination:

```text
limit default 20
limit max 100
no mandatory total
no offset baseline
no generic filter/sort DSL
```

Public DTOs are purpose-built semantic read models, not database/module dumps.

`allowed_actions` are UX hints only and must share canonical T3/domain predicate components; commands always recheck authority.

---

## 9. Idempotency

Natural HTTP/resource idempotency is preferred before durable replay machinery.

`Idempotency-Key` is required only for non-idempotent semantic POST creation where an uncertain retry could duplicate a fact/resource.

Replay record and semantic mutation commit atomically in the same local PostgreSQL transaction.

Completed replay still rechecks current T3 access/disclosure authorization before returning stored response; the already-completed historical mutation and its original lifecycle preconditions are not re-executed.

Replay retention is bounded operational policy; `24h` is only an implementation-default candidate, not architecture invariant.

Replay storage must not become an unintended retention root for erasable UserProfile PII.

---

## 10. Launch operation census

The T6 durable authority owns the closed Launch `/api/v1` application operation families. New route families require a named Product Contract journey or an explicit bounded T6 reopen.

No generic `/actions`, separate Approval API, separate Template lifecycle API, generic Search API or user Release/publish mutation exists.

---

## 11. Future seams preserved

T6 was explicitly checked against:

```text
Distribution / Read & Acknowledge
Periodic Review
Dossier
Evidence
Retention / Legal Hold / Disposition
Governed Export
External Repository IMPORT/PUBLISH
Training/LMS
multi-document Change Control
pooled tenancy
CRDT/realtime collaboration
```

All remain attachable without duplicating Document/Revision/Submission/Release/User/Group/Area authority and without dormant Launch implementation.

---

## 12. T7 status

```text
T7 — Historical Migration & Cutover = NEXT / ACTIVE ONLY AFTER T6 PROMOTION + STAGING CLEANUP
```

T7 consumes only its parent-registry REOPEN set:

```text
actual source evidence census
CURRENT_STATE/FULL_HISTORY or smaller real modes
imported target-owned fact shapes
ordinal/content/governance provenance
plan/dry-run/idempotency/reconciliation
semantic-unit atomicity
cutover/readiness/rollback/deletion map
concrete restore/erasure and post-snapshot security-teardown reconciliation choreography where cutover/recovery requires it
```

Implementation remains BLOCKED.