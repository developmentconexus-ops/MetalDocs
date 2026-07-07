# Approval Remediation — Design Spec (M2b backend kernel + M2c FE screen)

**Date:** 2026-07-07
**Status:** Ratified by operator (brainstorm §A/§B/§C approved in session)
**System-impact analysis:** `docs/superpowers/analysis/2026-07-07-approval-remediation-m2b-system-impact.md` (Yellow, committed 124f2c37)
**Origin:** adversarial macro review of the approval system (workflow findings W1–W13, permission findings P1–P8, FE findings P0–P2) + 5 operator observations (URL gating/oversight, sidebar clutter, duplicated timeline, sidebar scroll, editor-shell reuse + suggestion mode).
**Standard:** professional eQMS, industrial SaaS, global maximum, evidence-based (industry research, not assumption).

---

## 0. Decomposition

Two milestones, backend first:

| Milestone | Scope |
|---|---|
| **M2b** `approval-kernel-backend` | Workflow model (stage kinds, freeze boundary, route versioning, SLA, delegation, pool validation, SoD, meaning-of-signature), permissions (tier-1/tier-2 fixes, visibility gating, `approval.oversee`), contract + DB migrations, all W/P minors |
| **M2c** `approval-screen-fe` | Cockpit = editor shell reuse, sidebar IA, suggestion UX (eigenpal), worklist single destination, author request-changes panel |

M2b ships and validates before M2c starts. Both run under the `milestone` skill with HS-1 operator gates.

---

## 1. Evidence base (why this shape)

Industry research performed during brainstorm (Veeva QualityDocs, MasterControl, Documentum, 21 CFR Part 11, ISO 9001):

- **Review stage ≠ approval stage.** Vendors separate a collaborative review stage (comments/markup, NO e-signature) from an approval stage (e-signature). MetalDocs today has only signature stages.
- **Freeze before signature.** Documentum Draft→Final Draft auto-accepts all tracked changes before signature routing. Markup never persists into the signed artifact (Veeva keeps annotations in a rendition layer; Adobe treats post-signature modification as signature invalidation).
- **"Requires changes" ≠ "Rejected".** Veeva distinguishes a review verdict returning the doc to the author from a terminal rejection. MasterControl reject auto-returns to originator.
- **Meaning of signature.** 21 CFR 11.50(a)(3): the signed record must state the meaning (review, approval, responsibility, authorship). Our signoffs today record decision but not the signature meaning.
- **Consumer visibility.** ISO 9001 §7.5.3 / 21 CFR 820.40: consumers see only the effective (published) revision. NO surveyed vendor shows a consumer-facing "revision in progress" banner.

---

## 2. Domain model — stage kinds and lifecycle (W2, W8, W10)

### 2.1 Stage kinds

`approval_route_stages` gains `stage_kind ∈ {review, approval}` (DB CHECK + Go enum, aligned — closes the Go/DB enum drift class).

**Review stage** (collaborative):
- Document opens in eigenpal `suggesting` mode; reviewer edits become tracked changes (`w:ins`/`w:del`), comments via native OOXML comment marks.
- Verdicts: `ready` (advance) or `request_changes` (return to author). **No password re-auth** — review is not a signature event. `request_changes` requires a mandatory comment.
- Quorum semantics identical to today (any_1_of / all_of / m_of_n) applied to verdicts.

**Approval stage** (signature):
- Document opens read-only (`viewing` mode). No writable session, no autosave — the current W2 vector (`ReviewDocumentCanvas` opens a writable `useDocumentSession` during `under_review`) is structurally removed.
- Verdicts: `approve` / `reject`, both with password re-auth (existing Part 11 signoff path: bcrypt, rate limit, fail-closed, content-hash echo).
- **Meaning of signature (W8):** signoff record gains `signature_meaning` (text, from a fixed vocabulary per 11.50(a)(3), default `approval`), rendered in the signature manifest and audit trail.

### 2.2 Freeze boundary (W2 core)

Transition **last review stage → first approval stage** is the freeze point:

