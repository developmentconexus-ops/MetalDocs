# Feature F0.2 — Plan — binding re-census against ADR-0039

> **Milestone:** 0 · **Folder:** `f0.2-binding-census` · **Status:** Done
> **Spec:** `spec.md` (consumer contract + Validation Gate). **Rule applied:** `wiki/decisions/0039-*.md`.

Static-analysis feature: the "build" is the grep+inspect sweep; the "tests" are the per-site re-greps and the
owner-map diff recorded in `census.md`. No production code touched.

## Source

- Milestone spec row F0.2 (`../milestone.md`): *implement* — `census.md` (table→owner map, authoritative
  in-scope list reproducing the ~20 + any new sites, brief-delta, coverage statement, `unclassified: 0`).
  *Validate* — every brief site reproduced; owned-table set widened beyond the named tokens; 0 unclassified;
  new shape-changing sites → HS-6.
- Governing spec: `../mission.md` §4 (deferred-to-M0 assumptions), §5 (inventory), §2 (Non-Goals).

## Plan (executed)

1. **Owner map (Step A)** — grep every `INSERT INTO`/`UPDATE`/`DELETE FROM` target across
   `internal/modules/**` non-test; the mutating module = the table's owner. Hand-curate out grep noise
   (prose words captured by `UPDATE <word>`). → owner map table in `census.md`.
2. **Read map (Step B)** — grep every `FROM`/`JOIN`/`EXISTS` target; record (file:line, table, reader module).
3. **Cross-module diff** — reader ≠ owner ∧ table ∈ owned set ⇒ candidate H-G. Filter same-module.
4. **Per-site inspect (Step C)** — open each candidate's SQL + tx context; classify per ADR-0039 D1/D3;
   correct ownership facts (`document_process_areas`→taxonomy; `auth_failure_counters`→documents/approval).
5. **Reproduce the brief's ~20** — re-grep each cited token at its `file:line`; confirm none dropped.
6. **Widen completeness sweep** — sweep the full owned-table set (iam_users, templates_template(_version),
   document_comments/families, auth_*, audit_events, approval_*) for non-owner reads.
7. **Write `census.md`** — owner map, Part 1 (mission-scoped ~20), Part 2 (NEW: N1 + X1–X8), coverage
   statement (`unclassified: 0`), brief-delta.
8. **HS-6 gate** — the widen surfaced N1 (M2 shape change) + X1–X8 (Non-Goal/terminal-bar contradiction).
   Wrote `hs-6-scope-decision.md`; STOPPED; surfaced both decisions to the operator (one question each).
9. **Apply the ruling** — operator: 1a (fold N1 into M2) + 2a (ADR-0039 exemption). Amended `wiki/decisions/
   0039-*.md` (D3(d)–(f) + N1/X table + honest "0-outside-allowlist" scope note); replanned mission §2/§4/§5
   (N1 row 16, exemption note); recorded resolution in `hs-6-scope-decision.md` + `census.md`.

## Files touched

- `census.md` (new) — the census of record.
- `hs-6-scope-decision.md` (new) — the HS-6 surface + operator resolution.
- `wiki/decisions/0039-cross-module-base-table-read-boundary.md` (amended — D3(d)–(f), N1/X table, scope note,
  header amendment stamp). **This is an HS-4 → F0.1 amendment triggered by F0.2's findings**, recorded here
  and reflected in F0.1's evidence addendum.
- `../mission.md` (replan — §2 Non-Goal, §4 discovery summary status, §5 row 16 + out-of-scope resolution).
- `../../milestone-0-adr-and-census/f0.1-adr-0039/evidence.md` (addendum noting the amendment).

## Test strategy

No automated tests (static artifact). Acceptance = the `spec.md` Validation Gate: per-site re-grep over the
live tree (real, recorded in `census.md`), owner-map diff vs a full-tree read grep, explicit `unclassified: 0`,
brief-delta present, coverage statement with named residual. No fixture; the rigor is the per-site re-grep.

## Execution notes

- Main-session execution (a census is judgement-dense; fan-out would not pay). Model: main (Opus).
- Scope held: **measured only** — zero production SQL edited. The ADR amendment is a definition change
  (operator-ruled), not a port. N1's actual port lands in M2; X1–X8 are exempt, not touched.
- No runtime/Docker: static read; residual (dynamic/aliased SQL) recorded, not reproduced. No false green.
- HS-6 honored as designed — STOPPED and surfaced before finalizing F0.3, because F0.3's allowlist depends on
  the ruling.