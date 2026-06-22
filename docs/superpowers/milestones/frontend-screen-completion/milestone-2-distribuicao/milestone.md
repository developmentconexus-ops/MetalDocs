# Milestone 2 — Distribuição coverage-scope (full-stack, derive-on-read)

> **Program:** frontend-screen-completion  ·  **Governing spec:** `../mission.md`
> **Status:** Spec (drafting) — **awaiting operator approval (HS-1 gate) before any feature begins.**
> **Authored:** 2026-06-21 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is,
> **which features** it contains, **what each feature implements**, and **what gets
> validated**. It contains **no execution steps** — the "how" of each feature lives in
> that feature's `plan.md`. The end-of-milestone QA (`qa/milestone-qa.md`) validates the
> milestone against *this* document.

## Objective

After M2, the Distribuição & Cobertura screen (`/documents/:documentId/distribution`) renders the
**real obligated-reader set** for a controlled document — who must read it, broken down by area, with
a true total — sourced from a new **Grade-A, read-only** backend endpoint. The illustrative
scaffolding (`MOCK_DISTRIBUTION`, watermarks, aria-hidden blocks) is removed for those surfaces and
replaced by live data. The operator opening the screen sees **a real recipient list and real
coverage totals**, never illustrative placeholders.

The read/acknowledge **numerator** (read %, acknowledged %, overdue, adoption timeline) and the
**action layer** (mass reminder, export, add recipients, fanout policy) are **out of scope for this
milestone** — that data domain does not exist at baseline and is parked as a designed mission
([`wiki/backlog/document-distribution-mission.md`](../../../../wiki/backlog/document-distribution-mission.md)).
Those surfaces render an explicit, honest **"tracking pending"** state (not illustrative data, not a
fabricated `0%`), and the hero CTAs stay **deferred-with-trigger** pointing at that mission.

**Quality bar moved:** the program's "no illustrative / `MOCK_`/'em breve' data on a routed screen"
bar (`wiki/quality/screen-definition-of-done.md` D2) goes from **violated** on Distribuição (the whole
screen is illustrative scaffolding today) to **met for the denominator surfaces** — the obligated set
is live, the old `Dados ilustrativos · Em breve` markers are gone, and the numerator gap is honestly
disclosed rather than faked. Re-measured by:
`grep -rEn "Dados ilustrativos|MOCK_DISTRIBUTION|Em breve" …/pages/DocumentDistributionPage.tsx` = 0,
plus runtime proof the recipient list + by-area totals render live values.

### Operator-locked scope decision (2026-06-21)

The mission §7 framed M2 as building a "distribution/**fanout**/coverage endpoint … live coverage."
Runtime truth at authoring time (recon this session, HEAD `d477e9f0`): the **denominator** (the
obligated reader set) is **derivable** from existing visibility grants + area membership
(`controlled_document_area_grants`, `controlled_document_user_grants`, `document_process_areas`,
`user_process_areas`, view `metaldocs.v_active_user_areas`), but the **numerator** — any read event,
acknowledgement event, distribution target, or reminder job — **does not exist anywhere**.
`approval_signoffs` is pre-publish approval only and must not be repurposed as a read/ack signal. A
"coverage" endpoint built today therefore has a real denominator and **nothing real to count**.

The operator was shown the gap and **chose to split M2**:
- **M2 (this milestone)** builds the **derive-on-read coverage-*scope*** subset: a Grade-A, **read-only**
  endpoint serving the obligated set (total + recipients + by-area totals) from the existing tables.
  **No new table, no migration, and zero change to the publish path** (`PublishApproved()` is not
  touched — snapshot-at-publish is explicitly refused here, see rabbit holes / HS-2).
- **The full numerator + action domain** (read/ack tables, reader-side acknowledge surface,
  snapshot-vs-derive decision, reminders/export/fanout worker) is **parked** as a designed mission,
  [`document-distribution-mission.md`](../../../../wiki/backlog/document-distribution-mission.md), to
  execute **after** the frontend-screen-completion mission finishes.

