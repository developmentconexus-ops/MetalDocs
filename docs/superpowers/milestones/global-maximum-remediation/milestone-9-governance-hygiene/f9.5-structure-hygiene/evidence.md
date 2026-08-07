# F9.5 — structure-hygiene: evidence

> Feature folder: `f9.5-structure-hygiene/`. Contract: `../validation-contract.md` §5. Executed 2026-07-06.
> Mini-gate: `docs/superpowers/analysis/2026-07-06-f95-approval-structure-system-impact.md` (YELLOW).
> ADR: `wiki/decisions/0072-approval-nested-exception-and-boundary-model.md`.

## 1. Mini-gate verdict (input, not re-derived)

YELLOW — proceed with (1) ADR-recorded nested exception for `documents/approval` (promotion
rejected), (2) mechanical `repository/`→`infrastructure/` rename, (3) boundary-guard realignment to
REQ-TOP-1 published-surface model, (4) zero contract/schema/capability diffs. Full text in the
mini-gate file above; not reproduced here.

## 2. Task 1 — violation census

Baseline run (old script, old domain-only allow-model), captured before any rename:

```
powershell -File scripts/check-module-boundaries.ps1
```

Result: `[module-boundaries] FAIL`, **53 flagged lines** (spec/plan referenced "~55"; exact count
re-measured 2026-07-06 is 53). Classification of all 53:

| Class | Count | Disposition |
|---|---|---|
| `iam/authz` (tier-2 in-tx capability check) | 42 | Sanctioned published package |
| `render/fanout` (incl. `render/fanout/dispatchjobs`) | 4 | Sanctioned published package |
| `render/resolvers` | 3 | Sanctioned published package |
| `iam/application` (cross-module application-service consumption) | 2 | Already-allowed layer under the realigned model (`application`) |
| `auth/application` (iam→auth) | 2 | Already-allowed layer under the realigned model (`application`) |

**True cross-module violations found: 0.** The mini-gate's system-impact analysis named
`internal/modules/jobs/stuck_instance_watchdog/job.go` importing `approval/repository` as a known
violation. Re-verified against the file's actual import list at execution time (2026-07-06): it
imports `metaldocs/internal/modules/documents/approval/application` only — an already-sanctioned
layer. This finding is corrected in ADR 0072 rather than silently dropped.

A supplemental cross-module import scan (custom script, not the boundary guard, run to double-check
the census independent of the tool being replaced) found **188 total cross-module import lines**
across production code; of those, exactly **one** targets a layer outside
`{domain, application, api, authz, fanout, resolvers}` —
`internal/modules/documents/delivery/http/handler.go` → `documents/approval/repository`
(now `.../infrastructure`) — which is the `documents`↔`documents/approval` nested-family edge,
allowed under ADR 0072's exception (both sides are the documents bounded context).

## 3. Task 2 — renames (mechanical, compiler-guided, per-module checkpoints)

| Module | Move | Importers updated | `go build ./...` checkpoint |
|---|---|---|---|
| documents | `git mv internal/modules/documents/repository → .../infrastructure`; package `repository`→`infrastructure` | 39 files across `apps/`, `internal/modules/documents/**`, `internal/modules/controlleddocuments/**`, `internal/modules/jobs/**`, `internal/composition/tenantdata/registry/registry.go` | PASS (exit 0) |
| templates | `repository/*` folded into existing `infrastructure/` (no collisions: 7 repository files vs 3 pre-existing infrastructure files, all distinct names); package unified `infrastructure` | 6 files (`templates/jobs/orphan_object_sweeper*.go`, `internal/composition/tenantdata/registry/registry.go`, `apps/api/cmd/metaldocs-api/main.go`) | PASS (exit 0) |
| documents/approval | `repository/*` folded into existing `approval/infrastructure/` (idemp stores + `signature/`); package unified `infrastructure` | 42 files (all `approval/application`, `approval/http`, `approval/jobs` production+test files, `documents/delivery/http/handler*.go`, `iam/integration_test.go`, `internal/composition/tenantdata/registry/registry.go`, `apps/api`+`apps/jobs` mains) | PASS (exit 0, after cycle fix below) |

