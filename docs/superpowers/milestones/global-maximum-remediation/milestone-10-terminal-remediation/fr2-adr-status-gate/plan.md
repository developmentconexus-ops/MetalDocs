# F-R2 Plan

## Files touched
1. `scripts/check-adr-status.sh` (new) — single-source sweep; `[dir]` arg; exit 0/1/2.
2. `.github/workflows/governance-check.yml` — blocking step in the `check` job invoking the script.
3. `wiki/standards/documentation-governance.md` — replace the "optional future extension" text with
   the CI-enforced statement + point the manual sweep at the script.

## Order
1. Write the script (logic lifted verbatim from the doc's awk one-liner).
2. Negative proof against a synthetic over-budget fixture in a temp dir → exit 1.
3. Positive proof against `wiki/decisions/` → exit 0.
4. Wire the blocking CI step.
5. Update the doc.

## Test strategy
- The `[dir]` argument makes the gate unit-testable: the negative proof points it at a `mktemp -d`
  fixture, so no throwaway ADR is ever committed to `wiki/decisions/`.
- Positive proof is the real tree (already clean post-F9.1) — proves no false positive.

## Risk / rollback
- Minimal. New script + one CI step + doc text. Rollback = revert the 3 files.
- CRLF risk on the `.sh` (Windows authoring): normalized to LF + `chmod +x`.