1. Gate: all tracked changes resolved (accepted/rejected) AND all `document_comments` on the instance resolved. **W10's unresolved-comments gate moves here** (today it fires at final approve — too late).
2. Comment marks stripped from the buffer (`removeCommentMark` per mark) → clean OOXML, no markup in the signed artifact.
3. Canonical content hash computed over the clean buffer and **pinned** on the instance (W9). Precise defect statement: `content_hash_at_submit` IS written at submit today (`submit_service.go`), but two structural defects remain. (a) The document stays editable during `under_review`, so the submit-time pin can diverge from signed content — fixed by the freeze re-pin over the clean post-review buffer. (b) The signoff comparison path (`LoadActiveDocumentContentHash`, `postgres_approval_repository.go:1132-1154`) contains a `COALESCE` fallback to the head revision hash: at signature time the pin MUST exist, so this branch is either dead or — if it ever fires — silently validates the signature against floating head content, masking a defect. **No-fallback rule for the hash chain:** signoff and publish read ONLY the frozen pin and fail closed (`ErrNoActiveContentHash` → 409/problem+json) when it is absent; the display path asks a status-explicit question (draft → head revision hash; frozen instance → pin) instead of one polymorphic `COALESCE`. Both `COALESCE` expressions are deleted, not preserved. One canonical hash chain: pin at freeze → echo at signoff → verify at publish — no substitute value at any link.
4. From freeze onward the document version is immutable for the instance; any content change requires reject/withdraw → new revision.

