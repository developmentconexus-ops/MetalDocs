# Milestone 2 — Distribuição coverage-scope (full-stack, derive-on-read)

> **Program:** frontend-screen-completion  ·  **Governing spec:** `../mission.md`
> **Status:** Re-decomposition approved (operator HS-1 + HS-6 path A, 2026-06-21) — **execution restarts in a fresh `/milestone` session.**
> *Subagent model: Sonnet 4.6 (operator directive 2026-06-21).*
> **Authored:** 2026-06-21 — *before any feature in this milestone began.*  ·  **Re-decomposed:** 2026-06-21 after evidence-based ADR-0039 boundary analysis (recon report in commit body).

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
  endpoint serving the obligated set (total + recipients + by-area totals) from existing tables.
  **No publish-path change** (`PublishApproved()` is not touched — snapshot-at-publish is explicitly
  refused here, see rabbit holes / HS-2).
- **The full numerator + action domain** (read/ack tables, reader-side acknowledge surface,
  snapshot-vs-derive decision, reminders/export/fanout worker) is **parked** as a designed mission,
  [`document-distribution-mission.md`](../../../../wiki/backlog/document-distribution-mission.md), to
  execute **after** the frontend-screen-completion mission finishes.

### Re-decomposition (2026-06-21) — published-view prerequisite

The original F2.1 spec assumed the distribution module could read the obligated set directly from
`metaldocs.v_cd_grantee` (migration `0243`). Mid-feature recon surfaced a Grade-A blocker: that view
carries only `(tenant_id, controlled_document_id, grantee_user_id)`, is gated
`visibility_scope='restricted'` **by design** (company-scope CDs contribute zero rows; search handles
those via `v_cd_search_facts.is_company` separately), and ADR-0039 forbids the new `distribution`
module from reading CD/taxonomy base tables raw. The contract M2 promised (per-recipient
`area_code`/`source`, by-area `coverage`, company-scope obligated set) is **not serveable from any
existing published view**.

Operator chose **Option A** (HS-6 path, 2026-06-21) after an evidence-based subagent recon:
**publish new sibling views owned by the rightful owners; do not extend `v_cd_grantee` in place** (no
DROP/ALTER VIEW precedent across 245 migrations; non-additive view DDL is policy-hostile per
`wiki/database/migration-policy.md`; mutating `v_cd_grantee` forces search to carry distribution-domain
knowledge → module-boundary leak). M2 therefore now contains **five features** instead of three:

- **F2.1a** (CD module) publishes `metaldocs.v_cd_obligated_readers` carrying the obligation shape
  distribution needs (`tenant_id`, `cd_id`, `user_id`, `area_code|null`, `source`, includes company-scope
  leg via `v_cd_search_facts.is_company` × active tenant users). Search untouched.
- **F2.1b** (taxonomy module) publishes `metaldocs.v_process_area_name` `(tenant_id, area_code,
  area_name)` so distribution can resolve area names without reading taxonomy's base tables.
- **F2.1c** (distribution module, new) authors the OpenAPI + ADR + codegen for the three
  denominator-only endpoints, reading the two new views + the ADR-0029 iam display-name read-port.
- **F2.2** is now the distribution handler implementation over those views.
- **F2.3** is unchanged: wire the FE.