This re-scope is the approved F2.x consumer contract; each feature's `spec.md` distills it. The
feature slugs are renamed from the mission's `fanout-*` to `coverage-*` to reflect that the action
("fanout") layer is parked. Reconciling the mission §7/§8 wording ("fanout", "zero em breve" for the
whole screen) is a **doc-only `wiki-curator` follow-up** (non-blocking) — the screen reaches full
mission-`§8` completeness in two stages: denominator now, numerator when the parked mission runs.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F2.1 | `f2.1-coverage-contract` | ADR + OpenAPI spec for the **denominator-only** coverage endpoints, consumer-contract-first from `DocumentDistributionPage`: `GET /documents/:id/distribution` (→ `total_targets` + document identity facts), `GET /documents/:id/distribution/recipients?page=&pageSize=` (→ obligated recipients: `user_id`, `name`, `area`, `source`; `X-Total-Count` + `Link` headers), `GET /documents/:id/distribution/coverage` (→ by-area `{area_code, area_name, total}`). **No numerator fields** (`read`/`acknowledged`/`overdue`/`pending`/timeline) — the parked mission extends the contract additively later. Regen FE types. | `api-lint -strict` parses the new paths = **0** violations; the response schemas contain **no** read/ack/overdue field; ADR recorded under `wiki/decisions/`; generated FE types present in `lib/api-types/`; consumer (`spec.md`) shape == the generated producer types. |
| F2.2 | `f2.2-coverage-backend` | Implement the three endpoints to the Grade-A bar as a **read-only projection** over existing grant/area tables + `v_active_user_areas` (obligated set = user-grants ∪ active members of granted/process areas). New tier-2 capability via `authz.Require` (+ `trg_require_cap_asserted`). **No new table, no migration, no publish-path change.** | Integration test against live PG asserts the obligated set against a fixtured grant/area graph (union + active-membership correct; fail-closed on an unhandled grant type); `api-lint -strict` = **0**; all **6 CI guards** green; `go build ./...` / `go vet ./...` / `go test ./...` green; `git diff db/migrations` = empty; `git diff …/approval/application/publish_service.go` = empty. |
| F2.3 | `f2.3-distribuicao-wire` | Wire `DocumentDistributionPage` to the real coverage endpoints via TanStack Query hooks; **remove** the illustrative scaffolding (`MOCK_DISTRIBUTION`, watermarks, aria-hidden) for the **denominator** surfaces (total, recipient list, by-area). Render the **numerator** surfaces (read/ack donut, timeline, per-recipient read/ack columns, overdue/pending) as an explicit, labeled **"tracking pending — leitura/ciência ainda não registrada"** state. Keep the 4 hero CTAs + row/bulk actions **`aria-disabled` deferred-with-trigger** linking the parked mission. | `grep -nE "Dados ilustrativos\|MOCK_DISTRIBUTION\|Em breve" …/DocumentDistributionPage.tsx` = **0**; recipient list + by-area totals render live values (consumer shape == generated types); the "tracking pending" state renders (truthful empty, not fabricated, not illustrative); CTAs disabled with a trigger note; a query-hook test passes against fixtured responses; **`frontend-screen-reviewer` + `frontend-code-reviewer` APPROVE on record** (D2); FE tests green. |

Order: **F2.1 → F2.2 → F2.3** — contract first (consumer-contract-first; unlocks codegen), then the
producer, then the wire. The screen wire (F2.3) is last and is the feature whose evidence closes the
quality-bar claim. Each "what to validate" is objectively checkable (a grep = 0, an `api-lint` count,
a contracted shape, a passing integration/hook test, a reviewer APPROVE-on-record), never "works".

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. What that gate
enforces for this milestone:

1. **Per-feature acceptance** — F2.1/F2.2/F2.3 each meet their declared "what to validate", and each
   feature's **consumer contract** (`spec.md`) was honored: the FE consumer shape matches the real
   generated producer types (denominator-only; no read/ack fields anywhere in the contract).
2. **Workflow-class QA** — **FE:** `wiki/quality/screen-definition-of-done.md` (D2: both reviewer
   APPROVEs on record) + the runtime functional pass by reference to `wiki/quality/screen-qa-checklist.md`.
   **BE:** `wiki/quality/backend-api-qa-checklist.md` + the 6 CI guards + `api-lint -strict` = 0 +
   the integration test.
3. **Regression** — **M0 + M1** still pass their gates (single index route, 0 dead-stub routes, M1
   dashboard greps hold); the FE suite holds at the operator-accepted baseline; **the publish path is
   untouched** (`git diff internal/modules/documents/approval/application/publish_service.go` = empty);
   **no new migration** (`git diff db/migrations` = empty); `go build ./...` / `go test ./...` green.
4. **Quality-bar re-measure (root cause, not symptom)** —
   `grep -rEn "Dados ilustrativos|MOCK_DISTRIBUTION|Em breve" …/pages/DocumentDistributionPage.tsx`
   = **0**; the obligated set (recipients + by-area + total) renders **live** at runtime (illustrative
   scaffolding deleted at root, not flag-hidden); the numerator surfaces render the explicit
   **honest "tracking pending"** state — not a fabricated metric, not the old illustrative block; the
   new endpoint serves real denominator data (proven by the F2.2 integration test).
5. **No unplanned scope** — only F2.1 + F2.2 + F2.3 are implemented; **no** numerator producer, **no**
   new table/migration, **no** publish-path change, **no** action layer (CTAs/reminders/export). Any
   pull toward those is routed to the parked mission, not built here. Anything beyond the three
   features is recorded with rationale.

## Dependencies & constraints

- **Depends on:** M0 (truth-reset) + M1 (dashboard) passed; backend Grade-A baseline `d477e9f0`; the
  existing grant/area tables + `metaldocs.v_active_user_areas`; the already-honest
  `DocumentDistributionPage` scaffolding as the wire target.
- **Appetite:** medium — one full-stack slice: a **read-only** backend projection + contract + screen
  wire. Larger than M1 (real backend), deliberately bounded well below the parked mission (no new
  table, no worker, no publish-path change, no numerator, no action layer).
- **Quality goals (ranked):** 1) **truthfulness** (render only the obligated set a real producer
  derives; the numerator is honestly disclosed as pending, never faked — no invented read %, deadline,
  overdue, or pending metric) > 2) **contract-correctness** (consumer shape == generated producer
  types; the denominator-only contract is **forward-compatible** — the parked mission extends it
  additively, never breaks it) > 3) **simplicity** (smallest backend that serves real data: a
  read-only projection over existing tables, reusing existing query/client patterns).
- **Architectural constraints (validator can fail on these):**
  - **Backend is read-only:** **no new table, no migration, and zero change to `PublishApproved()`**
    or any publish-path code. Snapshot-at-publish is **refused** here (it touches the Grade-A atomic
    publish tx → HS-2 + parked-mission scope). Derive the denominator live from existing tables.
  - **Contract-first regen order:** spec → OpenAPI → `oapi-codegen` → FE types; the FE consumes the
    **generated** types only, never hand-rolled shapes (HS-3).
  - **Numerator + action layer are out of scope** — parked to
    [`document-distribution-mission.md`](../../../../wiki/backlog/document-distribution-mission.md);
    rendered as honest "tracking pending" / deferred-with-trigger, **never fabricated**.
  - **Grade-A backend bar:** the new endpoint passes `api-lint -strict` = 0, all 6 CI guards, and an
    integration test against live PG; `go build`/`vet`/`test` green.
  - **authz:** gate the new reads with a tier-2 capability (`authz.Require` + the
    `trg_require_cap_asserted` tripwire). **H-PRE-1** respected — these are non-tx reads, so no
    authz-recording read sits inside a lock-holding atomic tx.
  - **Design system is consumed, not redesigned:** use `tokens.css` + existing primitives; changing a
    shared primitive trips HS-2.
  - Reads stay live (no caching workaround); **no merge / no push by the agent** (commits allowed
    after verified work).