Routes with no review stage (approval-only): freeze fires at submit (today's behavior, now explicit).

**Concurrency at the choke point (design commitment, detailed in ADR 2):**
- **Freeze race:** freeze executes atomically inside the stage-transition tx with OCC on the instance row (status CAS); concurrent transition attempts lose the CAS and get 409. Freeze is idempotent — re-entry on an already-frozen instance is a no-op. H-PRE-1 respected: no authz-recording read inside the lock-holding tx.
- **Concurrent verdicts:** review-verdict endpoint uses the same per-instance OCC/CAS pattern as signoff today; quorum evaluation happens in-tx after the verdict insert, so two simultaneous verdicts serialize and the second sees the first.
- **Comment-resolution scope:** the existing `HasUnresolvedComments` predicate (`postgres_approval_repository.go:1110-1126`) is document-wide; the freeze gate scopes it to comments created during the instance's review stages (instance-scoped predicate) so stale historical comments on prior revisions cannot block freeze.

### 2.3 Return-to-author and rejection

- `request_changes` (review stage): instance enters `changes_requested`; author sees pending suggestions/comments in the editor, resolves them (accept/reject per change, resolve per comment), re-submits into the **same instance** at the same review stage. Route version stays pinned.
- `reject` (approval stage): terminal — instance collapses, document returns to draft (existing behavior), reason mandatory.
- `cancel` (author/oversee): gains mandatory `cancel_reason` column (W13).

### 2.4 Deleted

- `SkipStage` service path (W11) — no product surface, violates route immutability. Deleted, not deprecated.

---

## 3. Route versioning (W1)

Route definitions become **versioned and immutable**: editing a route creates `approval_routes` version N+1; in-flight instances stay pinned to the version they started on. The `enforce_route_immutable` trigger patch approach (baseline:636–649) is superseded — immutability is structural (new row per version), the trigger becomes a tripwire on published versions only.

- `UNIQUE(tenant_id, profile_code)` (baseline:2890) becomes partial on the active version.
- Supersedes **ADR 0018 §1/§3** — new ADR required (see §9).
- Pool membership (`eligible_actor_ids`) stays frozen per stage at submit (existing design, kept), with **pool validation at submit (W6)**: empty effective pool → 422 with actionable problem detail, not a stuck instance.

## 4. SLA / escalation (W4) and delegation (W5)

- **SLA:** stage gains optional `due_in_days`; instance stage tracks `due_at`. The existing `document-review-surfacer` periodic job (jobs module, River `maintenance` queue) also surfaces overdue approval stages → notifications fanout. Alert-only; no auto-action (consistent with ADR 0068 watchdog philosophy).
- **Delegation:** explicit `approval_delegations` (tenant, delegator, delegate, window, reason, audit). Delegate acts *as themselves on behalf of* — signoff records both identities. No credential sharing, no role impersonation. Overlapping windows for the same delegator are allowed (union semantics — any active delegation makes the delegate eligible); self-delegation rejected. New ADR (see §9). **Sequencing:** delegation is separable from the freeze/versioning core — at plan time, slice it as a late M2b feature (or its own follow-up milestone) so it cannot destabilize the choke-point work.

## 5. SoD unification (W7)

Single SoD predicate (author ≠ approver on same instance; reviewer may be anyone but author self-verdict blocked) implemented once in the application service and mirrored by ONE DB trigger — the current app+trigger duplication is collapsed to app-checks-first + DB-tripwire-last (invariant pattern), same rule text in both.

---

## 6. Permissions (P1–P8) and capabilities

### 6.1 New capabilities

Two new capabilities, wired through the full 10-touchpoint checklist (`developing-new-work/references/capability-wiring.md`), including `TestCapabilityRegistrySize` bump and tripwire/lint parity:

| Capability | Grants |
|---|---|
| `approval.review` | act on a **review** stage where actor is in the stage pool (verdicts, suggesting-mode session). Named `approval.review` — NOT `document.review`, which already exists (`CapDocumentReview`, ADR 0069 periodic mark-reviewed workflow, `internal/modules/iam/domain/model.go:89`) with a different production meaning |
| `approval.oversee` | read-only oversight of ANY instance in tenant (worklist "all", cockpit observer mode, cancel with reason). Replaces the "admin wandered into approval URL" hole — oversight is a capability, never a role (ADR 0022) |

Existing signoff capability continues to gate approval-stage verdicts.

### 6.2 Tier-1 fix (P1)

Generic `/approval/` prefix fallback in `permissions.go:250-253` deleted. Every approval route gets an explicit route→capability entry; tier-2 `authz.Require` in-tx per handler; DB tripwire arms regenerated (M2-generation pipeline). **Scope note:** this targets the *runtime* verbs (submit/signoff/verdict/cancel/publish) still falling through the generic block — the route-admin tier-1/tier-2 gap was already closed as BE-9 (2026-07-02, ADR 0018 §6); do not redo it.

### 6.3 Visibility (P2/P3/P8) — ratified model

- **Consumers** (view-only capability holders): see ONLY the published/effective revision. Unpublished revisions, instances, and cockpit URLs → 404 (cross-boundary = not-found, consistent with tenancy rule). **No public "em revisão" flag.**
- **Author, route participants (current+past stages), `approval.oversee`, `document.edit` holders:** see the in-flight revision + "Em aprovação/revisão" badge.
- **Post-publication approval history:** visible to any actor who can view the document (ISO traceability).

---

## 7. Contract + data (expand/contract)

Contract-first: all route/verb changes land in `api/openapi` + `oapi-codegen` regen before handlers. New/changed surface (names indicative, final at plan time):

- `POST /approval-instances/{id}/stages/{stageId}/review-verdict` (`ready` | `request_changes` + comment)
- Signoff request gains `signature_meaning`
- `POST .../cancel` gains required `reason`
- Route admin CRUD becomes version-creating (PUT → new version)
- Worklist endpoint gains stage-kind/due filters; oversee variant
- Instance DTO exposes `stage_kind`, `due_at`, `frozen_content_hash`

DB migrations expand/contract: add columns/tables (`stage_kind`, `due_*`, `signature_meaning`, `cancel_reason`, `approval_delegations`, route version columns) → backfill (existing stages ⇒ `approval`; existing routes ⇒ version 1) → tighten constraints. All tenant tables carry `tenant_id` + RLS FORCE per house rules.

Suggestion persistence: **no new table.** Tracked changes live natively in the docx buffer (OOXML `w:ins`/`w:del`); comments stay in `document_comments` anchored by eigenpal `library_comment_id` marks. Review verdicts live on stage-action records (existing action table extended), giving the audit trail without duplicating content state.

---

## 8. FE — M2c (ratified §C)

- **C1 shell único:** cockpit dies as standalone page; `DocumentEditorPage` shell + sidebar slot by mode — `author` (unchanged) / `review` (eigenpal `suggesting`, suggestion+comment cards via `extractTrackedChanges`, verdict CTAs) / `approval` (eigenpal `viewing`, decision CTA).
- **C2 sidebar IA:** stage context (etapa N/M, pool, due, quorum) → single timeline (duplicate dies) → integrity collapsed to "Conteúdo verificado ✓ · detalhes" (hash/ETag copy demoted to expandable auditor view) → decision CTA pinned at sidebar footer; sidebar scrolls internally, page doesn't. "Dados podem estar desatualizados" banner dies — react-query invalidation replaces the polling adapter.
- **C3 worklist destino único:** `/approvals` lists instances where actor is eligible on the active stage (or oversee). Real filters (stage kind, due, doc type), due-date sort, teaching empty state, items deep-link into cockpit in the right mode. Notifications point here.
- **C4 visual:** wine tokens 100% (`--brand #6b1f2a`), legacy slate palette removed from approval screens, visible focus, PT-BR, loading/empty/error states designed.
- **C5 autor pós-request-changes:** editor opens with "mudanças solicitadas" panel — per-change accept/reject (`acceptChangeById`/`rejectChangeById`), per-comment resolve; all resolved → re-submit enabled.

Eigenpal risk closed during brainstorm: track-changes + comments are first-class in `@eigenpal/docx-editor-core@1.9.0` (verified typedefs: `extractTrackedChanges`, `acceptChangeById`/`rejectChangeById`, `acceptAllChanges`/`rejectAllChanges`, `addCommentMark`/`removeCommentMark`; modes `editing`/`suggesting`/`viewing`; `MetalDocsEditor.tsx:172-173` already maps review→suggesting). Work is product UX + persistence, not engine building.

---

## 9. ADRs (written during M2b implementation)

1. Route definition versioning + pinned instances (supersedes ADR 0018 §1/§3).
2. Content freeze boundary + review layer (stage kinds, markup stripping, canonical hash chain, choke-point concurrency: freeze OCC/CAS + verdict serialization per §2.2).
3. `approval.oversee` + revision visibility model.
4. Approval delegation.

## 10. Deferred (registered, bounded)

| Item | Trigger to revisit |
|---|---|
| **W12** parallel stages / DAG routing | first customer requirement for concurrent stage execution; serial covers standard eQMS routing |

## 11. Constraints carried from system-impact analysis (locked)

- Oversight = capability, never a ROOT role (ADR 0022).
- New capabilities ⇒ registry-size bump + tripwire/lint parity (REQ-AUTHZ-5).
- H-PRE-1: no authz-recording read inside lock-holding tx; `authz.Require` needs writable tx (G1).
- Contract-first; expand/contract migrations; watchdog stays alert-only (ADR 0068).
- Canonical test frameworks (testdb factory for DB integration); evidence-based close-out per milestone QA gates.
- **No-fallback principle (operator-locked):** integrity-critical reads (hash chain, signature payloads, authz decisions) never substitute a fallback value — absent/invalid state fails closed with a typed error. Fallbacks that mask impossible states are defects; where two states legitimately need different answers, model them as explicit status-scoped queries, not polymorphic `COALESCE`/default expressions.

## 12. Testing strategy

- **M2b:** integration suites per new behavior on testdb factory — freeze gate (unresolved suggestion/comment blocks), hash pin/echo/verify chain (incl. fail-closed test: pin absent at signoff/publish ⇒ typed error, never head-hash substitution), route-version pinning under concurrent edit, review verdict quorum, SoD single-predicate, delegation signoff dual-identity, visibility 404s (consumer vs participant vs oversee), tier-1 explicit map (no prefix fallback), tripwire parity lints, `TestCapabilityRegistrySize` bump. Live QA: full SOP walkthrough (submit → review suggesting → request_changes → resolve → freeze → sign → publish) against running stack.
- **M2c:** vitest per surface + live QA of the same walkthrough through the real screens; a11y focus pass; empty/error states exercised.