This re-decomposition is the approved F2.x consumer contract; each feature's `spec.md` distills it.
Reconciling the mission §7/§8 wording ("fanout", "zero em breve" for the whole screen) is a **doc-only
`wiki-curator` follow-up** (non-blocking) — the screen reaches full mission-§8 completeness in two
stages: denominator now, numerator when the parked mission runs.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F2.1a | `f2.1a-cd-obligated-view` | New forward-only migration owned by the `controlleddocuments` module: `CREATE VIEW metaldocs.v_cd_obligated_readers AS …` carrying `(tenant_id, controlled_document_id, user_id, area_code TEXT NULL, source TEXT)` with three UNION legs — direct user-grant (`source='user_grant'`, `area_code=NULL`), area-grant ⋈ `v_active_user_areas` (`source='area_grant'`, `area_code=upa.area_code`), and company-scope leg (`source='company_scope'`, all rows of `v_cd_search_facts` where `is_company=true` × active tenant users via `v_active_user_areas` DISTINCT by `user_id`). ADR-0040 records the new view + adds an inventory row to ADR-0039. `v_cd_grantee` is **not** modified. | Migration applies cleanly + idempotent re-run = no-op; SQL spot-check against fixtured graph: a CD with 1 user-grant + 1 area-grant + a company-scope sibling returns the expected distinct rows with correct `source`/`area_code` per row; ADR-0040 present under `wiki/decisions/`; ADR-0039 inventory updated; `git diff db/migrations/0243*` = empty (search untouched); `git diff internal/modules/search` = empty; all 6 CI guards green. |
| F2.1b | `f2.1b-taxonomy-area-name-view` | New forward-only migration owned by the `taxonomy` module: `CREATE VIEW metaldocs.v_process_area_name AS SELECT tenant_id, area_code, name AS area_name FROM <taxonomy's process-area base table>`. ADR-0041 records the new view + adds an inventory row to ADR-0039. | Migration applies cleanly + idempotent; SQL spot-check returns one row per `(tenant_id, area_code)`; ADR-0041 present; ADR-0039 inventory updated; `git diff internal/modules/taxonomy` consists only of the migration + any required generated artifacts; all 6 CI guards green. |
| F2.1c | `f2.1c-distribution-contract` | ADR + OpenAPI for the three denominator-only endpoints, consumer-contract-first from `DocumentDistributionPage`: `GET /documents/:id/distribution` → `{total_targets}`; `GET /documents/:id/distribution/recipients?cursor=&limit=` → `{items: DistributionRecipient[], page: CursorPage}` (`DistributionRecipient {user_id, name, area_code\|null, area_name\|null, source: "area_grant"\|"user_grant"\|"company_scope"}` — `role` omitted, no honest producer); `GET /documents/:id/distribution/coverage` → `DistributionAreaCoverage[]` (`{area_code, area_name, total}`). New tier-2 cap `CapDistributionRead` (tenant-scope) registered + added to `deferredCaps` (operator grants to roles separately). ADR-0042 records the new `distribution` module + cap + denominator-only contract + additive-extension commitment to the parked mission + recipient distinct/`source` precedence rule (`user_grant` > `area_grant` > `company_scope`) + `role` omission. Regen Go server types (`oapi-codegen`) + FE types (`npm run gen:api`). | `api-lint -strict` parses the new `distribution` paths = **0** violations; response schemas contain **no** numerator field (`grep -nE "read\|acknowledg\|overdue\|pending\|deadline\|timeline\|reminder"` over the new `Distribution*` schema blocks = 0); generated Go + FE types present; ADR-0042 + cap registration in `internal/modules/iam/domain/model.go` + scope in `capability_scope.go` + entry in `scripts/api-lint/registry_rules.go:37 deferredCaps`; `go build ./...` green. |
| F2.2 | `f2.2-coverage-backend` | Implement the three endpoints to the Grade-A bar as a **read-only projection** over `v_cd_obligated_readers` (F2.1a) ⋈ `v_process_area_name` (F2.1b) + ADR-0029 iam display-name read-port for `name`. New handlers in `internal/modules/distribution/delivery/` enforce `CapDistributionRead` via `authz.Require` (+ `trg_require_cap_asserted`). Cursor pagination per the existing `CursorPage` convention, keyset on `(area_name, name, user_id)`. **No new table, no migration here (F2.1a/b own the only new DDL), no publish-path change.** | Integration test against live PG asserts the obligated set against a fixtured grant/area/company-scope graph (union + active-membership + company-scope correct; distinct-by-user with `source` precedence; fail-closed on an unhandled grant type); `api-lint -strict` = **0**; all **6 CI guards green** (notably `hgcrossmodule` — distribution reads only `metaldocs.v_*` published views + the iam port); `go build ./...` / `go vet ./...` / `go test ./...` green; `git diff db/migrations` consists only of F2.1a's + F2.1b's migrations; `git diff …/approval/application/publish_service.go` = empty. |
| F2.3 | `f2.3-distribuicao-wire` | Wire `DocumentDistributionPage` to the real coverage endpoints via TanStack Query hooks; **remove** the illustrative scaffolding (`MOCK_DISTRIBUTION`, watermarks, aria-hidden) for the **denominator** surfaces (total, recipient list, by-area). Render the **numerator** surfaces (read/ack donut, timeline, per-recipient read/ack columns, overdue/pending) as an explicit, labeled **"tracking pending — leitura/ciência ainda não registrada"** state. Keep the 4 hero CTAs + row/bulk actions **`aria-disabled` deferred-with-trigger** linking the parked mission. | `grep -nE "Dados ilustrativos\|MOCK_DISTRIBUTION\|Em breve" …/DocumentDistributionPage.tsx` = **0**; recipient list + by-area totals render live values (consumer shape == generated types); the "tracking pending" state renders (truthful empty, not fabricated, not illustrative); CTAs disabled with a trigger note; a query-hook test passes against fixtured responses; **`frontend-screen-reviewer` + `frontend-code-reviewer` APPROVE on record** (D2); FE tests green. |

