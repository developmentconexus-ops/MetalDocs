# Feature F8.6 — Plan

> Engine: inline (gate/doc/CI feature — no product TDD). Input: approved `spec.md`.

## Plan

### Files touched
- **CREATE** `tools/cilint/internal/analyzers/noresponsemap.go` — `NoResponseMap` AST analyzer + helpers.
- **CREATE** `tools/cilint/internal/analyzers/noresponsemap_test.go` — positive/negative/exemption/scope tests.
- **EDIT** `tools/cilint/internal/analyzers/analyzers.go` — register `NoResponseMap` in `RunAll`; (drive-by) add the M4c `tests/integration/testdb/` seed harness to `allowedTxPackages`.
- **EDIT** `wiki/architecture/api-contract.md` §5b — widen Part A/B path globs to the full public-route surface; add the mechanical-guard paragraph; add health + declared-dynamic-metrics named exemptions; bump Last-verified.
- **EDIT** `docs/superpowers/milestones/grade-a-completion/mission.md` §8 — scope-amendment note restating H-D/H-G against the true surface + recorded exemptions + CI guard.
- **No workflow edit** — `.github/workflows/invariants.yml`'s existing `cilint` job (`go run ./tools/cilint ./...`) runs `RunAll`, which now includes `noresponsemap`; the guard rides the existing gate.

### Design (analyzer)
Scope = registered-route packages (`*/delivery/http/`, `documents/approval/http/`, `iam/presence/`, `platform/observability/`). Per function: pass 1 collects idents bound to a `map[string]any` composite literal; pass 2 flags a writer call (`writeJSON`/`writeFillInJSON`/`WriteJSON`) whose arg is such a literal or a bound ident. Exemptions: `noResponseMapExemptFiles` (health.go) + inline `//cilint:allow-responsemap`. Laundering-resistant: catches the built-then-written local the historical Grep A missed; ignores non-response maps (never reach a writer).

### Verification strategy
1. Analyzer unit tests (positive direct/laundered/alias, negative typed/non-response/exempt/allow/out-of-scope).
2. Real-repo run: `go run ./tools/cilint ./internal/...` → noresponsemap 0.
3. Planted-violation proof: temp `writeJSON(w,200,map[string]any{})` in a route pkg → guard flags it; remove → clean.
4. Literal widened §5b greps at HEAD: Part A = recorded-exempt health only; Part B survivors = allowlist only.
5. Full gate green: `go run ./tools/cilint ./...` exit 0 (after the testdb allowlist drive-by).

### Ordering
analyzer + tests (green) → register → real-repo + planted proof → docs (§5b, §8) → drive-by testdb allowlist → full gate green → evidence → commit.
