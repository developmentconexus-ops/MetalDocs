# F9.4 — widen noresponsemap to any map[string]<T> response literal

> **Milestone:** M9  ·  **Status:** approved (operator Option-A proceed, 2026-06-20) — code may begin.

## Consumer contract (the gate is the consumer)

The mission §8 H-D class = "no `map[string]any` **response literal** on a public route." The post-M8
re-audit proved the F8.6 `noresponsemap` analyzer (`tools/cilint`) is scoped to `map[string]any` /
`map[string]interface{}` only — its `isMapStringAnyLiteral` returns false for `map[string]string`. The 3
post-M8 Majors evaded the gate by emitting `map[string]string`. The class the gate must enforce is **any
untyped `map[string]<T>` reaching a 2xx body writer**, not one value type.

## What to implement

- In `tools/cilint` `noresponsemap`: replace the `map[string]any`-only key/value check with a check that
  flags **any** `map[string]<T>` composite literal (any value type `T`) reaching a 2xx body writer
  (`writeJSON` / `writeFillInJSON` / `WriteJSON`) on a registered-route package — including
  built-then-written locals. Keep the existing scope (`inRegisteredRoutePackage`), exemptions
  (`noResponseMapExemptFiles` — health.go), and `//cilint:allow-responsemap` suppression unchanged.
- Update `wiki/architecture/api-contract.md` §5b: the rule now reads "no `map[string]<T>` response
  literal" (state the widened type scope); the allowlist categories (recordAudit emit params,
  command-input maps, domain-mirror fields, declared-dynamic metrics leaves, health probes) are
  unchanged — they are non-response uses.

## Non-goals

- No change to the route-package scope or the recorded exemptions (those were settled in F8.6).
- No new analyzer; widen the existing one.

## Validation Gate

- Unit test in `tools/cilint/internal/analyzers`: a `map[string]string` literal written to a 2xx writer
  in a registered-route fixture **is flagged**; an allowlisted/exempt use is **not** flagged; a typed
  struct body is not flagged.
- `GOFLAGS=-mod=mod go run ./tools/cilint ./...` exits **0 at HEAD only after F9.1–F9.3 land** (and would
  have exited non-zero against the pre-F9.1 `map[string]string` sites — demonstrate on a fixture).
- §5b documents the widened scope.
- `go build ./...` + `go test ./tools/cilint/...` green.