Order: **F2.1a → F2.1b → F2.1c → F2.2 → F2.3** — owner-published views first (CD then taxonomy),
then the consumer module's contract, then its producer, then the wire. The two new views are the
prerequisite for any of distribution's reads to be Grade-A compliant (`hgcrossmodule`). Each "what to
validate" is objectively checkable (a grep = 0, an `api-lint` count, a contracted shape, a passing
integration/hook test, a reviewer APPROVE-on-record), never "works".

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. What that gate
enforces for this milestone:

1. **Per-feature acceptance** — F2.1a/F2.1b/F2.1c/F2.2/F2.3 each meet their declared "what to
   validate", and each feature's **consumer contract** (`spec.md`) was honored: distribution reads
   only published views + the iam port (no base-table reads); FE consumer shape matches the real
   generated producer types (denominator-only; no read/ack fields anywhere in the contract).
2. **Workflow-class QA** — **FE:** `wiki/quality/screen-definition-of-done.md` (D2: both reviewer
   APPROVEs on record) + the runtime functional pass by reference to `wiki/quality/screen-qa-checklist.md`.
   **BE:** `wiki/quality/backend-api-qa-checklist.md` + the 6 CI guards + `api-lint -strict` = 0 +
   the integration test.
3. **Regression** — **M0 + M1** still pass their gates (single index route, 0 dead-stub routes, M1
   dashboard greps hold); the FE suite holds at the operator-accepted baseline; **the publish path is
   untouched** (`git diff internal/modules/documents/approval/application/publish_service.go` = empty);
   **`v_cd_grantee` is untouched** (`git diff db/migrations/0243*` = empty); search module untouched
   (`git diff internal/modules/search` = empty); `go build ./...` / `go test ./...` green.
4. **Quality-bar re-measure (root cause, not symptom)** —
   `grep -rEn "Dados ilustrativos|MOCK_DISTRIBUTION|Em breve" …/pages/DocumentDistributionPage.tsx`
   = **0**; the obligated set (recipients + by-area + total) renders **live** at runtime (illustrative
   scaffolding deleted at root, not flag-hidden); the numerator surfaces render the explicit
   **honest "tracking pending"** state — not a fabricated metric, not the old illustrative block; the
   new endpoint serves real denominator data (proven by the F2.2 integration test).
5. **No unplanned scope** — only F2.1a + F2.1b + F2.1c + F2.2 + F2.3 are implemented; **no** numerator
   producer, **no** publish-path change, **no** action layer (CTAs/reminders/export), **no**
   modification of `v_cd_grantee` or any search-owned code. Any pull toward those is routed to the
   parked mission, not built here.

## Dependencies & constraints

- **Depends on:** M0 (truth-reset) + M1 (dashboard) passed; backend Grade-A baseline `d477e9f0`; the
  existing grant/area tables + `metaldocs.v_active_user_areas` + `metaldocs.v_cd_search_facts` (read by
  F2.1a's company-scope leg); the already-honest `DocumentDistributionPage` scaffolding as the wire
  target.
- **Appetite:** medium-plus — one full-stack slice across **three modules** (controlleddocuments
  publishes a view, taxonomy publishes a view, distribution module is greenfield) + the screen wire.
  Larger than M1 (real backend), bounded well below the parked mission (still no new mutable table, no
  worker, no publish-path change, no numerator, no action layer). The two new published views are
  read-only DDL.
- **Quality goals (ranked):** 1) **truthfulness** (render only the obligated set a real producer
  derives; the numerator is honestly disclosed as pending, never faked) > 2) **module-boundary
  cleanliness** (each module publishes its own read contract; distribution never base-table-reads
  another module; ADR-0039 inventory grows by exactly two rows) > 3) **contract-correctness**
  (consumer shape == generated producer types; denominator-only contract is **forward-compatible** —
  the parked mission extends it additively, never breaks it) > 4) **simplicity** (smallest backend
  that serves real data: read-only projection over the new views, reusing existing query/client
  patterns).
