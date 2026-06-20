# Milestone 0 — ADR-0039: lock the definition + binding census + CI guard

> **Program:** backend-module-boundary-hardening  ·  **Governing spec:** `../mission.md`
> **Status:** All features closed (F0.1/F0.2/F0.3) · milestone-validator **PASS** (`qa/milestone-qa.md`) · HS-1 operator gate pending
> **Authored:** 2026-06-20 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is,
> **which features** it contains, **what each feature implements**, and **what gets
> validated**. It contains **no execution steps** — the "how" of each feature lives in
> that feature's `plan.md`. The end-of-milestone QA (`qa/milestone-qa.md`) validates
> the milestone against *this* document.

## Objective

**Make the done-bar mechanical before any code moves.** Today "cross-module raw SQL read" is a
judgement call — the parent re-audit counted "14", the discovery census found "~20" and corrected two
ownership facts. A program that remediates against a fuzzy, undercounted bar repeats the five-consecutive
Contract/API miss that caused this mission to exist. M0 closes that failure mode at the root:

After this milestone, a reviewer (or a CI job) can classify **any** cross-module read in `internal/modules`
as **compliant** or **violation** *mechanically*, with no judgement, because:

1. **F0.1** — ADR-0039 states the rule in one sentence (raw read of another module's **base table** =
   violation; reading its **published view / owner read-port** = compliant; reading one's **own** tables =
   compliant) plus a named exemption list (D4).
2. **F0.2** — the binding census re-runs the discovery sweep *against that ADR definition*, widened beyond
   the named H-G tokens, producing the **authoritative** in-scope site list with **0 sites unclassified** and
   a written coverage statement. This is the work inventory M1–M4 execute against — not the re-audit's "14".
3. **F0.3** — a cilint H-G analyzer (sibling of the H-D `noresponsemap` guard) **fails the build** on any
   raw cross-module base-table read outside the published-contract allowlist. The class cannot silently
   re-open.

**Quality bar moved:** the H-G definition goes from *implicit / disputed* to *ADR-locked + CI-enforced*.
Proof it moved: ADR-0039 exists with status Accepted; `go run ./tools/cilint ./...` exits non-zero today
(the guard is **red** on the current tree, flagging the in-scope sites) — a deterministic, re-measurable
signal, not an adjective. M0 deliberately makes the build red; M1–M4 turn it green site by site.

**Consumer of this milestone:** the *next four milestones* (M1–M4) and the *terminal mission-validator*.
They consume F0.2's site list as their work inventory and F0.3's guard as their per-site acceptance signal
(a site is "ported" when the guard stops flagging it). M0 produces exactly what they assume exists.

## Appetite + rabbit holes

- **Appetite:** definition + census + guard only. No production SQL is ported in M0 (that is M1–M4). One ADR,
  one census document, one analyzer + its tests, the runner wired. ~3 features, no schema migrations, no
  behavior change.
