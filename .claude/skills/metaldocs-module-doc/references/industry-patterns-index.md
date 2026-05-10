# Industry Patterns Index

Curated, source-linked patterns admissible in §5 (Industry comparison) of any module doc. Use this index BEFORE doing fresh web research.

## How to use

1. In Phase 5 (Industry comparison), pick patterns from this index by `id`.
2. Cite as: `[<id>] — <source URL> (accessed YYYY-MM-DD) — "<quote>"`
3. Tie each citation to a MetalDocs file:line that the pattern bears on.
4. If you need a pattern not listed here, ask the user once. On `yes`, add a row in this file in the same commit as the doc. No silent additions.

## Schema

| Field | Notes |
|---|---|
| `id` | `IP-NNN` |
| `topic` | short label (authz, errors, pagination, idempotency, observability, migrations, tenancy, …) |
| `pattern` | one-line summary |
| `source` | canonical URL (RFC, vendor docs, paper) |
| `version` | version / RFC number / accessed date |
| `quote` | one short verbatim sentence (≤ 30 words) |
| `applies_to` | which MetalDocs file:line(s) this can be checked against |
| `notes` | caveats (e.g. "Stripe's model — not blessed for ALL APIs") |

## Rows

| id | topic | pattern | source | version | quote | applies_to | notes |
|---|---|---|---|---|---|---|---|
| IP-001 | errors | RFC 9457 Problem Details JSON envelope | https://www.rfc-editor.org/rfc/rfc9457.html | RFC 9457 (2023-07) | "A problem details object can be extended with additional members." | `api/openapi/v1/openapi.yaml` (Problem schema) | obsoletes RFC 7807 |
| IP-002 | idempotency | Stripe-style `Idempotency-Key` header, body-hash check, replay returns original response | https://docs.stripe.com/api/idempotent_requests | accessed 2026-05-10 | "Keys are eligible to be removed from the system after they're at least 24 hours old." | `internal/platform/idempotency/` | scope: 24h replay window |
| IP-003 | pagination | Cursor pagination over offset for stable scroll | https://relay.dev/graphql/connections.htm | Relay Connections (2021) | "Pagination should be done with a forward-only cursor." | list ops in `api/openapi/v1/openapi.yaml` | adapt: opaque base64 cursor, sort+filter hash |
| IP-004 | authz | Defense-in-depth: edge check + in-tx check + DB constraint | NIST SP 800-95 §4.3 | NIST SP 800-95 (2007) | "Multiple layers of access control reduce single-point bypass risk." | `wiki/decisions/0007-two-tier-authz.md` | maps to tier-1 CapabilityService + tier-2 `authz.Require` + Postgres tripwire |
| IP-005 | api-design | OpenAPI as source-of-truth, codegen for server stubs and clients | https://learn.openapis.org/best-practices.html | OAI 3.0.3 (2020) | "The OpenAPI Specification … is the standard for HTTP APIs." | `api/openapi/v1/openapi.yaml`, `**/*.gen.go` | `oapi-codegen` v2.7.0 |
| IP-006 | migrations | Forward-only, append-only migration files, no edits to merged migrations | https://martinfowler.com/articles/evodb.html | Fowler (2016) | "Each change to the database is described by a migration script." | `db/migrations/`, `internal/modules/*/migrations/` | applies to all module migration sets |
| IP-007 | observability | Structured logs with request-scoped correlation id | https://www.rfc-editor.org/rfc/rfc7234 + ad-hoc | accessed 2026-05-10 | "Correlate spans across boundaries with a single id." | <fill when observability lands> | observability not yet wired in MetalDocs — flag as missing-ADR if a module assumes it |
| IP-008 | tenancy | Row-level tenant id + scoped indexes for multi-tenant Postgres | https://www.crunchydata.com/blog/designing-your-postgres-database-for-multi-tenancy | accessed 2026-05-10 | "Add tenant_id to every multi-tenant table and index it first." | every owned table in persistence map | check during Phase 4 audit |

## Rules

- One pattern per row.
- `quote` MUST be verbatim and ≤ 30 words. No paraphrase.
- `source` MUST be a stable URL (RFC, primary vendor doc, paper). Blog posts only if the author is the canonical source for the topic.
- When a pattern is rejected during Phase 5 (e.g. doesn't apply), DO NOT remove the row — leave it and note `not-applicable: <module> — <reason>` in the doc itself.
- Patterns added during a session: append at the bottom with the same commit that publishes the module doc.

## Forbidden

- Citing "Big Tech does X" without a source URL.
- Citing a blog from 2015 about a library at version 0.x.
- Citing the same source >2 times in one §5 section (means you are reaching).