- **Risks:**
  - **R1 — denominator query correctness.** The obligated set is a **union** (direct user-grants ∪
    active members of granted/process areas) over four tables + a view; getting the union +
    active-membership (`v_active_user_areas`) right is the core correctness risk. *Mitigation:* F2.2's
    integration test asserts the resolved set against a fixtured grant/area graph and **fails closed**
    on an unhandled grant type (never silently under/over-counts).
  - **R2 — "tracking pending" misread as a placeholder / trips the anti-"em breve" bar.** *Mitigation:*
    the state is explicitly labeled honest-disclosure (not illustrative, not a fake metric),
    operator-locked; the **old** illustrative watermark + `Em breve` CTA literals are removed so the
    program grep for the old markers holds at 0; the numerator gap is formally parked **with a
    trigger**. Mission §7/§8 wording reconciliation is a doc-only `wiki-curator` follow-up.
  - **R3 — scope creep into the numerator.** *Mitigation:* the rabbit-hole list + HS-6; any pull
    toward read/ack/CTAs/new tables **stops** and routes to the parked mission.
  - **R4 — authz cap design.** A new `CapDistributionRead` vs reusing a documents-read cap.
    *Mitigation:* F2.1's ADR settles it; default to a **dedicated read cap** to avoid over-grant.
  - **R5 — deadline / pending / overdue have no real producer.** *Mitigation:* **omit them** from M2;
    do not render a deadline or pending/overdue metric (it would be fabrication — violates quality
    goal 1). If the design wants the document's real effective/vigência date, F2.1 decides whether to
    show it **clearly labeled as a document date**, never relabeled as a "read deadline".
- **Rabbit holes (do NOT chase):**
  - *Building read/ack/target/reminder tables or any numerator producer* — parked mission; no consumer
    in M2 forces it once the screen shows the obligated set + honest pending state.
  - *Snapshot-at-publish / touching `PublishApproved()`* — HS-2; the publish path is Grade-A and must
    stay untouched; the snapshot-vs-derive decision is the parked mission's kickoff ADR.
  - *The 4 hero CTAs, row/bulk actions, the reminder worker, export* — action layer, parked mission.
  - *The timeline / adoption-curve endpoint and a chart library for it* — numerator-only; nothing real
    to chart in M2.
  - *Inventing a read deadline, pending, or overdue metric* — no producer exists → fabrication.

## Applicable hard-stops

| ID | What would trip it here |
|----|--------------------------|
| HS-1 | Milestone close: operator review gate before M3 and before any merge — mandatory. **Also gates the start:** this spec needs operator approval before F2.1 (M2 is full-stack/riskier than M1). |
| HS-2 | A "fix" turns out to require mutating the publish path (snapshot-at-publish), adding a table, or changing a shared primitive/token. **Stop**; report the boundary; route to the parked mission — do not symptom-patch and do not smuggle a publish-path side-effect. |
| HS-3 | A prerequisite fails at runtime: app won't start, no auth session, the document-detail route is broken, or the new endpoint's contract↔generated types drift. Repair the prerequisite (contract-first regen order), rerun the checkpoint, then resume the feature. |
| HS-4 | `milestone-validator` returns FAIL → open the named fix feature, re-run its lifecycle, re-dispatch the validator. |
| HS-6 | Scope drift: the numerator, a new table/migration, a third backend concern, or any action-layer surface appears in M2 → stop and replan; route to the parked mission. |
