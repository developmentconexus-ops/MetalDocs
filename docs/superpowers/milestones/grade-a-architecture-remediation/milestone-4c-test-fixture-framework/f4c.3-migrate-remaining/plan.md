# Feature F4c.3 — Plan

> **Spec:** `spec.md` (approved 2026-06-15) · **Folder:** `f4c.3-migrate-remaining`

The "how" for the contract in [`spec.md`](spec.md). Execution model: Workflow fan-out, one subagent
per cluster, pipeline (migrate → per-cluster test verify, no barrier between stages so fast clusters
land evidence while slow ones are still verifying).

## Concurrency + model split

- **Cap:** ≤8 concurrent (template-DB clone is IO-heavy on the postgres data dir; Windows C: SSD
  degraded — see memory `machine-ssd-degraded-writes`).
- **sonnet** (judgment): **C2** (snapshot+context_builder — local-helper deletion + factory-composite
  pick), **C4** (templates — owns the 2nd local helper + 4 files crossing version-state semantics),
  **C8** (pgtest classification — read intent of 7 files, decide stateful vs no-write).
- **haiku** (mechanical): **C1**, **C3**, **C5**, **C6**, **C7** — regex-shaped rewrite of inline
  `set_config` blocks / hardcoded tenant UUID / bare `documents` into factory calls.

## Per-cluster subagent brief (template)

Each subagent receives:
1. **Mission:** "Migrate `<file list>` onto `tests/integration/testdb/factory.go`. Fix-not-adapt:
   `db.go` MUST stay at HEAD; tripwire MUST NOT be weakened."
2. **Consumer-contract pointer:** `tests/integration/testdb/factory.go` — read API from there; do not
   extend it.
3. **Rules the file MUST satisfy after migration** (F4c.4 guard preview):
   - No `SELECT set_config('metaldocs.asserted_caps', …)` inline (use factory or
     `testdb.SeedWithCaps`).
   - No `is_local=false`.
   - No hardcoded tenant-UUID literal (factory mints fresh UUIDs).
   - No bare unqualified `documents` (use `testdb.Qualified("documents")` or factory-returned IDs).
4. **TDD step:** capture pre-migration test failure / red baseline (one `go test -tags integration
   -count=1 -run <Name> ./<pkg>/...` run), then migrate, then re-run GREEN. Real DB only.
5. **Hard-stops:**
   - **HS-2** — if migration surfaces a Family-A schema defect (production-source fix needed),
     **stop the cluster**, write a partial `cluster-<id>-report.md` describing the boundary +
     minimum prerequisite plan. Do NOT edit `internal/...` or `db/`.
   - **HS-6** — if migration tempts re-introducing inline seeding, editing `db.go`, weakening the
     tripwire, or touching an F4c.2 file → stop, surface.
6. **Return value (StructuredOutput):**
   ```jsonc
   {
     "clusterId": "C3",
     "status": "GREEN" | "HS2" | "HS6" | "RED",
     "filesMigrated": ["path/a_test.go", ...],
     "helpersDeleted": ["seedX", ...] | [],
     "testCmd": "go test -tags integration -count=1 ...",
     "testOutputTail": "...last 40 lines...",
     "grepProofs": {
       "set_config": 0, "is_local_false": 0, "tenant_uuid_literals": 0, "bare_documents": 0
     },
     "dbGoDiffEmpty": true,
     "hsBoundary": null | "HS-2 description w/ minimum prerequisite plan"
   }
   ```

## Stages (Workflow pipeline)

```
Stage 1 (migrate):  cluster.subagent(...) → returns StructuredOutput
Stage 2 (verify):   main-session reads structured result; if status==GREEN AND grepProofs all 0 AND
                    dbGoDiffEmpty → label cluster GREEN; else record for HS-2/HS-6 review.
```

No barrier between stages — fast cluster (C5/C6) lands GREEN while slow cluster (C4 / C8) still in
stage 1.

## Convergence (main session, after fan-out)

1. Aggregate StructuredOutputs → write `evidence.md` w/ per-cluster row.
2. Run AC6 full-suite: `go test -tags integration -count=1 ./...` from clean baseline; capture tail.
3. Run AC7 M4-blocker regression `-run` set; capture tail.
4. Run AC3 grep proofs across the whole migrated tree (not just per-cluster).
5. Run AC2: `git diff --exit-code tests/integration/testdb/db.go`.
6. Run AC8: `git diff --name-only origin/main...HEAD -- internal/ db/` → expect empty unless an
   HS-2 was approved.
7. If any cluster returned HS-2 → surface to operator; do NOT commit until each HS-2 is resolved
   the F4c.2 / fillin way (forward migration + ADR + factory migration re-run).
8. If everything GREEN → single squashed commit per cluster cluster-batch (or one commit for all
   GREEN clusters w/ named files in body), then close `evidence.md`.

## Rollback / safety

- Each cluster's edits are isolated to its file list. A failed cluster doesn't block GREEN clusters
  from committing (per spec Q6).
- The fix-not-adapt invariant (empty `db.go` diff) is checked **per cluster** and **on aggregate**.
  Any drift = revert that cluster.
- No subagent has Write access to `internal/...` or `db/migrations/...` in its brief — only test
  files + (within scope) the factory file (which it MUST NOT edit). Enforce in prompt.

## Out of plan

- F4c.3b iam-membership micro-task (Q4 defer).
- F4c.4 CI grep-guard.
- F4c.5 docs + ADR + wiki-curator dispatch.
