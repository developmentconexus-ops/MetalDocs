# F9.5 — structure-hygiene (plan)

> Input: `spec.md` (approved) + mini-gate analysis. Executor: sonnet subagent(s), main reviews.

## Plan

### Task 1 — violation census (read-only, feeds Tasks 2/4/5)
Run current boundary script; classify all 55 edges: (a) sanctioned-published under the new model
(iam/authz, iam/application, render/fanout, render/resolvers, auth/application, cross-module
domain/application/api), (b) true violations (persistence/internal imports across boundaries —
known: stuck_instance_watchdog → documents/approval/repository; find all). Census table → ADR draft.

### Task 2 — rename (mechanical, compiler-guided)
- documents: `git mv internal/modules/documents/repository internal/modules/documents/infrastructure`;
  package `repository`→`infrastructure`; update ~39 importer files (alias `docrepo`-style names may
  stay as aliases — minimal diff).
- templates: move `repository/*` files into existing `infrastructure/`; package rename; collision
  check (`template_version_reader*.go` live there — no name clashes expected with postgres.go,
  mappers.go, tenant_data_port.go, list_* tests); update 6 importers.
- approval: `git mv .../approval/repository .../approval/infrastructure_repo-merge` → fold into
  `approval/infrastructure/` (existing files: idemp stores + signature/); package unify
  `infrastructure`; update importers (42 documents-side + external).
- `go build ./...` after each module's move (checkpoint commits allowed).

### Task 3 — true-violation mechanical fixes
For each Task-1(b) edge with an existing published equivalent (e.g. watchdog can consume an
approval/application service or approval/domain port already exported): swap the import. If no
published equivalent exists → DO NOT create ports (HS-2) → debt list entry (edge, reason, trigger).

### Task 4 — guard realignment (TDD analog)
Rewrite the allow-model in `scripts/check-module-boundaries.ps1`:
- module identity = first segment under `internal/modules/`, EXCEPT `documents/approval/*` edges:
  intra documents↔approval = internal; external → approval surface list.
- allowed cross-module target layers: `domain`, `application`, `api` (+ explicit published-package
  list: `iam/authz`, `render/fanout`, `render/resolvers`); everything else forbidden.
- debt-list mechanism: explicit `$debtAllowList` at top of script, each entry commented with ADR
  anchor — the ONLY sanctioned suppressions.
- Proof order: script RED on pre-fix tree captured (or on planted case), GREEN on final tree,
  negative-plant RED, revert.

### Task 5 — ADR
`wiki/decisions/00NN-approval-nested-exception-and-boundary-model.md` (next free number): decision
(a) approval nested exception + external surface + promotion trigger, (b) guard model = REQ-TOP-1
published-surface, (c) debt table. Status per F9.1 rule; index.md row.

### Task 6 — gates + evidence
Spec Validation Gate 1–7 executed; evidence.md filled per contract §8.

## Files touched
`internal/modules/{documents,templates}/**` (moves+imports), external importer files,
`scripts/check-module-boundaries.ps1`, `wiki/decisions/*` (new ADR + index), feature folder.

## Test strategy
Compiler = completeness oracle for the rename. Boundary script RED/GREEN/negative-plant = the guard's
test. Targeted `go test` for touched modules; DB-dependent suites labeled honestly.
