# Feature F0.1 — ADR-0039: lock the H-G definition + exemption list

> **Milestone:** 0 — ADR-0039: lock the definition + binding census + CI guard  ·  **Folder:** `f0.1-adr-0039`
> **Status:** Done

This is a **documentation** feature (one durable decision → one ADR). No code, no TDD; the "build" is the
authoring, the "tests" are the grep/read assertions in `spec.md`'s Validation Gate.

## Source

- Milestone spec row F0.1: *implement* — ADR `wiki/decisions/0039-*.md`: H-G definition (raw read of another
  module's **base table** = violation), exemption list (D4), active-now membership-view contract, H-PRE-1
  note, supersession/links to ADR 0022/0037/0038. *Validate* — ADR exists, status Accepted; definition +
  exemption list unambiguous (a reviewer can classify any site mechanically); operator-ratified at M0 HS-1.
- Governing-spec reference: `../mission.md` §3 (D1–D6), §7 (F0.1), §10 (constraints).

## Plan

1. **Verify ADR number is free** — `ls wiki/decisions/0039-*.md` empty; MAX existing = 0038, so next = 0039
   (per `wiki/decisions/README.md` numbering rule). ✅
2. **Read the antecedents** — ADR 0022 (authz root cause), 0037 (`effective_to IS NULL`), 0038 (owner-port
   precedent), and the docs-governance status vocabulary (`wiki/decisions/README.md`). ✅
3. **Author `wiki/decisions/0039-cross-module-base-table-read-boundary.md`** following the 0037/0038 house
   style: front-matter blockquote (Status `Accepted 2026-06-20`, Last verified, Deciders, Related ADRs, code
   anchors) → Context → Decision (D1 rule / D2 base-vs-view refinement / D3 exemption list / D4 view contract
   / D5 H-PRE-1 / D6 CI enforcement) → **worked classification table over mission §5 rows 3–15** (the proof
   of unambiguity) → Consequences → Verification.
4. **Sync the decisions index** — `wiki/decisions/index.md`: add the 0039 row; backfill the missing 0038 row
   (governance gap found while editing the file); bump the `Last verified` stamp.
5. **Run the Validation Gate proofs** — the grep/read assertions in `spec.md`; record in `evidence.md`.
6. **Operator ratification** — deferred to the M0 HS-1 gate (post-milestone-validator), per the acceptance
   row. Recorded as a bounded defer with that trigger.

## Files touched

- `wiki/decisions/0039-cross-module-base-table-read-boundary.md` (new)
- `wiki/decisions/index.md` (0039 row + 0038 backfill + stamp)

## Test strategy

No automated tests (doc artifact). Acceptance = the `spec.md` Validation Gate grep/read assertions, run and
recorded in `evidence.md`. The substantive proof of the "unambiguous" criterion is the worked classification
table: every mission §5 site (rows 3–15) carries a verdict + deciding clause, 0 unclassified.

## Execution notes

- Authored directly in the main session (a single ADR; fan-out would not pay). Model: main (Opus).
- Scope held: declared the published view/ports compliant-by-design; did **not** create any view/port/SQL
  (M2–M4 build them). No reinterpretation of ADR 0037 — `effective_to IS NULL` cited verbatim.
- Docs-governance hygiene: backfilled the 0038 index row (pre-existing gap, encountered while adding 0039).
