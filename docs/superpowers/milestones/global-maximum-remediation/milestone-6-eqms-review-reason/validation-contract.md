# M6 validation contract (D4 — binding, authored before implementation)

> **Program:** global-maximum-remediation · **Milestone:** M6 (eQMS periodic review/expiry +
> structured reason-for-change)
> **Authored:** 2026-07-04, **before any implementation** (mission D4). Committed before the first
> code change (F6.2).
> **Binding:** the `milestone-validator` checks the implementation against this document **section by
> section**. Any divergence between shipped code and this contract is **HS-7**: fix the code to the
> contract, or re-open this contract **with operator approval** — never silently edit the contract to
> match the code (mission §9 HS-7). The load-bearing clauses are the **§2 review/expiry column +
> CHECK table**, the **§3 `document.review` capability 10-touchpoint table** (registry 34→35, arm
> generated via M2, negative tripwire proof), the **§4 River surfacer idempotency + tenant isolation +
> published read-port**, and the **§5 structured reason-for-change audit-trail capture**.
>
> **Design rails locked before this contract (D7):** gate verdict 🟡 Yellow
> (`../../../analysis/2026-07-04-m6-eqms-review-reason-system-impact.md`, 93cd6114). This contract is
> the concrete enumeration the gate's 9 locked constraints (§10) imply. Per mission D7 M6 needs the
> gate but not an up-front ADR; the capability-registry bump convention makes an ADR a **gated
> deliverable of F6.2** (§3.2).

---

## 0. Runtime-truth basis (the facts this contract is built on)

All claims traced to source at authoring time (2026-07-04; targeted-verify). Runtime truth beats docs
(CLAUDE.md).

### 0.1 What already exists on `public.documents`

| Column | State today | Anchor |
|---|---|---|
| `effective_from timestamptz NULL` | **present + partially wired** — written at schedule/publish (`SET effective_from = $1` on the draft→scheduled UPDATE) as the effective/target date | `db/baseline/0001_current_schema.sql:1868`; `documents/approval/application/publish_service.go:282` |
| `effective_to timestamptz NULL` | **present but UNWRITTEN** — no code path writes it; this is the "expiry" column to wire | `db/baseline/0001_current_schema.sql:1869`; grep: 0 writes in the publish path |
| `revision_title text` | present — free-text, set at submit (`SET revision_title = $4`) | `db/baseline/0001_current_schema.sql:1893`; `submit_service.go:188,194` |
| `status` | 10-state CHECK; transitions via the M4 unified fn `CanTransitionDocumentStatus` | `submit_service.go:182`; M4 kernel |
| `review_due_at`, `last_reviewed_at`, `reason_for_change`, `reason_category` | **absent** — genuinely new | — |

**Refinement of the gate's constraint #3:** `effective_from` is *already* the effective date at
schedule time — the review's "no effective-date distinct from publish-date" means today
`effective_from` is only ever set = the schedule/publish moment. M6 **reuses** `effective_from` (does
not duplicate it) and **wires the unwritten `effective_to`** as expiry. Only `review_due_at` /
`last_reviewed_at` / `reason_for_change` (+ optional `reason_category`) are new columns.

### 0.2 The submit path (F6.3 capture site)

- `SubmitService.SubmitRevisionForReview` runs one `runner.Do` tx (`submit_service.go:53,86`).
- Free-text `RevisionTitle` is a field of `SubmitRequest` (`:33-43`), normalized (`:113`), written to
  `documents.revision_title` (`:188,194`).
- The audit trail for the submit is the **governance event** `approval_submitted`, emitted **in the
  business tx** via `s.emitter.Emit(ctx, tx, GovernanceEvent{...})` with a JSON `PayloadJSON`
  (`:208-229`). This is the audit sink already on the path.
- The published cross-module audit port also offers `audit.Writer.RecordTx(ctx, tx, Event)` for an
  in-tx audit append (`audit/domain/port.go:137`).

### 0.3 The River periodic-job base (F6.2 surfacer model)

- `maintenance.PeriodicJobs()` returns `[]*river.PeriodicJob` via `river.NewPeriodicJob(
  PeriodicInterval(d), argsFn, &PeriodicJobOpts{ID, RunOnStart})`, args tagged `InsertOpts{Queue:
  "maintenance"}` (`jobs/maintenance/periodic.go:27-51`). Leader-only enqueue ⇒ singleton
  cluster-wide (M5 / ADR 0067).
- The pattern: schedule/args wiring in `maintenance`; the actual Worker + DB live in the job's own
  package under `internal/modules/jobs/<name>`, subscribed only by `metaldocs-jobs`.

