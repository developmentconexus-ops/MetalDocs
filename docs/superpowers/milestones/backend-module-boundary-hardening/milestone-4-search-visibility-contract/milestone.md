# Milestone 4 — search consumes published visibility contracts (C4, risk-isolated, last)

> **Program:** backend-module-boundary-hardening  ·  **Governing spec:** `../mission.md`
> **Status:** Passed — milestone-validator **PASS** 2026-06-21 (C1–C7, gates re-run from clean state; pre-existing failures independently confirmed via pre-M4 worktree). All 3 features (F4.1/F4.2/F4.3, within the ~4-feature appetite cap) closed. **Awaiting HS-1 operator acceptance** (not merged, not pushed). Verdict: `qa/milestone-qa.md`.
> **Authored:** 2026-06-21 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is,
> **which features** it contains, **what each feature implements**, and **what gets
> validated**. It contains **no execution steps** — the "how" of each feature lives in
> that feature's `plan.md`. The end-of-milestone QA (`qa/milestone-qa.md`) validates
> the milestone against *this* document.

## Objective

After this milestone, **`internal/modules/search`'s v2 documents reader reads no other
module's base table.** Today (`internal/modules/search/infrastructure/v2documents/reader.go`)
its one unified list query inlines five foreign base-table reads — `public.documents`
(documents), `public.controlled_documents` + `controlled_document_area_grants` +
`controlled_document_user_grants` (controlleddocuments), `public.user_process_areas`
(iam) — and even hardcodes CD's `visibility_scope` literals (`'company'`, `'restricted'`
at `:92,:94`). It is the H-G worst offender: it re-implements CD's entire
company/owner/restricted+area-grant+user-grant visibility predicate from CD's and iam's
raw tables.

This milestone makes search **consume published contracts** instead: the producers
(documents, controlleddocuments) publish purpose-built, versioned read views — and
search JOINs those plus iam's already-published `metaldocs.v_active_user_areas` (M3/F3.1).

**Quality bar moved:** the cilint H-G debt ledger `hgPendingRemediation`
(`tools/cilint/internal/analyzers/hgcrossmodule.go`) drops its **final 5 rows** (C4a–C4e)
to **EMPTY** — `go run ./tools/cilint ./...` stays exit 0, and the mission's terminal
acceptance precondition (`mission.md` §8: ledger empty) is met. This is the last debt
milestone of the mission. **Observable proof:** per-site integration parity test (raw
result == published-contract result, GREEN on real PG) before each raw read is deleted,
plus a clean `go run ./tools/cilint ./...` and a green `go test ./tools/cilint/...` after
the ledger and its negative-baseline test are realigned to the empty end-state.

**Non-negotiable:** this is a **seam** change only — **zero** visibility/authz semantics
change. Search visibility is exactly the authz-drift risk class, so the parity tests seed
a **revoked-membership** and an **ungranted-user** discriminator (M3/F3.2 precedent:
`internal/modules/controlleddocuments/infrastructure/membership_view_parity_integration_test.go`).

## Appetite

