# System-impact analysis — M6 eQMS: periodic review/expiry + structured reason-for-change

**Date:** 2026-07-04
**Intent (one line):** Land the two ISO-core eQMS product gaps (finding 14): document periodic **review-due / effective-date / expiry** with scheduled surfacing + a capability-gated review workflow (F6.2), and **structured reason-for-change** captured at revision creation and carried into the audit trail (F6.3).
**Work type:** feature (two features across existing modules — no new module)
**Author:** developing-new-work skill (F6.1 gate for milestone-6-eqms-review-reason)
**Verdict:** 🟡 **Yellow** *(see §10)*
**Standards driving the work:** ISO 9001 §7.5.3 (control of documented information — periodic review); 21 CFR Part 11 (attributable, structured change reason on the audit trail).
**Dependency:** M5 async-River consolidation base (operator-approved HS-1 2026-07-04) — F6.2's scheduled surfacing rides `internal/modules/jobs/maintenance` River periodic jobs.

> Same ten sections for module and feature work. Module-only rows are marked **N/A** with a one-line reason.

---

## 1. Classify & own
*(CLAUDE.md Orientation rule)*

- **Work type:** feature (two: F6.2 review/expiry, F6.3 reason-for-change). No new module.
- **Owning module(s):**
  - **`documents`** — owns the `public.documents` instance row (the governed revision: `status`, `revision_number`, `revision_version`, `revision_title`, `effective_from/effective_to`, `controlled_document_id`). Both the review/expiry **dates** (F6.2) and the reason-for-change **field** (F6.3) are properties of this row. Domain enum `DocumentStatus` (`internal/modules/documents/domain/model.go:8`).
  - **`documents/approval`** (sub-package) — owns the **revision-creation path**: `SubmitService.SubmitRevisionForReview` (`internal/modules/documents/approval/application/submit_service.go:43`), where `RevisionTitle` free-text is set today and the reason-for-change field (F6.3) is captured; and the decision/publish path (`decision_service.go`) that sets `effective_from` — where the review-cycle clock (F6.2) starts.
  - **`jobs`** — owns the River **periodic surfacer** for F6.2 (a new periodic job beside `maintenance.PeriodicJobs()`, `internal/modules/jobs/maintenance/periodic.go:22`).
  - **`iam`** — owns the new capability `document.review` (registry in `internal/modules/iam/domain/model.go`).
- **Explicitly NOT owning:**
  - **`controlleddocuments`** — owns the CD *identity* + its own lifecycle (`active|obsolete|superseded`, `domain/controlled_document.go`). Review-due/effective/expiry attach to the **published Document revision**, not the CD slot; reason-for-change is set at revision submit, which is a documents/approval concern. CD creation calls the documents port, but neither M6 feature mutates CD-owned tables. (Watch: `CreateRevision` also exists on the CD service `service.go:584` — but the *governed submit* that sets `revision_title` is the documents/approval path; F6.3 lands there, not on the CD service.)
  - **`distribution`** — owns training-acknowledgment / obligated-reader tracking. **Out of scope by decision** (finding 14: "training acknowledgment legitimately out-of-module").
  - **`audit`** — not owning; it is a *consumer sink* (F6.3 writes a reason-carrying event via its published `Writer.RecordTx` port, no reach into audit internals).