### 0.4 Invariants that STAY the last line (not moved to the app)

- **AuthZ = capabilities** — `document.review` tier-2 `authz.Require` in-tx + DB tripwire arm; never a role check.
- **Contract-first** — routes/fields exist only via `openapi.yaml` + regen; no hand-edited generated code.
- **Transactional outbox** — any surfacing side effect enqueues; no inline network call in the job tx.
- **M3 RLS backstop** — the surfacer seeds per-message identity in the jobs binary before tenant-scoped reads.
- **DB enforces invariants** — expiry/review-due CHECKs; mark-reviewed illegal-status rejected by a DB guard consistent with the M4 transition fn.
- **Cross-module via published interface** — `jobs → documents` via a new documents read-port; `documents/approval → audit` via the emitter/`RecordTx`.

---

## 1. Post-milestone target shape (binding)

After M6:

- `public.documents` carries a wired **review/expiry model**: `effective_from` (effective date,
  reused), `effective_to` (expiry, newly wired), `review_due_at` + `last_reviewed_at` (new), all
  nullable (expand-only), with DB CHECKs.
- A **River periodic surfacer** in `metaldocs-jobs` flags documents due for review, reading through a
  **new documents published read-port** — idempotent, tenant-seeded.
- A **`document.review`** capability gates a **mark-reviewed** workflow (contract-first route, tier-1
  + tier-2 authz, M2-generated tripwire arm, M4-routed transition), setting `last_reviewed_at` + the
  next `review_due_at`.
- Revision creation captures a **structured `reason_for_change`** (+ optional `reason_category`) — a
  contract field, not `revision_title` — persisted on the row **and** carried into the audit trail in
  the business tx.
- All new routes/fields are **spec-first** in `openapi.yaml`, regenerated, with a named FE/consumer.

## 2. ★ F6.2 — review/expiry model (binding column + CHECK table)

One forward migration `db/migrations/0NNN_document_review_and_reason.sql` on `public.documents`:

| Column | Type | Constraint | Notes |
|---|---|---|---|
| `effective_from` | (exists) | — | **reused** as effective date; no new column |
| `effective_to` | (exists) | `CHECK (effective_to IS NULL OR effective_from IS NULL OR effective_to > effective_from)` | newly **wired** as expiry |
| `review_due_at` | `timestamptz NULL` | review-due sanity (`review_due_at IS NULL OR effective_from IS NULL OR review_due_at >= effective_from`) | new; NULL = "no review cycle set" |
| `last_reviewed_at` | `timestamptz NULL` | — | new; set by mark-reviewed |

Binding rules:
- **Expand-only** — all columns nullable; legacy rows keep NULL (= no cycle). No backfill.
- **No duplicate column family** — do **not** add `published_at`/`effective_date`/`expiry_date`
  parallel to `effective_from`/`effective_to`. Reusing the existing pair is a **HS-7-checked** clause.
- CHECK constraints are **DB-enforced** — a proof seeds `effective_to <= effective_from` and asserts
  the DB rejects it (not just the app).
- `tenant_id` already present (no new table).

### 2.1 F6.2 model exit criteria

Migration applies on the testdb template; a DB-CHECK integration proof (testdb) seeds (a) a valid
review/expiry row → accepted, (b) `effective_to <= effective_from` → rejected by the DB, (c) a bad
`review_due_at` → rejected. Matches this §2 table.

## 3. ★ F6.2 — `document.review` capability (binding 10-touchpoint table)

New capability **`document.review`** (scope `ScopeTenant` — the mark-reviewed act is tenant-wide like
`document.publish`; confirm at design, ScopeTenant passes areaCode `"tenant"`).

| # | Touchpoint | Binding requirement | Anchor |
|---|---|---|---|
| 1 | const + `validCapabilities` | `CapDocumentReview Capability = "document.review"` registered | `iam/domain/model.go` const block + `validCapabilities` |
| 2 | scope classify | `ScopeTenant`; `TestEveryCapabilityClassified` green | `iam/domain/capability_scope.go` |
| 3 | tier-1 route→cap | the mark-reviewed route mapped (unmapped = silent escalation = FAIL) | `apps/api/cmd/metaldocs-api/permissions.go` |
| 4 | tier-2 in-tx | `authz.Require(ctx, tx, string(CapDocumentReview), "tenant")` after `SeedTxIdentity` | mark-reviewed service |
| 5 | seed grants | granted to the holding roles; system_admin bypasses | `db/reference-data/0001_product_reference_data.sql` |
| 6 | DB tripwire arm | the `documents` UPDATE tripwire arm accepts `document.review`, **generated from the registry via M2 (F2.1)** — never a hand-edited `TEXT[]`; regen + drift check green | M2 generated arms |
| 7 | guard tests | `TestEveryCapabilityClassified`, `TestAreaGradeCapabilitySet` green | iam domain tests |
| 8 | registry size | **`TestCapabilityRegistrySize` 34 → 35** | `iam/domain/model_test.go:96` |
| 9 | capability-coherence | const/classify/tier-1/seed/test agree (M2 lints, REQ-AUTHZ-5) | M2 CI lints |
| 10 | H-PRE-1 | the mark-reviewed `authz.Require` recording read takes **no** advisory lock (satisfied by construction; keep off any lock) | `[[advisory-lock-deadlock-constraint]]` |

