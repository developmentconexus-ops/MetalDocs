# F9.3 — test-policy (plan)

> Input: `spec.md` (approved). Executor: sonnet subagent, main session reviews.

## Plan

### Task 1 — policy doc
Write `wiki/quality/legacy-test-policy.md`:
- Taxonomy table (repair-class triggers: REQ-ID guard / tripwire arm / contract shape / DB invariant;
  delete-class: one-off scaffolding) + decision flowchart (text).
- Procedure: classify → repair on testdb factory (ADR 0034, hard gate) or delete with commit rationale
  `test: delete one-off scaffolding <name> (legacy-test-policy delete-class)`.
- ≥2 worked examples FROM THIS REPO: pick one clear repair-class (e.g. a tripwire/RLS/contract test)
  and one delete-class candidate (a drive-a-fix scaffold), citing real paths.
- Anti-pattern appendix: NumGoroutine lifecycle assertions in parallel suites (TST-04).
- Links: add row/link in `wiki/quality/index.md`; cross-link from `test-discipline.md` (small
  "related policies" pointer — do not restructure that doc).

### Task 2 — pick + freeze measurement set
Choose ≥3 heavy DB-integration packages (candidates: `internal/modules/documents/repository`,
`internal/modules/documents/approval/repository`, `internal/modules/templates/repository`; adjust to
post-F9.5-independent paths — F9.3 runs BEFORE F9.5, use current paths). Record chosen set + baseline:
`go test <pkgs> -count=1` wall-clock (capture full output + real time), on this box, real Postgres up
(`.\scripts\start-api.ps1` stack prerequisites per local-dev-startup; testdb factory needs its DB).

### Task 3 — t.Parallel expansion
For each file in the set (and other safe integration files, breadth allowed within contract §3.2):
- Add `t.Parallel()` to top-level tests + subtests where the test uses testdb-factory-per-test DB or
  is stateless. Skip + record (file, reason) where: t.Setenv, shared package fixture, global mutation,
  port binding, ordered dependence.
- No other edits.

### Task 4 — measure after + green gate
Re-run identical command; capture. Compute delta. If ≤0 gain, record honest number + analysis.
All touched suites green.

### Task 5 — evidence.md
Fill per contract §8: doc path+links, before/after outputs, serial-files table, diff scope.

## Files touched
`wiki/quality/legacy-test-policy.md` (new), `wiki/quality/index.md`, `wiki/quality/test-discipline.md`
(link only), `internal/**/ *_test.go` (t.Parallel lines only), feature folder.

## Test strategy
The measurement runs ARE the test. Green-after is the regression gate. TDD analog: baseline timing
captured before edits (the "failing" measurement), improvement verified after.
