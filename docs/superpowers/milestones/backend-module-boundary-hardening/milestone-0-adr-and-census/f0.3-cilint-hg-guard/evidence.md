# Feature F0.3 — Evidence — cilint `hgcrossmodule` H-G guard

> **Milestone:** 0 · **Feature:** `f0.3-cilint-hg-guard` · **Closed:** 2026-06-20
> **Contract:** `spec.md` (consumer contract + Validation Gate). **Mechanizes:** ADR-0039 D6.

## What was implemented

- **`tools/cilint/internal/analyzers/hgcrossmodule.go`** — the `HGCrossModule(files []string) []Finding`
  analyzer (sibling of `noResponseMap`). Detects a raw `FROM`/`JOIN` against another **top-level** module's
  owned base table, in a non-owner package, via an **AST-BasicLit-STRING-only** scan (comment/identifier
  mentions structurally cannot flag). Suppressed by `hgPendingRemediation` (M1–M4 debt ledger = the F0.2
  in-scope sites), `hgExempt` (permanent ADR-0039 D3(d)–(f) carve-outs = X1–X8), or `//cilint:allow-hgcrossmodule`.
  - `hgOwnerByTable` = the census table→owner map (top-level module; `documents/approval ⊂ documents`,
    `iam/presence ⊂ iam`).
- **`hgcrossmodule_test.go`** — 8 bite/green tests.
- **`analyzers.go`** — `HGCrossModule(files)` appended to `RunAll`.

## Verification (commands + real output)

| Check | Command | Result | Real vs fixture |
|-------|---------|--------|-----------------|
| TDD red (test authored first) | `go test ./tools/cilint/internal/analyzers/ -run TestHGCrossModule` | **FAIL [build failed]** — `undefined: analyzers.HGCrossModule` (×8) | real |
| TDD green (8/8) | same, after implementing | `PASS` ×8, `ok metaldocs/tools/cilint/internal/analyzers` | real (fixtures) |
| Full cilint suite | `go test ./tools/cilint/...` | `ok metaldocs/tools/cilint/internal/analyzers` | real |
| `go build ./...` | `go build ./...` | clean (`BUILD OK`) | real |
| **Full-tree guard green** | `go run ./tools/cilint ./...` | **exit 0** — every live cross-module read ∈ pending ∪ exempt | **real (live tree)** |
| **Baseline is load-bearing** (not a vacuous green) | drop the B1 entry, `go run ./tools/cilint ./...` | **exit 1** — `internal\modules\documents\repository\repository.go:1701: [hgcrossmodule] module "documents" reads "controlleddocuments"'s base table "controlled_documents" with raw SQL (H-G, ADR-0039 D1)…` — fires at the exact real site/line; restored after | **real (live tree)** |

> The drop-B1 experiment is the proof the green is earned: the guard *does* detect the real in-scope reads;
> exit 0 on the full list means the F0.2 baseline is **complete** (no orphan cross-module read the analyzer
> can see), not that the regex matched nothing.

## Acceptance vs spec Validation Gate

| Acceptance criterion (spec.md) | Met? | Evidence |
|--------------------------------|------|----------|
| Flags a cross-module base-table read | yes | `TestHGCrossModule_Positive_CrossModuleRead` PASS; drop-B1 real fire |
| Does NOT flag own-table read | yes | `TestHGCrossModule_Negative_OwnTable` PASS |
| Does NOT flag sub-package as cross-module | yes | `TestHGCrossModule_Negative_SubpackageSameModule` PASS (documents/approval ⊂ documents) |
| Does NOT flag comment-only foreign mention | yes | `TestHGCrossModule_Negative_CommentMention` PASS (AST-BasicLit scope) |
| Pending-baseline suppresses a known site | yes | `TestHGCrossModule_Negative_PendingBaseline` PASS |
| Exempt suppresses an X-site | yes | `TestHGCrossModule_Negative_Exempt` PASS |
| Inline allow suppresses | yes | `TestHGCrossModule_Negative_AllowDirective` PASS |
| Full-tree run green today | yes | `go run ./tools/cilint ./...` exit 0 |
| Analyzer + suite compile and pass | yes | `go test ./tools/cilint/...` ok; `go build ./...` clean |

## Review / QA disposition

- Self-review vs `spec.md` consumer contract: analyzer name/signature, owner map, both allowlists, inline
  directive, `RunAll` wiring — all present. Code-quality: mirrors the established `noResponseMap` shape
  (AST scan, slash-normalized path matching, `getLine`/`readSource`/`parseFile` reuse). `gofmt` clean.
- The independent `milestone-validator` (Phase 4) re-runs `go test ./tools/cilint/...` and `go run ./tools/cilint
  ./...` from clean state as the separation-of-powers gate.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| `hgPendingRemediation` non-empty (16 in-scope sites still listed) | **By design** — this is the debt ledger M1–M4 drain. Emptying it in M0 would be a false green (spec non-goal). | **Trigger:** each of M2 (B-sites + N1), M3 (C-sites), M4 (C4-sites) removes its entries on porting; terminal §8 asserts the slice is empty. **Owner:** M1–M4 + mission-validator. |
| Dynamic / aliased cross-module SQL invisible to the literal-token scan | Same residual as the H-D guard; recorded in the F0.2 census coverage statement | **Trigger:** any reported runtime cross-module read the guard missed → extend the analyzer. **Owner:** mission. |
| `hgExempt` X-sites are file+table-coarse (a future *new* read of the same table in the same file would also be suppressed) | Coarseness is inherent to a file+table allowlist; the alternative (line numbers) drifts; the X-files are stable, low-churn | **Trigger:** any material new SQL in an X-file → re-census that file. **Owner:** mission-validator spot-check. |