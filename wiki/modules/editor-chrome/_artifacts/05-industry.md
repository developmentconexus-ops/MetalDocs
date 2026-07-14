# Phase 5 — Industry Comparison

editor-chrome is a frontend React primitive — a slot-based layout component overlaying a custom toolbar atop a third-party editor (eigenpal). The pre-vetted patterns in `references/industry-patterns-index.md` are all backend / API-design concerns: RFC 9457 error envelope (IP-001), Stripe idempotency (IP-002), cursor pagination (IP-003), defense-in-depth authz (IP-004), OpenAPI-as-source (IP-005), forward-only migrations (IP-006), structured-log correlation (IP-007), row-level tenancy (IP-008).

None of these admissible patterns map onto a presentation-layer primitive that owns no routes, no SQL, no auth surface, and no observability sink.

## Pattern-by-pattern mapping

| Index id | Topic | Applies to editor-chrome? | Note |
|---|---|---|---|
| IP-001 | RFC 9457 errors | No | Module emits no HTTP responses. Error rendering belongs to consumer pages. |
| IP-002 | Idempotency | No | No write surface. |
| IP-003 | Cursor pagination | No | No list endpoint. |
| IP-004 | Defense-in-depth authz | No | Module is unprivileged UI primitive; no capability check, no tier-1/tier-2 split applies. |
| IP-005 | OpenAPI codegen | No | No HTTP route, no codegen target. |
| IP-006 | Forward-only migrations | No | No SQL. |
| IP-007 | Observability correlation | No | Not yet wired anywhere in MetalDocs FE; non-issue here. |
| IP-008 | Row-level tenancy | No | No data layer. |

## Industry patterns explicitly NOT added in this session

The user constraint for this run is: **"One module. Push back on additions."** Per the skill's industry-comparison guard, adding new index rows requires user consent. No new rows were proposed or added — the module is documented without frontend-primitive patterns (slot composition, compound components, tokenised design systems, etc.). Those can be added in a future module that genuinely needs them, alongside source-linked quotes and a per-module application anchor.

## What replaces §5 industry-comparison in the composed doc

The composed module doc records this as `§9: Architecture Decisions → n/a — industry patterns in index do not apply; presentation-layer module.` and `§11: Top-3 risks → derived from internal gap analysis, not industry comparison.` No "Stripe does X" / "Google does Y" prose is permitted; none is written.

## Citations used

None.