**Deviation flagged loudly (resolved, mechanical, no interface redesign):** folding approval's
`repository/*` directly into the same `infrastructure` package as the pre-existing
`postgres_route_admin_idemp_store.go`/`postgres_signoff_idemp_store.go` created a real Go import
cycle (`infrastructure` idemp stores → `application` port interfaces; `application` services →
`infrastructure.ApprovalRepository`, an interface defined in the persistence package). Per the hard
rule ("if a rename forces anything beyond mechanical → STOP, debt-list, flag loudly"), this was
stopped and resolved by moving the two idemp-store files (+ test) into a new subpackage
`internal/modules/documents/approval/infrastructure/idempotency/`, mirroring the pre-existing
`infrastructure/signature/` subpackage convention — zero interface/behavior/signature changes; one
external caller (`apps/api/cmd/metaldocs-api/main.go`) updated to the new import path. Full narrative
in ADR 0072 §(b).

Stale doc comments (`// Package repository ...`) on 2 files updated to `// Package infrastructure ...`
(comment-only, zero behavior change): `internal/modules/templates/infrastructure/postgres.go`,
`internal/modules/documents/approval/infrastructure/approval_repository.go`,
`internal/modules/documents/approval/infrastructure/idempotency/idemp_store_helpers.go`.

One hardcoded test path fixed (mechanical): `internal/modules/documents/delivery/http/handler_test.go`
read `../../../../modules/documents/repository/repository.go` as a source-content assertion; updated
to `.../modules/documents/infrastructure/repository.go`.

Full-tree checkpoints after all three module moves:
```
go build ./...   → exit 0
go vet ./...     → exit 0 (compiles every _test.go across the whole repo)
```

## 4. Task 3 — true-violation swaps vs debt list

**Swaps performed: 0** (none needed — census found zero true cross-module violations in production
code after the rename).

**Debt list: empty.** No entries added to `$debtAllowList` in `scripts/check-module-boundaries.ps1`.
Recorded formally in ADR 0072's debt table.

## 5. Task 4 — guard realignment + four proofs

`scripts/check-module-boundaries.ps1` rewritten: module identity = first path segment under
`internal/modules/`, except `documents/approval/*` (treated as identity `documents/approval`, with
`documents`↔`documents/approval` edges always internal in both directions per the ADR 0072 nested
exception). Allowed cross-module targets: layer ∈ `{domain, application, api}`, or an explicit
published-package entry (`iam/authz`, `render/fanout`, `render/fanout/dispatchjobs`,
`render/resolvers`). Everything else forbidden unless in the (currently empty) `$debtAllowList`,
each entry of which must cite an ADR anchor.

**Proof 1 — RED on pre-fix baseline** (old script/model, captured before any F9.5 edit):
```
[module-boundaries] FAIL
Violacoes encontradas:
 - internal/modules/controlleddocuments/application/service.go -> metaldocs/internal/modules/iam/authz
 ... (53 lines total, see §2 census)
```

**Proof 2 — GREEN on final tree** (new script/model, post-rename):
```
[module-boundaries] OK
```

**Proof 3 — negative plant.** Added a blank import
`_ "metaldocs/internal/modules/documents/approval/infrastructure"` to
`internal/modules/jobs/stuck_instance_watchdog/job.go` (genuinely external module — `jobs`, not
`documents`):
```
[module-boundaries] FAIL
Violacoes encontradas:
 - internal/modules/jobs/stuck_instance_watchdog/job.go -> metaldocs/internal/modules/documents/approval/infrastructure
```
Correctly named the planted violation exactly.

