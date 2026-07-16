# MetalDocs Development Harness — SUPERSEDED POINTER

**Status: SUPERSEDED 2026-07-16** (operator-directed harness reorganization, at the clean
unit-4.5 close boundary — per the harness-hub swap rule for legacy combined harness files).

The binding doctrine is now TWO files read together:

1. **Core (method):** `HARNESS-CORE.md`, shipped by the `mnfs-harness` plugin (v0.2.0) —
   model matrix, execution loop P1–P8, event grammar, six-axis collision matrix, anti-slop
   contract, verification-ladder semantics, failure protocol, codex dispatch paths. Companion
   files: `PROFILE-TEMPLATE.md`, `REVIEW-STANDARD.md` (binding for every code review), skills
   `harness-hub` / `harness-worker` / `harness-init` / `codex-dispatch`.
2. **Profile (repo bindings):** [`docs/HARNESS-PROFILE.md`](../HARNESS-PROFILE.md) — exact
   ladder commands, fresh-workspace bootstrap + false-alarm signatures, testdb strategy,
   collision-axis instantiation, shared seams, non-negotiables, truth order, human gates,
   superseded-protocols denylist. Profile wins over core for in-flight missions.

**Queue:** the active program's ordered unit queue stays `docs/superpowers/ROADMAP.md`
(legacy mission layer). A future mission planned via `/mission-init` moves the queue to
`.mnfs/MIS-*/` per core layer 3.

Everything the 2026-07-10..14 version of this file ruled is either (a) covered by core 0.2.0
— which is NEWER (codex worker matrix + sonnet fallback, batch P2 planning with write-sets +
contract-satisfiability, `COMMITTED` event, additive contract-locks, REVIEW-STANDARD) — or
(b) migrated into the profile with provenance. Do not cite this file as doctrine.

---

## Method addenda — pending upstream to HARNESS-CORE

Repo-authored METHOD content not yet in core 0.2.0. Binding here until upstreamed
(core §0 amendment protocol, classification: method-level).

### GM fork procedure (binding when ≥2 implementations compete)

Trigger: at spec time OR mid-implementation, two viable shapes exist — typically A = improve
in-place on the current base (local maximum) vs B = the right structure, but it changes the base.
Never pick silently; run this:

1. **Judge the base first.** Sound base → A and B both legitimate, judge on merit. Bad base
   (legacy/patch/workaround) → step 2 rules.
2. **Write the fork record** (unit spec or evidence — 5 lines): the options, which is the
   global maximum and why, cost of each, whether B crosses the unit boundary (module ownership,
   contract, DB shape, budget).
3. **Route:** B fits inside boundary + budget → **B, always** ("A is faster" never beats
   structure). B crosses the boundary → STOP, `BLOCKED` with the fork record; hub/operator
   picks: expand the unit to B / file B as a named queue unit + minimal bridge slice / accept A
   as named debt with an ADR-recorded reopen trigger.
4. **No-deepening rule.** A bridge on a base marked-for-replacement is minimal and reversible:
   may touch the bad base, may NOT grow it (no new capabilities, abstractions, or consumers on
   the structure B will delete).
5. **Reversibility veto.** If choosing A makes B materially more expensive later (new callers of
   the wrong seam, data in the wrong shape, contract surface that must break), A is FORBIDDEN at
   worker discretion — operator sign-off only.
6. **Named debt or it didn't happen.** Any accepted local maximum gets a debt entry naming the
   reopen trigger. An unnamed local max discovered at review = reject.

Field proof: unit 4.5 round-2 FAIL — budget bumps (60s→5m→10m) on the GC hang were the A-path;
the ratified fix A (`classifyGCCandidate` extraction, 572f6827) was the B-path and closed the
defect class permanently.
