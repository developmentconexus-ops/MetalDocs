# F9.2 — traceability-gate (plan)

> Input: `spec.md` (approved). Executor: sonnet subagent, TDD, main session reviews.

## Plan

### Task 1 — tool skeleton + extraction (TDD)
`scripts/req-trace/` Go package (`main.go`, `extract.go`, `extract_test.go`).
- Failing test first: parse fixture doc lines → `[]Req{ID, Class(MUST/SHOULD/MAY), Annotation}`.
  Cases: plain `(MUST)`, annotated `(MUST — satisfied Wave 1 (F-01): …)`, `(SHOULD)`, multiline-safe,
  non-REQ bullets ignored, duplicate ID mentions collapsed (first normative line wins).
- Real source path: `wiki/architecture/backend-target-architecture.md` (flag `-doc`, default set).

### Task 2 — evidence scan + map merge (TDD)
`scan.go` + tests: walk `internal/` + `apps/` `*_test.go` for `REQ-[A-Z]+-[0-9]+` literals →
req→[]file. `mapfile.go` + tests: parse `wiki/architecture/req-trace-map.yaml`
(entries: `req`, `kind: commit`, `ref`, `note`; kind `doc` NOT accepted from the map — doc evidence
only auto-derives from the REQ line's annotation, anti-gaming). Commit-hash existence check via
`git cat-file -e` when `.git` present + not shallow; warn-skip otherwise.

### Task 3 — gate + report (TDD)
`report.go` + tests: coverage join → per-REQ row (id, class, evidence kind, pointer). MUST with zero
evidence → exit 1 listing ids. `-write` emits `wiki/architecture/req-traceability.md` (header: local
command, generation note). Default mode regenerates in-memory and diffs vs committed report → mismatch
= exit 1 "stale report".

### Task 4 — populate the map honestly
Run scan; for each uncovered MUST, research REAL evidence: the doc's own annotations, milestone
evidence files (docs/superpowers/milestones/**/evidence.md cite commits per REQ), git log. Populate
req-trace-map.yaml with commit-hash entries + note naming the source (e.g. "M2 F2.1 evidence").
A MUST with genuinely no evidence: DO NOT invent — surface it in the final report to main session
(HS-6 candidate). Generate committed report with `-write`.

### Task 5 — CI wiring
Extend `.github/workflows/governance-check.yml` (inspect its triggers first; if scoped wrong, new
`req-traceability.yml` mirroring module-boundaries.yml shape): setup-go + `go run ./scripts/req-trace`.

### Task 6 — proofs + evidence.md
Positive run; negative plant run (REQ-TST-99); anti-rot run; `go test ./scripts/req-trace/...`;
fill evidence.md per contract §8 row.

## Files touched
`scripts/req-trace/*.go`, `wiki/architecture/req-trace-map.yaml`,
`wiki/architecture/req-traceability.md` (generated), `.github/workflows/governance-check.yml`
(or new workflow), feature folder.

## Test strategy
Pure-Go unit tests per component (fixtures under `scripts/req-trace/testdata/`); the three
entrypoint-level proofs (positive/negative/anti-rot) run the real binary against the real tree.