### 3.1 Negative authz proof (binding)

An integration proof asserts: (a) mark-reviewed **succeeds** only with `document.review`; (b) with the
capability **withheld**, the call is denied; (c) a write that reaches the `documents` UPDATE **without
asserting** `document.review` trips the DB tripwire (P0001) — proving the arm + assert are wired, not
decorative.

### 3.2 ADR (binding — gated deliverable of F6.2)

**One ADR** authored with F6.2 (required by the `model_test.go:96` "bump only via ADR" convention +
the standing product-policy addition): records the review-cycle model, effective-date-vs-publish-date
semantics (reusing `effective_from`/`effective_to`), the `document.review` capability, and the
surfacing-via-River decision. Must be **Accepted** + indexed before milestone close.

## 4. ★ F6.2 — River surfacer + published read-port (binding)

### 4.1 The read-port (cross-module boundary)

A **new documents published read-port** (e.g. `ReviewDueReader.ListDueForReview(ctx, tx, now, limit)`
on the documents module) returns due documents (`review_due_at <= now()`), tenant-scoped. `jobs` calls
**this port** — **never** raw SQL on `public.documents` (invariant 6). Census: no `documents` table
SQL in the `jobs` module for the surfacer.

### 4.2 The periodic job

- A River periodic job (mirrors `maintenance.PeriodicJobs()` — `NewPeriodicJob(PeriodicInterval(d),
  argsFn, &PeriodicJobOpts{ID:"document-review-surfacer", RunOnStart:false})`, `InsertOpts{Queue:
  "maintenance"}` or a dedicated queue), hosted in `metaldocs-jobs`. **No hand-rolled scheduler.**
- **Tenant seed** — the job seeds per-run identity (`SeedTxIdentity`/`SeedTxTenant`, M3 backstop)
  before the tenant-scoped read.
- **Idempotent** — surfacing is idempotent: running the job twice surfaces a due doc **once** (flag is
  a set/`ON CONFLICT` write or an idempotent state read; no duplicate side effect). If surfacing emits
  a notification, it enqueues via the **outbox** (never inline).

### 4.3 F6.2 surfacer exit criteria

Surfacer integration proof (testdb, real Postgres): seed a `review_due_at <= now()` doc in tenant A +
a not-yet-due doc + a due doc in tenant B → run the job → tenant-A due doc surfaced, not-yet-due
untouched, **tenant-B doc not surfaced under tenant-A identity** (cross-tenant isolation); run **twice**
→ surfaced **once** (idempotent). Matches this §4.

### 4.4 Mark-reviewed workflow

The mark-reviewed route (contract-first, §6) sets `last_reviewed_at = now()` and the next
`review_due_at`, routed through the **M4 unified transition function** (friendly first-line legality
check mirroring the DB trigger, like `submit_service.go:182`) + an OCC/`revision_version` CAS. Illegal
status is rejected by the DB guard, not app code alone.

## 5. ★ F6.3 — structured reason-for-change (binding)

### 5.1 Contract + column

- **Contract:** `reason_for_change` (string) + optional `reason_category` (enum) added to the
  **submit-revision request schema** in `openapi.yaml`; regen BE (`SubmitRequest`/generated DTO) + FE.
  **Not** `revision_title` — the free-text title stays; the structured reason is a distinct field.
- **Column:** `+ reason_for_change text NULL` (+ optional `reason_category text NULL` with a CHECK
  enum) on `public.documents` (expand-only, same migration as §2).

### 5.2 Capture + audit trail (binding)

- The field is threaded into `SubmitRequest` and, inside the submit business tx, **persisted on
  `public.documents`** (extend the existing draft→under_review UPDATE at `submit_service.go:185`, NOT a
  second write, NOT reusing `revision_title`).
