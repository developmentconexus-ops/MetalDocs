# Milestone 3 — Mechanical Quality (Wave C)

> **Program:** grade-a-architecture-remediation  ·  **Governing spec:** `docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md`
> **Status:** Passed 2026-06-15 — `milestone-validator` C1–C7 PASS (`qa/milestone-qa.md`); awaiting operator HS-1 to open M4
> **Authored:** 2026-06-14 — *before any feature in this milestone began.*

> **⚠ HS-6 scope reconciliation (2026-06-14).** Before executing F3.1, an investigation against the
> current tree found the governing spec's §5.2 dead-surface/tx-hazard list was **stale** — it was
> derived from the 2026-06-13 audit, but **Wave 2.11** (`63f74368`, 2026-06-12) and **Wave 2.12**
> (H-6a) had already removed the orphans and hoisted the deadlock-bearing read *before* the audit was
> written. Verified current-code findings:
> - **F3.1** — all 7 orphans already deleted (`CutoverService`, `CompositionConfig`,
>   `AreaService.SetParent`, `resolvePermissionFallback`, `WorkerConfig.ReviewReminderDays`,
>   `SnapshotFromTemplate`); `coverage_boost_test.go` survives but is **live** (exercises real
>   production symbols). Zero deletes remain. **Already done.**
> - **F3.2** — the H-PRE-1 deadlock root cause (authz-recording `GetByCode` via
>   `ensureTemplateArtifact`) is **already hoisted off-tx** (`service.go:278` comment + audit runtime
>   proof). The residual in-tx reads at `service.go:308,331` are **plain non-authz `SELECT`s**
>   (`GetTemplateVersionState`, `CodeExists`) — **no H-PRE-1 hazard**. The `GetTemplateVersionState`
>   port refactor is M4 F4.2 (reach-without-a-port, a different concern). **Already done.**
> - **F3.6** — no "dead camelCase `MarshalJSON`" exists. Both auth marshallers (`Config.MarshalJSON`,
>   `AuthenticatedSession.MarshalJSON`) are **live security-redaction guards** with passing redaction
>   tests, implicitly invoked by `json.Marshal`. Deleting one risks leaking secrets. **Struck** as a
>   stale, security-unsafe finding.
>
> **Net real remaining work:** F3.3, F3.4, F3.5. F3.1 and F3.2 close as verify-already-done evidence
> rows; F3.6 is struck. Separately surfaced: stale wiki/GitNexus refs to the deleted symbols (doc
> drift — handed to wiki-curator, not an M3 feature).

> This file is a **spec**, authored up front. It says **what** this milestone is,
> **which features** it contains, **what each feature implements**, and **what gets
> validated**. It contains **no execution steps** — the "how" of each feature lives in
> that feature's `plan.md`. The end-of-milestone QA (`qa/milestone-qa.md`) validates
> the milestone against *this* document.

## Objective

Harden the backend's two remaining formerly-C-adjacent audit dimensions — **code-quality** and
**persistence** — by clearing the bounded, single-purpose tail surfaced in the governing spec §5.2.
Every item is mechanical and locally scoped: dead-surface removal, a transaction-hazard hoist, a
raw-map/raw-error tightening, and a pagination TOCTOU fix. No shared-API redesign, no new ports
(that is M4). 

**Bar moved & criterion:** the **code-quality** and **persistence** dimensions reach the §6 pass bar
(≥ A−). Post-reconciliation, the two original root causes are confirmed **already closed** in prior
waves (verified, not asserted — see HS-6 note above): dead-surface (7 orphans gone, each
zero-caller-proven) and tx-hazard (H-PRE-1 deadlock read hoisted off-tx, runtime-proven). The real
remaining bar-movement is the **code-quality tail** — three bounded single-purpose fixes (F3.3–F3.5)
that remove a raw-map response, a pagination TOCTOU race, and three silently-discarded delete errors.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F3.1 | `f3.1-dead-surface-deletes` | **Verify-already-done** (Wave 2.11/2.12). No deletes to perform. | Zero-caller proof recorded for all 7 named symbols against the current tree (`grep` → 0 `.go` matches); build green; `coverage_boost_test.go` confirmed live (exercises only production symbols). Evidence row only. |
| F3.2 | `f3.2-tx-hazard-hoist` | **Verify-already-done** (prior wave). No hoist to perform. | Evidence that the H-PRE-1 deadlock read (`ensureTemplateArtifact`→`GetByCode`) is already off-tx (`service.go:278` comment + audit runtime proof, `pg_locks`=0); residual in-tx `GetTemplateVersionState`/`CodeExists` shown to be **non-authz** plain SELECTs (`repository.go:702,72`) → no hazard. Port refactor explicitly deferred to M4 F4.2. Evidence row only. |
| F3.3 | `f3.3-cd-raw-map-to-type` | Replace CD `routes.go:123` raw `map[string]any` 201-Created response with the generated response type. | Handler returns the generated/declared `AtomicCreateResponse`; the **only** wire delta is `null`→omitted on absent optionals (`department_code`, `override_template_version_id`, `sequence_num`) — a drift-*fix* onto the already-declared `,omitempty` contract the FE already types (no OpenAPI edit, no FE regen, HS-2 does not trip; *not* byte-identical, and that is the point); `go build` + lint clean; no `map[string]any` at that response site. backend-api-qa-checklist green. |
| F3.4 | `f3.4-documents-pagination-toctou` | Fix documents list pagination TOCTOU in `repository.ListDocumentsPaginated` — derive total from a single statement instead of a separate `repo.CountDocuments`. **HS-6 note:** the code is **keyset/cursor** pagination, so a bare `COUNT(*) OVER()` on the cursor-filtered query would count only the post-cursor tail. Correct fix (operator-approved Approach **B**): a CTE computes `COUNT(*) OVER()` over the **base-filtered** set *before* the cursor predicate; the cursor + `LIMIT n+1` apply in the outer query → page rows and grand total share one MVCC snapshot. See `spec.md` HS-6 reconciliation. | Total count and page rows come from **one** query (no separate-COUNT race window). Grand total is page-independent (counted pre-cursor), keyset `hasMore`/cursor/`ErrInvalidCursor` preserved; existing pagination tests green; a test demonstrates total/rows consistency. |
| F3.5 | `f3.5-deleteobject-error-log` | Surface the silently-discarded `DeleteObject` error(s) as WARN logs. **Site reconciliation:** the named `:537/:740/:331-334` are stale — the current tree has **one** `.DeleteObject(` swallow in documents `service.go` (the `:534` orphan-cleanup on hash mismatch in `CommitAutosave`); file is 710 lines so `:740` cannot exist, `:331` is not a DeleteObject. The other two named sites do **not** apply (prior waves removed/relocated them). Documented per this row's own "or documented" clause — not an HS-6 stop. | The previously-swallowed error logged at WARN with key context (`storage_key`, `document_id`, `err`) via the in-module `slog` convention; no behavior change (still returns `ErrContentHashMismatch`, delete stays best-effort); happy paths unchanged; build + tests green. The single real site covered; the two stale-named sites documented as non-existent in the current tree. |
| ~~F3.6~~ | ~~`f3.6-auth-dead-marshaljson`~~ | **STRUCK** (HS-6). No dead `MarshalJSON` exists; both auth marshallers are live security-redaction guards. | n/a — see HS-6 note. |