**Proof 4 — revert.** Removed the planted import; verified:
```
git diff --exit-code internal/modules/jobs/stuck_instance_watchdog/job.go   → exit 0 (clean)
powershell -File scripts/check-module-boundaries.ps1                        → [module-boundaries] OK
```

## 6. Task 5 — ADR

`wiki/decisions/0072-approval-nested-exception-and-boundary-model.md` (next free number after 0071).
Contains: (a) approval nested exception + external surface (`domain`/`application`/`api` only) +
promotion trigger, (b) guard-model realignment rationale + the cycle-fix deviation narrative, (c)
empty debt table with the mechanism documented for future entries. Status field: 1 line, 33 chars
(well within the F9.1 ≤3-line/≤400-char rule). Index row added at
`wiki/decisions/index.md` (after 0071). CLAUDE.md footnote placeholder
`(ADR: approval nested exception, F9.5.)` → `(ADR 0072.)` — single-line edit, verified via
`git diff --stat CLAUDE.md` (1 file changed, 1 insertion, 1 deletion).

ADR status-field sweep re-run on the full `wiki/decisions/` tree after adding 0072:
```
cd wiki/decisions && for f in [0-9]*.md; do awk ... ; done
→ (no output — 0 violations, 72 files swept)
```

## 7. Task 6 — gate outputs

| Gate | Command | Result |
|---|---|---|
| 1. No `repository/` dirs | `find internal/modules -type d -name repository` | empty (pass) |
| 2. Build | `go build ./...` | exit 0 |
| 2. Targeted tests | `go test ./internal/modules/documents/... ./internal/modules/templates/... -count=1` | all PASS (unit tier, no `-tags integration`); integration-tagged files (`//go:build integration`) present in `documents/infrastructure`, `templates/infrastructure`, `approval/infrastructure` — **not run** (DB-dependent, box constraint, being repaired in a separate session per operator note); labeled honestly here as skipped-with-reason, not silently green |
| 3/4. Boundary guard GREEN + negative proof | see §5 | GREEN; 4 proofs captured |
| 5. ADR + sweep | see §6 | 0 violations |
| 6. api-lint blocking | `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | `0 violation(s)` (was 40 before path-key fixes to `tripwire-allowlist.txt`/`seed-chokepoint-allowlist.txt` — those allowlists reference source files by path:function/path:line and needed the same mechanical path update as every other importer; no rule/behavior change) |
| 6. api-lint unit tests | `go test ./scripts/api-lint/... -count=1` | PASS |
| 7. Diff class | `git status --short` (129 entries) | moves/renames (documents, templates, approval `repository`→`infrastructure`, approval idemp-store split into `infrastructure/idempotency/`) + import-path updates in movers' callers + `scripts/check-module-boundaries.ps1` + `scripts/api-lint/*-allowlist.txt` path keys + ADR 0072 + `wiki/decisions/index.md` + CLAUDE.md 1-line footnote. Zero `api/openapi/`, `migrations/`, or `db/` diffs (`git status --short \| grep -i "api/openapi\|migrations\|/db/"` → empty). `docs/release/` untracked directory pre-exists this session and was not touched. |

## 8. Forbidden self-check (spec + contract §7)

- No `api/openapi/` change — confirmed empty grep.
- No `migrations/`/`db/` schema change — confirmed empty grep.
- No capability/authz edit — zero `authz.Require`/capability registry files touched; only import
  paths and one package-cycle-driven file relocation (idemp stores), no interface/signature change.
- No ADR history deleted or summarized away — 0072 is new; no other ADR content edited.
- No gate/lint/test weakened — api-lint invocation, flags, and strictness unchanged; boundary guard
  is **stricter** (adds real enforcement of REQ-TOP-1, was previously a domain-only false-negative
  net); no test assertions altered, only import/package-decl mechanics and one hardcoded test path.
- No `docs/release/` or `docs/superpowers/plans/` content committed by this feature (untracked,
  untouched).
- No `.env` material read, printed, or committed.
- No push to any remote.