- The structured reason is carried into the **audit trail in the business tx** — via the existing
  `approval_submitted` governance-event payload (`emitter.Emit`, `submit_service.go:208-229`, add
  `reason_for_change`/`reason_category` to `payloadMap`) and/or `audit.Writer.RecordTx`. The binding
  requirement is: **one audit event on the submit carries the structured reason** (21 CFR Part 11
  attributable change reason). Mechanism = the existing in-tx emit path (no new inline network call).

### 5.3 API requiredness (expand/contract)

- **Required at the API for REV≥1** (friendly first line): submitting a REV≥1 revision **without**
  `reason_for_change` is rejected with RFC 9457 `problem+json` (422). REV 0 (initial creation) follows
  the existing `revision_title` default convention (`defaultInitialRevisionTitle`) — reason optional at
  REV 0 unless design decides otherwise.
- **Nullable in DB** for legacy rows (expand/contract; no backfill).

### 5.4 F6.3 exit criteria

Revision-creation proof (testdb + live): submitting a revision with `reason_for_change` (a) persists it
on `public.documents`, (b) emits **one** audit event whose payload carries the reason (+ category); a
REV≥1 submit **without** the field is rejected at the API (422 problem+json); no code path writes the
structured reason into `revision_title`. Pin tests for the new generated request field. Matches §5.

## 6. Contract-first discipline (binding — cross-feature)

- Every new route/field lands in `api/openapi/v1/openapi.yaml` **first**, then `go generate` regens BE
  + FE; **zero hand-edits** to generated code. New shapes: (a) `reason_for_change`/`reason_category` on
  submit request; (b) `review_due_at`/`effective_from`/`effective_to`/`last_reviewed_at` on document
  response DTO(s); (c) a **mark-reviewed** operation; (d) a review-due **list filter**.
- `oasdiff` (M1 gate) green; pin tests for each new request/response shape; a **named FE/consumer**
  consumes each new shape (the documents view for response fields + review-due filter; the
  revision-submit form for the reason field).
- **HS-7:** any forced change to an **existing** generated shape (breaking, not additive) **stops and
  surfaces** — never a silent hand-edit of generated code or the spec's meaning. All M6 additions are
  additive/optional at the wire (new fields nullable; reason required only at the handler for REV≥1).

## 7. DB-as-last-line + cross-feature constraints (binding)

- **AuthZ** — `document.review` tripwire arm generated via M2; negative proof (§3.1) shows it fires.
- **Multi-tenant** — new columns on the already-tenant-scoped `documents`; surfacer seeds identity in
  the jobs binary (M3); cross-tenant read → 0 rows / 404.
- **Async = outbox** — surfacer side effects enqueue; no inline network call in the job tx.
- **DB enforces** — §2 CHECKs; mark-reviewed illegal-status rejected by a DB guard consistent with M4.
- **Cross-module via port** — `jobs → documents` via the new read-port; `documents/approval → audit`
  via the emitter/`RecordTx`. No module reaches another's tables.
- **M4 unified transition** — mark-reviewed routes through `CanTransitionDocumentStatus` (or the
  unified fn), no scattered `if status !=`.
- **Subagent-driven implementation** (`superpowers:subagent-driven-development`; sonnet
  implement/review, haiku mechanical, never fable, ≤15 concurrent); main session
  orchestrates/reviews/commits; the `milestone-validator` judges + writes `qa/milestone-qa.md`; main
  flips status only on PASS.
- **testdb factory** for every proof (real Postgres for surfacer + CHECK + audit-event proofs, not
  sqlmock); **targeted `-run`** only — the full integration suite is NOT run locally (20-min box);
  bounded defers recorded in `evidence.md` with triggers.
- **All 4 binaries build** + `.\scripts\check-system-runnable.ps1` + a **live QA drive** (due-review
  cycle + structured reason capture on the audit trail) at close.
- **Commit after verified work; never push; never commit `docs/release/`; plans dir gitignored (never
  force-add).**

## 8. Bounded defers (recorded, with triggers)

| Defer | Rationale | Trigger / owner |
|---|---|---|
| Training acknowledgment / obligated-reader review-attestation | `distribution` owns it; finding 14 scopes it out | A distribution-module milestone, if a consumer contract asks |
| Backfilling historical `reason_for_change` onto legacy rows | Intent can't be reconstructed; would be fabrication | Never (legacy rows stay NULL by design) |
| Notification/escalation on review-overdue (beyond surfacing) | M6 surfaces + gates the workflow; escalation policy is broader | M8 ops-readiness or a product decision |
| Full `-tags integration` run of the M6 proofs on the local box | 20-min box constraint (mission §10) | CI / capable box before program close; targeted `-run` drives authored regardless (M1–M5 precedent) |
