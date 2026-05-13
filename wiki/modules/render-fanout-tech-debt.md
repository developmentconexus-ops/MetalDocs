# Tech Debt Register - render-fanout

> Companion to `wiki/modules/render-fanout.md`. Debt only; no fix prescriptions.

**Last verified:** 2026-05-13

## Items

### T-001 · Reconstruction/fanout behavior is spread across fanout and resolver packages without a consolidated module doc
- **Severity:** major
- **Surface:** `internal/modules/render/fanout/reconstruction.go:55`
- **Observation:** implementation is strong but documentation is still pipeline-stub level.
- **Evidence:** module page is currently a high-level stub with no flow/error matrix.
- **Linked backlog row:** `R-001`
- **Linked ADR:** missing-ADR

### T-002 · Outbox retry/terminal semantics are not documented as an explicit contract
- **Severity:** major
- **Surface:** `internal/modules/render/fanout/pdf_outbox_repository.go:80`
- **Observation:** retry progression, finalize behavior, and stale-claim reset exist in code but are not captured in module-level contract docs.
- **Evidence:** `MarkFailed` / `ResetStaleClaims` semantics.
- **Linked backlog row:** `R-002`
- **Linked ADR:** missing-ADR

### T-003 · Resolver registry compatibility/version policy lacks explicit ADR
- **Severity:** minor
- **Surface:** `internal/modules/render/resolvers/registry.go:13`
- **Observation:** resolver key/version behavior exists, but compatibility policy is convention-only.
- **Evidence:** registry and resolver contract tests.
- **Linked backlog row:** `R-003`
- **Linked ADR:** missing-ADR

## Coverage stats

- Public symbols undocumented: n/a (not fully audited)
- Operations missing C4 placement: n/a (stub-level doc)
- Cross-deps missing in section map: n/a (stub-level doc)
- State transitions missing: n/a (pipeline module)
- Decisions without ADR link: 3