For each live feature, "what to validate" is **objectively checkable** — a clean build, a byte-identical
response diff, a pagination total that matches rows, a WARN line emitted on the error path. No "works"
/ "looks right". F3.1 and F3.2 are verify-already-done: their acceptance is recorded **evidence of
prior closure**, not new code.

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. What that gate
enforces for this milestone:

1. **Per-feature acceptance** — every feature above meets its declared "what to validate", and each
   feature's `spec.md` consumer contract (where a contract exists — F3.3's response shape, F3.4's
   pagination envelope) was honored (producer matches consumer; shapes unchanged).
2. **Workflow-class QA checklists** — `backend-api-qa-checklist` (all features) **+**
   `workflow-async-qa-checklist` (F3.2 touches CD-create, an async/lock-bearing path).
3. **Regression** — M0, M1, M2 still pass their gates; the M1 test-infra-rebaseline full-HTTP
   `seed→finalize→signoff` E2E stays green (F3.2 touches the same CD/approval area).
4. **Quality-bar / root-cause check** — **focused audit slice** on **code-quality** + **persistence**:
   - each of the 7 deletes proven caller-free (root cause = dead surface gone, not hidden);
   - tx-hazard gone with **H-PRE-1 intact** (root cause = reads hoisted off-tx, not a Tx-variant
     slipped inside the lock);
   both dimensions re-measured ≥ A−.
5. **No unplanned scope** — anything implemented beyond F3.1–F3.6 is recorded with rationale.

## Dependencies & constraints

- Depends on: M0, M1, M2 passed (this milestone runs on the contract-clean baseline they established).
- Architectural constraints respected:
  - **H-PRE-1** advisory-lock deadlock rule (memory `[[advisory-lock-deadlock-constraint]]`): never
    call an authz-recording read on a fresh connection inside the audit-lock atomic tx; F3.2 fixes by
    **hoisting off-tx**, never by introducing a Tx-variant read inside the lock.
  - Contract-stability: F3.3 and F3.4 must not change response shapes (no FE regen expected this
    milestone — if any contract shape would change, that is a hard-stop, see below).
  - Surgical-changes rule (CLAUDE.md §5.3): each feature touches only its named site(s); no adjacent
    refactor, no opportunistic cleanup.

## Applicable hard-stops

Default catalog HS-1..HS-6 in force. What would trip each here:

- **HS-1** — milestone close gate. On validator PASS, the **main session** flips status and presents
  to the operator; **no M4 open and no merge without explicit approval**.
- **HS-2** — if any feature's fix turns out to require redesign outside its boundary (e.g. F3.2's
  hoist can't be done without changing the lock/tx model, or F3.3's type change forces a contract
  shape change rippling to FE): **stop**, report the boundary + minimum prerequisite plan, do **not**
  symptom-patch. (Esp. guard F3.2 against an H-PRE-1-violating "easy" fix.)
- **HS-3** — if a prerequisite boundary fails (build / runnable / auth-session / CD-create route /
  contract truth) while executing a feature: repair via `runtime-contract-prereq`, rerun the failed
  checkpoint, then resume.
- **HS-4** — validator returns FAIL (a delete not proven caller-free, tx-hazard symptom-patched /
  H-PRE-1 violated, a contract shape silently changed): open the named fix feature, re-run its
  lifecycle, re-dispatch the validator.
- **HS-5** — N/A mid-program (terminal acceptance is M5).
- **HS-6** — scope drift: if execution surfaces dead surface or hazards beyond the named 7 + 2,
  **stop**, surface it, replan before absorbing it (it may belong to M4 or a new micro-task).
