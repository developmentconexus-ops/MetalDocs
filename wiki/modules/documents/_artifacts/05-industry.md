# Phase 5 — Industry comparison (`documents`)

**Date:** 2026-05-10
**Source of truth:** `.claude/skills/metaldocs-module-doc/references/industry-patterns-index.md`

Each row: pattern ID → applies-here verdict → MetalDocs file:line anchor → gap (if any).

## IP-001 · RFC 9457 Problem Details

- **Source:** https://www.rfc-editor.org/rfc/rfc9457.html (RFC 9457, 2023-07) — "A problem details object can be extended with additional members."
- **Applies here:** `documents` returns errors via `httpErr` + `mapErr` (`internal/modules/documents/delivery/http/handler.go:958–1009`, `:1013`) — legacy `{error:{code,message,details,trace_id}}` shape, **not** Problem+JSON.
- **Status:** mid-migration. Codegen bootstrap is in (ADR 0012) but handlers not migrated. Mirrors `iam` T-006.
- **Gap → debt:** T-001 (Major).

## IP-002 · Stripe-style `Idempotency-Key`

- **Source:** https://docs.stripe.com/api/idempotent_requests (accessed 2026-05-10) — "Keys are eligible to be removed from the system after they're at least 24 hours old."
- **Applies here, partially:**
  - `POST /api/v1/controlled-documents` and the revisions write route are header-idempotent via `internal/platform/idempotency` (ADR 0011).
  - `POST /api/v1/documents/{id}/finalize` is **not** header-idempotent. Submit-side computes an internal deterministic `ComputeIdempotencyKey` (`internal/modules/documents/approval/application/idempotency.go:20`, called at `submit_service.go:61`) but no `Idempotency-Key` header is read and `metaldocs.idempotency_keys` is not written.
  - Replay safety on finalize relies on `ux_approval_instances_active` partial unique index (`migrations/0135_*.sql:33`): second call fails with 409 rather than silently double-instancing.
- **Gap → debt:** T-006 (Major — surface for duplicate submit exists; index prevents double-write but client gets 409, not idempotent replay).

## IP-003 · Cursor pagination (Relay Connections)

- **Source:** https://relay.dev/graphql/connections.htm (Relay Connections, 2021) — "Pagination should be done with a forward-only cursor."
- **Applies here:** `GET /api/v1/documents` uses offset pagination — `parseListOptions` reads `page` + `pageSize` (`handler.go:200`), repo `LIMIT/OFFSET` at `repository.go:343`, cap `pageSize ≤ 50`.
- **Verdict:** not-applicable today — offset is acceptable for the current library scale (`pageSize` capped at 50, totals shown). Cursor pagination would be a forward-looking refactor, not current debt.
- **Gap:** none surfaced as debt; recorded as a future consideration only.

## IP-004 · Defense-in-depth (NIST SP 800-95 §4.3)

- **Source:** NIST SP 800-95 (2007) — "Multiple layers of access control reduce single-point bypass risk."
- **Applies here, asymmetrically:**
  - **Approval-instance writes** (full layered): tier-1 role gate at `handler.go:870`, tier-2 `authz.Require(ctx, tx, "doc.submit", areaCode)` at `submit_service.go:85`, Postgres tripwire trigger `trg_require_cap_asserted_instances` (`migrations/0142b_role_capabilities_v2_enforce.sql:201`) and `..._signoffs` (`:207`).
  - **`documents` table writes** (single-layer): `CreateDocumentTx`, `UpdateDocumentName`, `UpdateDocumentStatus`, `MarkArchived`, `Unarchive` (`repository/repository.go:73, :216, :428, :1071, :1082`) have only tier-1 role gate. No `authz.Require` and no `enforce_capability_asserted` trigger attached to `documents`.
- **Gap → debt:** T-003 (Major — documents table is the one regulated mutation surface without defense-in-depth).

## IP-005 · OpenAPI source-of-truth

