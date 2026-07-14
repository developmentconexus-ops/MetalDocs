# Phase 5 — Industry Comparison

> Gated by `.claude/skills/metaldocs-module-doc/references/industry-patterns-index.md`.

## Module character

`editor-ui-eigenpal` is a thin FE wrapper around an external vendored editor library. Most of the industry-pattern rows (RFC 9457 errors, idempotency, pagination, authz, multi-tenancy, migrations) are n/a for an adapter package with no HTTP / no DB.

## Admissible row applied

| Row | Applies how | Module file |
|---|---|---|
| IP-005 — OpenAPI as source-of-truth | n/a here (adapter has no API surface). Recorded for completeness; the adapter is *type-only* re-export of `Comment` to BE-typed `documents` module. |
| IP-006 — Forward-only migrations | n/a (no migrations). |

## Patterns NOT in index — recorded, not invented

1. **Vendored-fork pinning** (e.g. Renovate/Dependabot fork-pin practice; npm `file:` URI for tarballs) — **no row in the index** and no source URL pre-vetted. Phase 5 guard: do not cite without the user's approval to add a row. Flagged in tech-debt T-001 instead, with a note that the gap is a *supply-chain availability* issue rather than a pattern-comparison issue.
2. **Thin-wrapper adapter pattern** (Adapter / Anti-Corruption Layer per DDD) — not in index. The module's structure already enforces the seam (ADR 0001's "MetalDocs only documents the integration contract") so no industry citation needed to justify the design.

## Not-applicable rows (recorded so future readers see the skip is intentional)

- IP-001 (RFC 9457) — adapter surfaces no errors to API.
- IP-002 (Idempotency-Key) — no HTTP layer.
- IP-003 (Cursor pagination) — no list endpoint.
- IP-004 (Defense-in-depth authz) — no authz layer; trust boundary is the parent page (`isEditable` gate, server-side write authz in `documents` module).
- IP-007 (Observability correlation id) — no log emission inside adapter.
- IP-008 (Tenancy row-level) — no tables.

## Outcome

§5 in the published doc will explicitly state "no admissible patterns from the index apply to this adapter; the module's central concern is the seam contract, captured in §2 (Architecture Constraints) and ADR 0001 — not in any industry-pattern row." No fresh research added.