- **Rabbit holes (do not chase):**
  - **Porting any site now.** M0 *defines and measures*; it must not *fix*. Touching `resolution.go` or any
    repository SQL is M1+ scope — porting before the bar is locked is the exact inversion this milestone
    exists to prevent.
  - **Publishing the views/ports themselves.** The membership view (M3) and read-ports (M2) are named in the
    ADR's exemption list as the *target mechanism*, but creating them is M2/M3 work. M0 only declares them
    compliant-by-design; it does not build them.
  - **A perfect general-purpose SQL parser.** F0.3 is a token/string-literal guard (like its H-D sibling),
    not a full SQL AST engine. It must flag every current in-scope site and hold the exemptions — chasing a
    bullet-proof parser for hypothetical aliased/dynamic SQL is out of appetite; that residual is recorded as
    a census assumption, not engineered away in M0.
  - **Re-litigating ADR 0037's temporal model.** The active-now predicate is `effective_to IS NULL`, settled
    in ADR 0037. ADR-0039 references it; it does not reinterpret it.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F0.1 | `f0.1-adr-0039` | ADR `wiki/decisions/0039-*.md`: the H-G definition (raw read of another module's **base table** = violation), the exemption list (D4: published versioned view JOIN / owner-published read-port / own tables = compliant), the active-now membership-view contract (`effective_to IS NULL`, ADR 0037, no reinterpretation), the H-PRE-1 note (no authz-recording read inside a lock-holding tx), and supersession/links to ADR 0022/0037/0038. Status **Accepted**. | ADR file exists; status Accepted; the definition + exemption list are **unambiguous** — a reviewer can mechanically classify each of the §5 mission sites as compliant/violation with no judgement (demonstrated by a worked classification table in the ADR or F0.1 evidence); operator-ratified at the M0 HS-1 gate. |
| F0.2 | `f0.2-binding-census` | Re-run the cross-module owned-base-table read census against ADR-0039's definition, **widened beyond** the named H-G tokens (`user_process_areas`, `controlled_document*`, `document_process_areas`, `documents`, `approval_instances`, `document_profiles`) to the full owned-table set across `internal/modules/**` non-test. Produce the authoritative in-scope site list + per-site owner + class (A/B/C) + milestone home + a written coverage statement (what was swept, what is assumed). | Census reproduces the ~20 discovery-brief sites (none dropped); every site is classified compliant/violation against ADR-0039 with **0 sites unclassified**; any **new** in-scope site is added to its milestone (HS-6 if it changes a milestone's shape) or recorded out-of-scope **with reason**; coverage statement names the residual assumptions (dynamic/aliased SQL). |
| F0.3 | `f0.3-cilint-guard` | A cilint analyzer (`hgcrossmodule`, sibling to `noresponsemap`) that flags a raw cross-module **base-table** read (`FROM`/`JOIN`/`EXISTS` against another module's owned base table) in a non-owner package, outside the published-contract allowlist (views + recorded exemptions + inline `//cilint:allow-*` directive). Wired into `analyzers.RunAll`; runs via `go run ./tools/cilint ./...`. | Analyzer **flags every current in-scope §5 site** (guard is **red** on today's tree — `go run ./tools/cilint ./...` exits non-zero listing them); the allowlist holds the ADR-0039 exemptions (published views, own-table reads, owner ports) with **zero false positives** on compliant code; unit tests prove the guard **bites** on a planted violation and is **green** on a planted exemption; `go build ./...` + `go test ./tools/cilint/...` green. |

For each feature, "what to validate" is **objectively checkable** — a file with a status field, a census with
a 0-unclassified count, a linter that exits non-zero on named lines. No "works" / "looks right".

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. What that gate
enforces for this milestone:

1. **Per-feature acceptance** — every feature above meets its declared "what to validate". F0.1's ADR is
   Accepted and mechanically classifies the §5 sites; F0.2's census is 0-unclassified and reproduces the ~20
   sites; F0.3's guard is red on today's tree (flags every in-scope site) with zero false positives, and its
   unit tests prove bite+green. Each feature's **consumer contract** (`spec.md`) was honored — the census
   consumes the ADR's definition; the guard consumes the census's token/owner set + the ADR's allowlist.
2. **Workflow-class QA checklist** — contract/architecture-truth (this is a definition + gate milestone, not
   a runtime-behavior one): `wiki/quality/qa-operating-system.md` close-out discipline + `wiki/standards/`
   docs governance for the ADR. Code-quality checklist for the F0.3 analyzer (`go build`, `go vet`,
   `go test ./tools/cilint/...`).
3. **Regression** — no prior milestones in *this* program (M0 is first). Cross-program regression: the
   existing cilint analyzers (`noresponsemap`, `nosqltxindomain`, `txownership`, …) still pass; adding the
   H-G analyzer must not change their findings. `go build ./...` + `go test ./...` not regressed (modulo the
   documented pre-existing failures, labeled not-green-by-this-change).