- **Source:** https://learn.openapis.org/best-practices.html (OAI 3.0.3, 2020) — "The OpenAPI Specification … is the standard for HTTP APIs."
- **Applies here, weakly:**
  - `api.gen.go` exists for documents (`internal/modules/documents/api/api.gen.go`) but is **bootstrap only** (ADR 0012).
  - Routes registered via stdlib `mux.HandleFunc` directly, not via the generated `ServerInterface`.
  - `operationId` mismatches:
    - List handler is `listDocuments` (`handler.go:145`); spec exposes `listDocumentsV2` (`api/openapi/v1/openapi.yaml:3156`).
    - `finalizeDocument`: spec path exists at `openapi.yaml:3251` but **no `operationId` set**; generated stub `PostApiV2DocumentsIdFinalize` (`api.gen.go:1215`).
    - `renameDocument`: route registered (`handler.go:115` and duplicate `:86`) but **absent from spec entirely**.
    - `duplicateDocument` + comments CRUD: handlers exist, no spec ops (per `wiki/backlog/contract-first-followups.md`).
- **Gap → debt:** T-002 (Critical — contract drift on the regulated path).

## IP-006 · Forward-only migrations (Fowler)

- **Source:** https://martinfowler.com/articles/evodb.html (Fowler, 2016) — "Each change to the database is described by a migration script."
- **Applies here:** all `documents`-owned migrations are append-only (`migrations/0001…0183`). Rename from `documents_v2` → `documents` was done as a paired forward migration (`migrations/0167_documents_v2_rename_to_documents.sql` + `0168_drop_documents_legacy.sql`), not by editing prior files. `documents.locked_at` removed in `0181_drop_locked_at.sql`. `documents.name` non-empty enforced by `0183_documents_name_not_empty.sql`.
- **Gap:** none.

## IP-007 · Observability (correlation id)

- **Source:** index row IP-007 (index marks observability "not yet wired in MetalDocs").
- **Applies here:** documents module emits no structured trace correlation; `OUT-deps` artifact records no `pkg/observability` or `pkg/tracing` import.
- **Verdict:** not-applicable in this audit — module-level observability is system-wide future work, not documents-specific debt.

## IP-008 · Row-level tenant_id (Crunchy multi-tenant)

- **Source:** https://www.crunchydata.com/blog/designing-your-postgres-database-for-multi-tenancy (accessed 2026-05-10) — "Add tenant_id to every multi-tenant table and index it first."
- **Applies here:** verified in `_artifacts/04-persistence.md`. Every owned table carries `tenant_id`:
  - `documents` (`migrations/0110_*.sql:14`), `editor_sessions` (`:34`), `document_revisions` (`:49`), `document_checkpoints` (`:90`), `document_placeholder_values` (`migrations/0152_*.sql:51`), `document_exports` (`migrations/0111_*.sql:3`), `document_comments` (`migrations/0118_*.sql:1`), `approval_routes` (`migrations/0134_*.sql:3`), `approval_route_stages` (`:14`), `approval_instances` (`migrations/0135_*.sql:9`), `approval_stage_instances` (`:40`), `approval_signoffs` (`:72`), `governance_events` (`migrations/0125_*.sql:24`), `metaldocs.pdf_dispatch_outbox` (`migrations/0176_*.sql:2`).
  - Repo queries scope by `tenant_id = $1` consistently (`repository.go:343, :376`).
- **Schema bug surfaced during this audit:** `document_placeholder_values.revision_id REFERENCES documents(id)` (`migrations/0152_*.sql:51`) — should reference `document_revisions(id)`. Captured as T-009.
- **Gap:** none for the tenant-id pattern itself; the placeholder-values FK target is a separate schema bug.

---

## Coverage summary

| Pattern | Verdict | Debt link |
|---|---|---|
| IP-001 RFC 9457 | mid-migration | T-001 |
| IP-002 Idempotency | partial (finalize gap) | T-006 |
| IP-003 Cursor pagination | not-applicable today | — |
| IP-004 Defense-in-depth | asymmetric (documents table uncovered) | T-003 |
| IP-005 OpenAPI source-of-truth | drift (spec ↔ handlers) | T-002 |
| IP-006 Forward-only migrations | satisfied | — |
| IP-007 Observability | not-applicable (system-wide) | — |
| IP-008 Multi-tenant `tenant_id` | satisfied | — (T-009 is a different bug) |

No new patterns were introduced. All citations trace to existing rows in `industry-patterns-index.md`.
