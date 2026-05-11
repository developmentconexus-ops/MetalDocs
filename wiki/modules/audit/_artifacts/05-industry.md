# Audit — Phase 5 industry comparison

> Composer: main agent (Opus 4.7). Date: 2026-05-10. Patterns admissible per `.claude/skills/metaldocs-module-doc/references/industry-patterns-index.md`.

## Picks

### IP-001 — RFC 9457 Problem Details
- **Cite:** https://www.rfc-editor.org/rfc/rfc9457.html (RFC 9457, 2023-07) — "A problem details object can be extended with additional members."
- **Applies to:** `internal/modules/audit/delivery/http/handler.go:97-105` (error envelope) and the success body at `:82-84` (`{"items":[...]}` — not RFC 9457 per se, but the error envelope is the contract drift).
- **Finding:** audit handler emits `{"error":{"code","message","details","trace_id"}}` — the same legacy envelope flagged in iam T-006, auth T-003, documents T-001. Status: drift. → tech-debt T-002 (Major).

### IP-004 — NIST SP 800-95 defense-in-depth
- **Cite:** NIST SP 800-95 (2007) §4.3 — "Multiple layers of access control reduce single-point bypass risk."
- **Applies to:** `internal/modules/audit/delivery/http/handler.go:34-35` (route registration) and `apps/api/cmd/metaldocs-api/permissions.go:211-221` (permission resolver default — no rule for `/api/v1/audit/events`).
- **Finding:** `GET /api/v1/audit/events` is reachable without authn or capability check (verified via grep — Phase 2 `02-flow-list.md`). Zero authz layers; defense-in-depth count = 0 on a regulated read surface. → tech-debt T-001 (Critical, authn/authz bypass trigger).

### IP-006 — forward-only migrations (Fowler 2016)
- **Cite:** https://martinfowler.com/articles/evodb.html — "Each change to the database is described by a migration script."
- **Applies to:** `migrations/0004_init_audit_events.sql`, `migrations/0005_grant_workflow_audit_privileges.sql`.
- **Finding:** Compliant. Migrations are forward-only, no ALTER/retention/partitioning since 0005. (Becomes a debt only paired with the retention gap — see T-003.)

### IP-008 — row-level tenant id + scoped indexes (Crunchy 2024)
- **Cite:** https://www.crunchydata.com/blog/designing-your-postgres-database-for-multi-tenancy (accessed 2026-05-10) — "Add tenant_id to every multi-tenant table and index it first."
- **Applies to:** `migrations/0004_init_audit_events.sql:1-14` (table + indexes).
- **Finding:** `metaldocs.audit_events` has no `tenant_id` column and no tenant-scoped index. Cross-module pattern (auth T-008 is the same gap on `auth_*` tables; both latent because MetalDocs is single-tenant today). On audit specifically: ListEvents has no tenant filter — when multi-tenant lands, the existing index `(resource_type, resource_id, occurred_at DESC)` is the only narrowing path. → tech-debt T-007 (Major, latent).

## Not applicable to this module

- **IP-002 (idempotency / Stripe key):** audit Writer is a fire-and-forget single-INSERT side-effect; idempotency-key replay window does not apply.
- **IP-003 (cursor pagination):** ListEvents supports only `limit` clamp; pagination is genuinely absent, but for a 50-row admin surface it is acceptable today. Filed as minor only if growth pressures appear.
- **IP-005 (OpenAPI as source-of-truth + codegen):** `/audit/events` is mounted via `http.ServeMux.HandleFunc` directly, NOT via oapi-codegen. The OpenAPI spec has the path but no `operationId` (Phase 2 found `(unclear: no operationId field under /audit/events get)`). → tech-debt T-008 (Minor).
- **IP-007 (observability / correlation id):** audit module records `trace_id` per event (`port.go:16`) but logs nothing on failure — orthogonal to the IP-007 baseline, not a gap.

## Patterns deliberately NOT cited (no index row, not added in this session)

- **Append-only audit-trail with tamper-evidence (hash-chain / WORM / digital signing):** common baseline (AWS CloudTrail log file integrity validation, Google Cloud audit logs, Postgres `pgaudit`). MetalDocs's audit table is append-only **by privilege only** (`migration 0005:2` grants INSERT, no UPDATE/DELETE — but `metaldocs` schema owner and Postgres superuser retain mutation rights). No hash chain, no signature, no Merkle root, no external WORM mirror. Recorded as tech-debt T-004 (Critical per user-supplied rubric: "audit-trail tampering path = Critical"). Filed without industry citation; if the user opts in to adding a new IP-NNN row, do it in the same commit. Avoiding "industry-standard says X without source" anti-pattern.

- **Retention policy (Postgres partitioning + `pg_cron` / declarative TTL):** audit_events grows monotonically. Regulated data has retention floor + ceiling obligations (ISO 9001 §7.5.3 record-retention; LGPD/GDPR right-to-erasure for personal data in payloads). No row for "retention" in the index today. Recorded as tech-debt T-003 (Major) without citation, same rationale as above.

## Summary

| IP | Outcome | Tech-debt row |
|---|---|---|
| IP-001 | drift | T-002 (Major) |
| IP-004 | bypass | T-001 (Critical) |
| IP-006 | compliant | — |
| IP-008 | latent gap | T-007 (Major) |

Two missing-from-index concerns (tamper-evidence, retention) filed without industry citation to keep §5 honest.