4. **Quality-bar / root-cause check** — the bar is "H-G definition locked + CI-enforced". Re-measure: ADR-0039
   status == Accepted; `go run ./tools/cilint ./...` is **deterministically red** on the in-scope sites and
   **green** on the exemptions. Root cause of the undercount (no machine-checkable definition) is **fixed**
   (the analyzer is the machine check), not symptom-patched (the count is not hand-asserted).
5. **No unplanned scope** — **no production SQL ported**, **no view/port created**, no schema migration. Any
   site the census surfaces beyond the ~20 is *recorded*, not *fixed*, in M0. Anything beyond F0.1–F0.3 is
   recorded with rationale (HS-6).

## Dependencies & constraints

- **Depends on:** the discovery brief (`../discovery-brief.md`) as the census starting point; ADR 0022
  (authz root cause), 0037 (membership temporal model — `effective_to IS NULL`), 0038 (owner-published port
  precedent) as the ADR-0039 antecedents. The existing cilint harness (`tools/cilint/`) as the F0.3 host.
- **Quality goals (ranked):** **1. correctness of the definition** (an ambiguous ADR poisons all four
  downstream milestones) > **2. exhaustiveness of the census** (a missed site is a missed remediation and a
  third undercount) > **3. zero false positives in the guard** (a noisy guard gets disabled, defeating its
  purpose). Performance is a non-goal — these are build-time tools.
- **Architectural constraints respected:**
  - **No behavior change, no migration, no production code edit** — M0 is definition + measurement + guard
    only. The only source files written are the ADR (wiki), the census (docs), and the analyzer + tests
    (`tools/cilint/`).
  - **ADR 0037 temporal model is referenced, not reinterpreted** — the published-view contract encodes
    exactly `effective_to IS NULL`.
  - **H-PRE-1** — the ADR records the rule (no authz-recording read inside a lock-holding tx); M0 writes no
    code that could trip it.
  - **Docs governance** — ADR follows `wiki/standards/documentation-governance.md`; new ADR number = 0039
    (next free); links/supersession stamped.
  - House rules: never read/print `.env`; PowerShell for any local startup; commits local only, **never
    push**; docker Postgres (:5433) may be down — any integration step skipped is noted, no false green
    (M0 has no integration step — it is static analysis + docs).
- **Risks (named):**
  - *Census surfaces materially new in-scope sites* (the brief deferred the full owned-table widen to M0) →
    **HS-6**: if it changes a milestone's shape, stop and surface to the operator before continuing.
  - *F0.3 false-positives on a compliant own-module read or a view* → mitigated by owner-aware detection +
    the exemption allowlist + bite/green unit tests; a false positive is a C-check fail, not shippable.
  - *F0.3 misses an aliased/dynamic-SQL site* → accepted residual, **recorded** in the census coverage
    statement (the guard matches literal table tokens, like its H-D sibling); not engineered away in M0.

## Applicable hard-stops

| ID | What would trip it here |
|----|--------------------------|
| HS-1 | M0 boundary — operator review gate. After the milestone-validator PASS, the main session flips status and presents the gate; **no M1 and no merge without operator approval.** The ADR's operator ratification happens at this gate. |
| HS-2 | If locking the ADR definition reveals that an exemption (e.g. a "published view") actually requires a cross-module API redesign to be coherent → stop, surface the boundary. (Low risk in M0 — M0 only *declares* the mechanism; M2–M4 *build* it.) |
| HS-3 | A prerequisite fails (`go build ./...` red on base, or `tools/cilint` won't compile). Repair/note, rerun, resume. Docker :5433 is irrelevant to M0 (no integration step). |
| HS-4 | The milestone-validator returns FAIL → open the named fix feature, re-run its lifecycle, re-dispatch. |
| HS-6 | The F0.2 census surfaces in-scope sites that change a milestone's shape, or scope drifts toward porting in M0 → stop, surface, replan before continuing. |
| HS-PRE-1 | (Recorded for downstream.) No authz-recording read inside a lock-holding tx. M0 writes no such code; the ADR documents the rule for M3's consumption. |