- **Appetite:** ~4 features cap; ≤2 new migrations (the documents projection view + the
  CD visibility view; iam's view already exists from M3). No new Go domain types unless a
  feature interview proves search needs one.
- **Rabbit holes (do not chase):**
  - **Re-porting documents↔CD coupling inside the producer views.** A producer view exposes
    only the columns its OWN module owns. documents' projection view must **not** JOIN
    `controlled_documents` (that would just relocate the H-G violation into documents). The
    document↔CD correlation stays in search's JOIN across the two published views. Out of
    scope: "fix" it by having one module own both.
  - **Materializing the full (cd × actor) visibility cross-product.** A naive
    `v_cd_actor_visibility(cd, actor)` view explodes on company-scope CDs (every tenant user).
    The contract shape is the F4.1 consumer-contract interview decision — do not pre-build a
    cross-product view.
  - **Touching the v1/legacy search path or non-visibility search filters** (text, status,
    profile, family, date). Only the foreign base-table reads and the visibility predicate
    move. Family resolution already went through the taxonomy port (ADR 0038) — untouched.
  - **Widening to other search readers or other modules.** Scope is exactly `reader.go`'s
    `ListDocuments` and the producer views it consumes.

## Features

> Hypothesis decomposition (one feature per producer-contract, then one search-consume
> feature). Each producer feature's `spec.md` derives its view contract **from the search
> consumer first** (consumer-contract-first); the exact column/shape is finalized in that
> feature's interview, NOT prescribed here. If any feature's interview shows the contract
> needs a cross-module API redesign beyond a contained view-consume → **HS-2 stop** and
> surface before building.

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F4.1 | `f4.1-cd-visibility-contract` | controlleddocuments publishes a versioned visibility read contract (view) expressing CD visibility facts the search consumer needs — company/restricted scope, owner, area-grant, user-grant legs — sourced from CD's own grant tables joined to iam's `v_active_user_areas` (compliant D3a), per ADR-0039 D3a. Covers C4b (`controlled_documents`), C4c (`controlled_document_area_grants`), C4e (`controlled_document_user_grants`) and removes the need for search to inline CD's `visibility_scope` literals. | Migration applies on real PG; a parity query proves the view's visibility decision == the current inline-predicate decision across all scopes incl. **revoked-membership** + **ungranted-user** discriminators; ADR-0039 exemption references the new view; consumer-contract approval line present in `spec.md` before code. |
| F4.2 | `f4.2-documents-search-projection` | documents publishes a versioned search-projection read contract (view) exposing exactly the `public.documents` columns search's list query projects/filters on (no CD columns — see rabbit hole). Covers C4a (`documents`). | Migration applies on real PG; parity query proves the projection rows == current `public.documents` selected/filtered rows for the search query's column set; ADR-0039 exemption references the new view; consumer-contract approval line present before code. **Census note:** C4a is the F0.2-census-recorded `reader.go:69` read (`census.md` C4a) — recorded M4 scope, not new. |
| F4.3 | `f4.3-search-consume` | search `reader.go` `ListDocuments` JOINs the three published views (CD visibility F4.1 + documents projection F4.2 + iam `v_active_user_areas`) instead of the five foreign base tables; removes the hardcoded `'company'`/`'restricted'` literals (:92,:94); set-based SQL + LIMIT/OFFSET pagination preserved. Drains all 5 C4 rows from `hgPendingRemediation`; **realigns `TestHGCrossModule_Negative_PendingBaseline` to the EMPTY-ledger end-state** (the ledger has no live entry left to point at — convert the negative test to a synthetic in-test ledger entry or to an empty-ledger invariant; do NOT leave it pointing at a drained row, do NOT ship a silent FAIL). | Per-site integration parity test (search results identical pre/post across all visibility scopes incl. revoked-membership + ungranted-user) GREEN on real PG **before** each raw read is deleted; `go build ./...` + `go test ./...` green; `go run ./tools/cilint ./...` exit 0 with `hgPendingRemediation` empty; `go test ./tools/cilint/...` green after the negative-baseline realignment. |

For each feature, "what to validate" is objectively checkable: a migration that applies, a
parity test that passes on real PG, a clean cilint exit, a green ledger test.

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. For this milestone:

1. **Per-feature acceptance** — each feature above meets its declared "what to validate", and each
   feature's **consumer contract** (`spec.md`) was honored (producer view matches the search
   consumer's required shape; consumer-contract-first approval line present before code).
2. **Workflow-class QA checklist** — `wiki/quality/qa-operating-system.md` +
   persistence/migration discipline (`wiki/database/index.md`) + module-boundaries (ADR-0039).
   Integration parity needs the test PG on `:5434`; if down, the parity steps are **not-run
   (HS-3)**, never false-green.
3. **Regression** — M0/M1/M2/M3 gates still pass; the M3 CD/approval membership-view consumers
   and their parity tests are unaffected; `go build ./...` + `go test ./...` green.
4. **Quality-bar / root-cause check** — `hgPendingRemediation` is **EMPTY** and
   `go run ./tools/cilint ./...` is exit 0 (root cause = search inlining foreign predicates,
   fixed by consuming published views, NOT symptom-patched with `//cilint:allow` directives or by
   silently dropping ledger rows without deleting the raw read). The negative-baseline cilint test
   asserts the correct empty-ledger end-state and is green.
5. **No unplanned scope** — anything beyond the four features / the rabbit-hole list is recorded
   with rationale. The five C4 sites trace to `mission.md` §5 row 15 + the F0.2 `census.md` C4a–C4e.

## Dependencies & constraints

- **Depends on:** M0 (ADR-0039 + cilint H-G guard + ledger), M3/F3.1 (iam's published
  `metaldocs.v_active_user_areas`, migration 0242 — search's C4d leg repoints to it). M1/M2/M3
  all `passed` + HS-1 approved.
- **Quality goals (ranked):** **(1) authz/visibility correctness** (zero drift — parity + revoked
  + ungranted discriminators) > **(2) boundary integrity** (no foreign base-table read; ledger
  empty; root-cause not symptom-patch) > **(3) set-based performance** (no N+1; pagination on the
  filtered set preserved).
- **Architectural constraints (hard rules):**
  - Published views are `SELECT`-only, versioned, owner-published (ADR-0039 D3a). A producer view
    exposes only its OWN module's tables (+ compliant reads of other published views, e.g. CD's
    visibility view may read iam's `v_active_user_areas`). No producer view reads a third module's
    **base** table.
  - The active-membership leg encodes exactly `effective_to IS NULL` (ADR 0037 D1) via
    `v_active_user_areas` — no interval reinterpretation.
  - **Parity-before-delete (D6):** no raw read removed until its per-site integration parity test
    is GREEN on real PG. Seed revoked-membership + ungranted-user discriminators.
  - **Ledger discipline:** draining a `hgPendingRemediation` row REQUIRES the raw read actually
    gone; never drop a row while the read survives. After any ledger change run
    `go test ./tools/cilint/...`.
  - PowerShell for local startup (never bash / `source .env`); never read/print `.env`. Go via Bash
    with `GOCACHE="$PWD/.gocache"` + `METALDOCS_DATABASE_URL` set to the `:5434` DSN.
  - Commits local after verified work (standing authorization). **Never push. Never merge.**
- **Risks (named):**
  - **R1 — HS-2 visibility redesign.** The CD visibility contract may resist a contained
    view-consume (e.g. company-scope cross-product, or the predicate genuinely needing a shared-API
    redesign). *Mitigation:* F4.1 consumer-contract interview surfaces it FIRST; if it exceeds a
    contained view-consume → HS-2 stop, surface, replan. Do not symptom-patch.
  - **R2 — authz drift.** A view that subtly differs from the inline predicate silently changes who
    can see a document. *Mitigation:* per-site parity test with revoked-membership + ungranted-user
    discriminators, GREEN before delete (R1 quality goal #1).
  - **R3 — empty-ledger negative-test end-state.** Draining the final C4 rows leaves
    `TestHGCrossModule_Negative_PendingBaseline` with no live entry. *Mitigation:* F4.3 explicitly
    converts it to a synthetic in-test entry or an empty-ledger invariant; never a silent FAIL.
  - **R4 — test PG `:5434` unavailable.** *Mitigation:* HS-3 — mark parity steps not-run, never
    false-green; resume when PG is live.

## Applicable hard-stops

- **HS-1** — milestone boundary: after validator PASS, main session flips status, then operator
  review gate. No merge/push, no mission terminal acceptance, without explicit operator approval.
  M4's HS-1 is also the gate before the mission's terminal acceptance (§8).
- **HS-2** — a fix implies redesign outside the assigned boundary (esp. the F4.1 CD visibility
  contract needing a shared-API redesign beyond a contained view-consume). Stop; report the
  boundary + minimum prerequisite plan; no symptom-patch.
- **HS-3** — a prerequisite boundary fails (build/runnable; test PG `:5434` unavailable for an
  integration parity test). Repair / note the gap; rerun the checkpoint; resume. Never false-green a
  skipped parity test.
- **HS-4** — `milestone-validator` returns FAIL. Open the named fix feature; re-run its lifecycle;
  re-dispatch the validator.
- **HS-6** — scope drift (a producer view balloons, or a feature interview surfaces an in-scope site
  not in this spec). Stop; surface; re-interview / replan before continuing.
- **HS-PRE-1** — a read would place an authz-recording read inside a lock-holding atomic tx. N/A by
  construction here (search reads are off-tx, SELECT-only views), but recorded: no feature may
  introduce one.