- **Cross-module edges (with direction):** `A → B` = A depends on B (via B's published Go interface).
  - `documents/approval → audit` — `audit.Writer.RecordTx` (published port `internal/modules/audit/domain/port.go:123`). ✔ interface, not tables.
  - `jobs → documents` — the review-due surfacer must read documents whose `review_due_at ≤ now()` and enqueue/flag. **Must go through a documents published read-port**, not raw SQL on `public.documents` from the jobs module (invariant 6). New port method on documents (e.g. `ReviewDueReader.ListDueForReview`) — design decision, flagged to brainstorming.
  - `documents → iam/authz` — `authz.Require(ctx, tx, CapDocumentReview, "tenant")` on the mark-reviewed path. ✔ existing helper.
- **Ambiguity?** Resolved — no AS-3. The documents-vs-controlleddocuments boundary was the one real question; verified by schema + domain reads: the review/expiry columns and `revision_title` live on `public.documents` (`db/baseline/0001_current_schema.sql:1868-1893`), so documents owns both features.

## 2. Foundation verdict
*(Global-Maximum rule)*

- **Base you'd build on:** the documents versioning kernel — **just unified in M4** into a single exhaustive 9-status transition function (`CanTransitionDocument`), publish-race proven safe/choked, ADR 0066 concurrency idiom ratified. The async base is M5's River consolidation (janitors + retention as River periodic jobs, operator-approved). The audit hash-chain writer + `Writer.RecordTx` port are stable.
- **Sound, or legacy/patch/workaround?** **Sound.** Building review/expiry + reason-for-change on top of a freshly-hardened state machine and a freshly-consolidated scheduler is the opposite of optimizing inside a patch — the two milestones that would have been the shaky base (M4 kernel, M5 scheduler) were just brought to a global maximum precisely so product features could land cleanly. No AS-2.
- **One reuse-not-reinvent note (global-max):** `effective_from`/`effective_to` **already exist** on `public.documents` (`:1868-1869`, nullable) but the review (finding 14) says "no effective-date distinct from publish-date" — i.e. columns present, semantics unwired. The global-maximum move is to **reuse and wire these columns** (effective_from = effective date, effective_to = expiry) rather than add parallel `published_at`-vs-`effective_date` columns. Only genuinely new state (`review_due_at`, `last_reviewed_at`, `reason_for_change`) gets new columns. → locked constraint; runtime-verify the current publish-path wiring of `effective_from` at design time.

## 3. Invariant alignment
*(the 6 non-negotiables)*

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | **Yes** | New capability `document.review` for the mark-reviewed workflow; tier-2 `authz.Require` in-tx; **never** "QA role can review". Reason-for-change capture rides existing `document.submit`. | `authz.Require(ctx,tx,cap,"tenant")` (`iam/authz/authz.go:76`); registry `iam/domain/model.go` |
| Contract-first (OpenAPI + oapi-codegen) | **Yes** | New route(s) (mark-reviewed; review-due fields on document responses; review-due filter/list) + new request field(s) (reason-for-change on submit) added to `api/openapi/v1/openapi.yaml` first, then regenerate. Zero hand-edits to generated. | `api/openapi/v1/openapi.yaml` + module `cfg.yaml`/`gen.go` |
| Multi-tenant pooled (`tenant_id`/GUC/404) | **Yes** | All new reads/writes carry `tenant_id`; GUC auto-seeded at the TxRunner chokepoint (M3); the jobs surfacer seeds per-message identity (M3 async backstop). New columns are on the already-tenant-scoped `documents` table (no new table). | `authz.SeedTxIdentity` via M3 chokepoint; `tenant.FromContext` |
| Async = transactional outbox | **Yes** | F6.2 surfacing is a **River periodic job** (M5 base), not an inline network call. If surfacing produces a notification, it enqueues via the outbox — the job reads due docs and enqueues; the consumer does any side effect idempotently. | `maintenance.PeriodicJobs()` (`jobs/maintenance/periodic.go:22`); outbox repo (`render/fanout/staging_outbox.go:29`) |
| DB enforces invariants (triggers/constraints) | **Yes** | `review_due_at`/`effective_to` sanity via CHECK (e.g. expiry > effective); mark-reviewed permitted only in valid statuses (align with the M4 transition fn, not a scattered guard); the documents **capability tripwire** must accept `document.review` on the mark-reviewed UPDATE — **arm generated from the registry via M2**, never hand-typed. | M4 `CanTransitionDocument`; M2 generated tripwire arms; `ck_*` CHECK pattern |
| Cross-module via published interface only | **Yes** | `jobs → documents` review-due read goes through a new documents published port; `documents/approval → audit` via `Writer.RecordTx`. No module reaches into another's tables. | documents `domain/port.go`; `audit/domain/port.go:123` |

No violation → **no AS-1**.

## 4. Capability wiring
*(F6.2 adds one capability — walk the 10 touchpoints)*

New capability: **`document.review`** (scope: `ScopeTenant`; the mark-reviewed act is tenant-wide like `document.publish`, not area-graded — confirm at design).
1. **const + `validCapabilities`** — add `CapDocumentReview Capability = "document.review"` (`iam/domain/model.go:75-117` const block; register in `validCapabilities` `:127-162`).
2. **scope classify** — `ScopeTenant` (`capability_scope.go`); keeps `TestEveryCapabilityClassified` green.
3. **tier-1 route→cap** — map the new mark-reviewed route in `apps/api/cmd/metaldocs-api/permissions.go` (unmapped = silent privilege escalation).
4. **tier-2 in-tx** — `authz.Require(ctx, tx, CapDocumentReview, "tenant")` in the mark-reviewed service, after `SeedTxIdentity`.
5. **seed grants** — grant `document.review` to the roles that hold it (`db/reference-data/0001_product_reference_data.sql`); system_admin bypasses.
6. **DB tripwire** — the `documents` UPDATE tripwire arm must accept `document.review`; **generated from the registry via M2 (F2.1)** — not a hand-edited `TEXT[]`. Regenerate arms + CI drift check green.
7. **guard tests** — `TestEveryCapabilityClassified`, `TestAreaGradeCapabilitySet` stay green.
8. **bump `TestCapabilityRegistrySize`** — **34 → 35** (verified current = 34, `model_test.go:96`). The test comment mandates *"bump only via ADR"* → this is why an ADR is required (see §9).
9. **CI capability-coherence (REQ-AUTHZ-5)** — const/classify/tier-1/seed/test surfaces must agree (M2 lints).
10. **H-PRE-1** — the mark-reviewed `authz.Require` recording read must not sit inside a lock-holding atomic tx; the path takes no advisory lock, so satisfied by construction. Keep it off any lock.

## 5. Module wiring
**N/A** — no new module is born. Both features extend `documents`/`documents.approval` and add one River job in `jobs`; all have existing module wiring, ports, and composition-root entries.

## 6. Frameworks to reuse, not reinvent
*(frameworks-catalog)*

- `TxRunner` (`Do`) — the mark-reviewed write and the reason-carrying submit run in a `TxRunner` tx; services depend on the tx port. ✔
- `authz.SeedTxIdentity` / `authz.Require` — GUC seed + tier-2 check. ✔ (no hand-rolled role check)
- `tenant.FromContext` — tenant read; never thread by hand. ✔
- `audit.NewEvent` / `Writer.RecordTx` — F6.3 reason-for-change written **in the business tx** via `RecordTx` (`audit/domain/port.go:123`); payload JSON carries `reason_for_change` (+ category). ✔
- Outbox repo — any surfacing side effect (notification) enqueues, not inline. ✔
- River periodic job — F6.2 surfacer mirrors `maintenance.PeriodicJobs()` / `retention.PeriodicJob()` (`render/fanout/retention/periodic.go:26`); **do not hand-roll a scheduler** (that would re-introduce exactly what M5 retired). ✔
- `problem.New`/`Write` — new routes return RFC 9457. ✔
- `testdb.Open` factory — all new integration tests. ✔

No hand-rolled equivalent of any primitive. No genuinely-new cross-cutting concern → no new platform framework needed.

## 7. Contract & data

- **OpenAPI-first:** add to `api/openapi/v1/openapi.yaml` (documents partial): (a) `reason_for_change` (+ optional `reason_category` enum) on the **submit-revision** request schema; (b) `review_due_at` / `effective_from` / `effective_to` / `last_reviewed_at` on document response DTO(s); (c) a **mark-reviewed** operation (e.g. `POST /documents/{id}/review` or `PATCH …/review-due`); (d) a review-due **filter** on the documents list. Regenerate BE (`gen.go`) + FE types. **HS-7 discipline: if any existing generated shape must change, stop and surface — never silently hand-edit generated code or the spec's meaning.**
- **Migration:** one forward migration `db/migrations/0NNN_document_review_and_reason.sql` on `public.documents`: `+ review_due_at timestamptz NULL`, `+ last_reviewed_at timestamptz NULL`, `+ reason_for_change text NULL`, optional `+ reason_category text NULL` with a CHECK enum. **Reuse existing `effective_from`/`effective_to`** — do not add duplicate published/effective/expiry columns. Add CHECK(s): expiry (`effective_to`) > `effective_from`; `review_due_at` sanity. No new table ⇒ `tenant_id` already present.
- **DB invariants:** mark-reviewed allowed only from published/effective states — route through the **M4 unified transition function**, plus a DB guard consistent with it. The tripwire arm for `documents` UPDATE gains `document.review` **via M2 generation**.
- **Destructive change?** None. All new columns nullable (expand-only); backfill not required (legacy rows have NULL review-due = "no cycle set"). Reason-for-change is required **going forward** for REV≥1 at the API layer (friendly first line), nullable in DB for legacy rows — expand/contract respected.

## 8. Test & QA plan
*(test-qa-gates)*

- **Canonical framework:** `testdb` integration factory (`tests/integration/testdb/`, `//go:build integration`), R1–R4 discipline. No bespoke harness.
- **QA gates that apply (feature subset):**
  - **Contract** — spec↔generated↔handler alignment; new fields/route present; `oasdiff` (M1) green; pin tests for the new request/response shapes.
  - **AuthZ** — `document.review` reachable only with the capability; tripwire fires without the arm/assert (negative); registry size 35; M2 drift check green.
  - **DB-invariant** — CHECK constraints (expiry>effective; review-due) reject bad rows; mark-reviewed rejected from illegal status by the DB, not just app.
  - **Async/idempotency** — the review-due surfacer is idempotent (running twice surfaces once); River periodic-job proof (scheduled tick surfaces a due doc).
  - **Multi-tenant** — surfacer + mark-reviewed are tenant-isolated (cross-tenant → 404 / 0 rows); GUC seeded in the jobs binary (M3 backstop).
  - **Docs** — wiki `documents.md` + `controlled-documents.md` (cross-link) + DB table docs refreshed; ADR authored.
- **Evidence shape:** `go build ./...`; targeted `go test -run` (NOT the full 20-min integration box — bounded defers recorded); `.\scripts\check-system-runnable.ps1`; **live/preview QA drive (mandatory, runtime-visible)** — start API via `.\scripts\start-api.ps1 -Build`, drive a due-review cycle (set review-due → surfacer flags → capability-gated mark-reviewed) and a structured reason-for-change capture at revision submit; capture proof (network/logs/DB row + audit event). Then `milestone-validator` → `qa/milestone-qa.md`.

## 9. Docs / ADR
*(docs-adr-governance)*

- **Wiki (feature → update existing, refresh `Last verified`):** `wiki/modules/documents.md` (new review/reason fields, mark-reviewed route, surfacer job), `wiki/modules/controlled-documents.md` (cross-link note), `wiki/modules/jobs.md` (new periodic job), `wiki/database/tables/documents.md` (new columns). No new module doc.
- **REQ IDs cited:** the eQMS/versioning REQ IDs in `wiki/architecture/backend-target-architecture.md` for finding 14 (dimension 6 product gaps) — cite in the milestone spec and each feature evidence.
- **ADR required? YES (Yellow driver).** Two reasons: (1) the capability-registry bump convention mandates *"bump only via ADR"* (`model_test.go:94`); (2) periodic review/expiry + structured reason-for-change is a **standing product-policy addition** (a new document-lifecycle obligation), which the governance rule classifies as ADR-worthy. → author **one ADR** ("Document periodic review/expiry + structured reason-for-change") recording: the review-cycle model, effective-date-vs-publish-date semantics (reusing `effective_from`/`effective_to`), the `document.review` capability, and the surfacing-via-River decision. D7 does **not** mandate an ADR for M6 (only M5/M7), but the registry convention does — so it is required in practice, not optional. Not a MUST-deviation from the target spec (no HS re: contract).

## 10. Verdict & locked constraints

- **Verdict:** 🟡 **Yellow** — the work fits cleanly (no invariant violated, owning modules unambiguous, foundation sound, frameworks all reused), but it carries named constraints and a required ADR, so it proceeds *with rails*, not free.
- **Open hard-stops:** **none.** AS-1 (invariant) — none. AS-2 (patch foundation) — none (M4/M5 hardened the base first). AS-3 (ambiguous owner) — resolved (documents owns both; controlleddocuments explicitly does not).
- **Locked constraints handed to the milestone plan / brainstorming:**
  1. **New capability `document.review`** ⇒ walk all 10 touchpoints; **bump `TestCapabilityRegistrySize` 34 → 35**; tripwire arm **generated via M2 (F2.1)**, never hand-typed.
  2. **One ADR required** (capability-bump convention + policy addition) — authored with the milestone.
  3. **Reuse existing `effective_from`/`effective_to`** for effective-date/expiry — do **not** add duplicate columns; runtime-verify the current publish-path wiring of `effective_from` before design.
  4. **F6.2 surfacing = River periodic job** on the M5 base (`maintenance.PeriodicJobs()` pattern) — no hand-rolled scheduler; idempotent; tenant-seeded in the jobs binary (M3).
  5. **`jobs → documents` via a new published read-port**, not raw SQL (invariant 6).
  6. **Contract-first**: openapi.yaml + regen, zero hand-edits to generated; **HS-7** — any forced change to an existing generated shape stops and surfaces.
  7. **F6.3 reason-for-change**: structured field(s) at revision submit (`submit_service.go:43` `SubmitRequest`), captured into the audit trail via `Writer.RecordTx` in the business tx; required at API for REV≥1, nullable in DB for legacy (expand/contract).
  8. **Route mark-reviewed through the M4 unified transition function** + a consistent DB guard — no scattered `if status !=` lifecycle checks.
  9. **Live/preview QA drive mandatory** (runtime-visible milestone) as contract evidence.

→ **Green/Yellow ⇒ proceed.** Hand these rails to the milestone plan (`milestone` skill) + `validation-contract.md` (D4) before implementation. Not Red — design is **not** blocked.
