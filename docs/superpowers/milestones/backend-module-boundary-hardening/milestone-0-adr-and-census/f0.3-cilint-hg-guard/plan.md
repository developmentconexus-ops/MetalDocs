# Feature F0.3 — Plan — cilint `hgcrossmodule` H-G guard

> **Milestone:** 0 · **Folder:** `f0.3-cilint-hg-guard` · **Status:** Done
> **Spec:** `spec.md`. **Engine:** subagent-driven-development inline (single-file analyzer, TDD red→green).

## Source

- Mission §5 row 1 (ADR-0039 CI grep-guard) + ADR-0039 D6 (CI-enforced) + F0.2 census (owner map +
  in-scope/exempt site sets). Sibling of the H-D `noResponseMap` analyzer.

## Plan (executed, TDD)

1. **Read the H-D sibling** — `noresponsemap.go` (AST-BasicLit scan, exempt-file list, inline directive,
   `inRegisteredRoutePackage`) + `analyzers.go` (`RunAll`, `Finding`, `collectGoFiles` which already excludes
   `_test.go`/vendor/dotdirs). Mirror its shape.
2. **Write failing tests first** — `hgcrossmodule_test.go`: 1 bite (cross-module read flagged) + 7 green
   (own-table, sub-package-same-module, comment-only mention, pending-baseline, exempt X-site, inline allow,
   non-module file). Ran `go test -run TestHGCrossModule` → **red** (`undefined: analyzers.HGCrossModule`).
3. **Implement `hgcrossmodule.go`:**
   - `hgOwnerByTable` — the census table→**top-level**-module owner map (documents/approval ⊂ documents).
   - `hgModuleOf(path)` — first segment under `internal/modules/`.
   - AST walk of `*ast.BasicLit` STRING nodes only (avoids comment/identifier false positives the census
     found at `people_service.go:690`, `observability_repository.go:164`).
   - `hgFromJoin` regex — `FROM`/`JOIN` + optional `public.`/`metaldocs.` + table token; `\s` spans newlines.
   - Flag when `owner ≠ reader` and the (file,table) is on neither `hgPendingRemediation` (M1–M4 debt ledger)
     nor `hgExempt` (permanent D3(d)–(f)) and the line lacks `//cilint:allow-hgcrossmodule`.
   - Per-(line,table) dedupe; message names reader, owner, table, ADR-0039.
4. **Wire into `RunAll`** — append `HGCrossModule(files)`.
5. **Green the unit suite** — `go test ./tools/cilint/...` → all 8 pass.
6. **Green the full tree** — `go run ./tools/cilint ./...` → exit 0 (every live cross-module read ∈
   pending ∪ exempt).
7. **Prove the baseline is load-bearing** — temporarily dropped the B1 entry → guard went **red** at the
   exact real site `documents/repository/repository.go:1701` with the correct message → restored. This rules
   out a false green from a non-matching regex.

## Files touched

- `tools/cilint/internal/analyzers/hgcrossmodule.go` (new — analyzer + owner map + two allowlists).
- `tools/cilint/internal/analyzers/hgcrossmodule_test.go` (new — 8 bite/green tests).
- `tools/cilint/internal/analyzers/analyzers.go` (one line — wire into `RunAll`).

## Test strategy

TDD: failing tests first (red, compile error), implement to green. Bite test proves detection; green tests
prove no over-flag (own-table, sub-package, comment-only, allowlists, inline directive, out-of-scope). The
**real full-tree run** (exit 0) proves the baseline ledger is accurate and complete; the **load-bearing
drop-B1 experiment** (exit 1 at the real line) proves it is not a vacuous green.

## Execution notes

- Inline main-session TDD (one ~230-line analyzer; fan-out would not pay). Model: main (Opus).
- Scope held: **detection only**, zero production SQL edited. `hgPendingRemediation` starts **full** (all
  in-scope debt) — emptying it is M1–M4's job; doing so here would be a false green (spec non-goal).
- Residual recorded (not engineered away): dynamic/aliased table names behind Go vars — same limitation as
  the H-D guard, noted in the F0.2 census coverage statement.