- **Architectural constraints (validator can fail on these):**
  - **Backend additive-only:** the only new DDL is two `CREATE VIEW` migrations (F2.1a, F2.1b). **No**
    new mutable table, **no** `ALTER`/`DROP` of existing views (forward-only per
    `wiki/database/migration-policy.md`), **zero** change to `PublishApproved()` or any publish-path
    code. Snapshot-at-publish is **refused** here (HS-2 + parked-mission scope).
  - **No raw cross-module base-table reads:** distribution reads only `metaldocs.v_cd_obligated_readers`
    + `metaldocs.v_process_area_name` + the ADR-0029 iam display-name port (`hgcrossmodule` = 0).
  - **`v_cd_grantee` is sacred:** untouched. Search semantics depend on its restricted-only invariant
    (migration 0243 COMMENT).
  - **Contract-first regen order:** spec → OpenAPI → `oapi-codegen` → FE types; the FE consumes the
    **generated** types only, never hand-rolled shapes (HS-3).
  - **Numerator + action layer are out of scope** — parked to
    [`document-distribution-mission.md`](../../../../wiki/backlog/document-distribution-mission.md);
    rendered as honest "tracking pending" / deferred-with-trigger, **never fabricated**.
  - **Grade-A backend bar:** the new endpoint passes `api-lint -strict` = 0, all 6 CI guards, and an
    integration test against live PG; `go build`/`vet`/`test` green.
  - **authz:** new tenant-scope cap `CapDistributionRead` registered in
    `internal/modules/iam/domain/model.go` + scoped in `capability_scope.go` + added to
    `scripts/api-lint/registry_rules.go:37 deferredCaps` (operator grants to roles separately —
    never pre-granted by the agent). Handlers gate with `authz.Require` + the
    `trg_require_cap_asserted` tripwire. **H-PRE-1** respected — non-tx reads only, no authz-recording
    read inside a lock-holding atomic tx.
  - **Design system is consumed, not redesigned:** use `tokens.css` + existing primitives; changing a
    shared primitive trips HS-2.
  - Reads stay live (no caching workaround); **no merge / no push by the agent** (commits allowed
    after verified work).
- **Risks:**
  - **R1 — obligated-set query correctness across three legs.** Union of direct user-grants ∪ active
    members of granted areas ∪ company-scope users (over four tables + two views), distinct by user
    with `source` precedence. *Mitigation:* F2.1a's view encodes the rule in one place; F2.2's
    integration test asserts the resolved set against a fixtured graph + **fails closed** on an
    unhandled grant type.
  - **R2 — "tracking pending" misread as a placeholder.** *Mitigation:* explicit honest-disclosure
    label, operator-locked; the **old** illustrative literals are removed so the program grep holds
    at 0.
  - **R3 — scope creep into the numerator.** *Mitigation:* the rabbit-hole list + HS-6.
  - **R4 — F2.1a's company-scope leg performance** (cross-join of all company-scope CDs × active
    tenant users could be wide). *Mitigation:* view is read-only; if F2.2's integration test or
    runtime measurement shows >acceptable latency, materialize via per-CD lateral or add an index — but
    only if proven slow (premature optimization is out of scope).
  - **R5 — deadline / pending / overdue have no real producer.** *Mitigation:* **omit them** from M2
    entirely (would be fabrication — violates quality goal 1).
- **Rabbit holes (do NOT chase):**
  - *Modifying `v_cd_grantee` or any search-owned code* — Grade-A boundary leak, HS-2.
  - *Building read/ack/target/reminder tables or any numerator producer* — parked mission.
  - *Snapshot-at-publish / touching `PublishApproved()`* — HS-2; the snapshot-vs-derive decision is
    the parked mission's kickoff ADR.
  - *The 4 hero CTAs, row/bulk actions, the reminder worker, export* — action layer, parked mission.
  - *The timeline / adoption-curve endpoint and a chart library for it* — numerator-only.
  - *Inventing a read deadline, pending, or overdue metric* — no producer exists → fabrication.

## Applicable hard-stops

| ID | What would trip it here |
|----|--------------------------|
| HS-1 | Milestone close: operator review gate before M3 and before any merge — mandatory. **Also gates the start:** this spec needs operator approval before F2.1a (re-decomposition already operator-approved 2026-06-21). |
| HS-2 | A "fix" turns out to require mutating the publish path (snapshot-at-publish), modifying `v_cd_grantee`, changing search-owned code, or changing a shared primitive/token. **Stop**; report the boundary; route to the parked mission — do not symptom-patch and do not smuggle a publish-path or search side-effect. |
| HS-3 | A prerequisite fails at runtime: app won't start, no auth session, the document-detail route is broken, or the new endpoint's contract↔generated types drift. Repair the prerequisite (contract-first regen order), rerun the checkpoint, then resume the feature. |
| HS-4 | `milestone-validator` returns FAIL → open the named fix feature, re-run its lifecycle, re-dispatch the validator. |
| HS-6 | Scope drift: the numerator, a new mutable table, a third backend concern, or any action-layer surface appears in M2 → stop and replan; route to the parked mission. *(Already raised + resolved 2026-06-21 via the F2.1a/b/c split; further drift trips it again.)* |
