# Audit module — Phase 0 context

> **Date:** 2026-05-10
> **Composer:** main agent
> **Skill version:** metaldocs-module-doc 1.2

## What this module is (one paragraph)

`internal/modules/audit/` is a **write-only sink + read-only query surface** for regulated mutation events. Consumer modules call `auditdomain.Writer.Record(ctx, Event)` after performing a regulated state change; an admin/IAM console reads recent rows via `auditdomain.Reader.ListEvents`. Storage is a single Postgres table `metaldocs.audit_events` (migration 0004), append-only by design — no UPDATE/DELETE privilege granted (`migrations/0005_grant_workflow_audit_privileges.sql:2` grants INSERT only). Module exposes one HTTP route (`GET /api/v1/audit/events`) wired in `internal/modules/audit/delivery/http/handler.go:35`.

## Existing wiki coverage (before this session)

- `wiki/modules/audit.md` — **does not exist** (no prior module doc; this run creates the first).
- No ADR mentions audit explicitly.
- IN-bound references from other module docs:
  - `wiki/modules/auth-tech-debt.md` — T-002 audit-trail gap **Critical** (auth login/logout/password ops do not emit audit events).
  - `wiki/modules/documents-tech-debt.md` — T-005 rename audit outside tx **Major**; T-007 audit port latent **Minor**.
  - `wiki/modules/iam-tech-debt.md` — T-005 `handleUserRoleUpsert` missing audit emission (per user prompt cross-ref).
  - `wiki/backlog/iam-refactor.md`, `wiki/backlog/auth-refactor.md`, `wiki/backlog/documents-refactor.md` carry the matching refactor rows.

Implication for §11 / tech-debt scoping: most "audit gap" debt is **consumer-side** (auth, documents, iam) and lives in their registers. Audit module's own register should focus on intrinsic gaps: tampering surface, retention, immutability proofs, envelope conformance, query authorization, payload schema.

## Module skeleton (file enumeration)

```
internal/modules/audit/
├── domain/port.go              Event struct, Writer + Reader interfaces, ListEventsQuery
├── application/service.go      Service{reader}; ListEvents normalize + clamp [1..200] default 50
├── delivery/http/handler.go    Handler{service}; GET /api/v1/audit/events; custom error envelope
└── infrastructure/
    ├── memory/writer.go        in-process Writer+Reader (dev/tests)
    └── postgres/writer.go      Writer+Reader, INSERT + SELECT on metaldocs.audit_events
```

Test file: `tests/unit/audit_http_handler_test.go`.

## Persistence surface (preview)

- Table `metaldocs.audit_events` (migration 0004): `id TEXT PK, occurred_at TIMESTAMPTZ, actor_id TEXT, action TEXT, resource_type TEXT, resource_id TEXT, payload JSONB, trace_id TEXT`. Indexes on `occurred_at DESC`, `(actor_id, occurred_at DESC)`, `(resource_type, resource_id, occurred_at DESC)`.
- Grants (migration 0005): `INSERT` only to `metaldocs_app`. No UPDATE/DELETE grant — append-only by privilege.
- No FK to other tables. No trigger. No retention policy.

## Wiring (Dependencies struct)

`internal/platform/bootstrap/api.go:37-38, 100-101, 129-130` exposes `AuditWriter auditdomain.Writer` + `AuditReader auditdomain.Reader` in the platform `Dependencies` aggregate. Postgres path uses `auditpg.NewWriter(db)` for both writer and reader (same type implements both interfaces). Memory path uses `auditmemory.NewWriter()`. Consumer modules receive the `Writer` through their handler constructors (e.g. `iam.NewAdminHandler(..., auditWriter)`).

## Top operations (Phase 2 picks)

Phase 2 should trace 3 operations to cover full surface:

1. **Writer.Record** (write path) — pick a representative caller: `iam.AdminHandler.recordAudit` at `internal/modules/iam/delivery/http/admin_handler.go:449-466` → `auditpg.Writer.Record` → INSERT.
2. **handleEvents** (query path, HTTP) — `GET /api/v1/audit/events?resourceType=&resourceId=&limit=` → `application.Service.ListEvents` → `auditpg.Writer.ListEvents` (table SELECT).
3. **IAM admin recent-events read** (internal Reader consumer) — `iam.AdminHandler` `WithAuditReader` path at `internal/modules/iam/delivery/http/admin_handler.go:77-135` (recent 25 events for admin console).

No state-transition operation — module is append-only. **§6 (Runtime View) state-machine table row: "n/a — append-only sink, no aggregate lifecycle."**

## Open questions (none blocking)

- No retention policy in place. Worth recording as a tech-debt row (Major: regulatory exposure but currently latent).
- No HTTP authorization on `GET /api/v1/audit/events` — needs Phase 2 trace to confirm whether middleware enforces it (the handler itself has no `authz.Require` call).
- No tamper-proofing (hash chain / signing). Worth recording — common audit-table industry baseline (HashChain / Merkle, see Postgres `pgaudit`, AWS CloudTrail integrity validation).

Proceeding without user input; all questions can be resolved during Phase 1–5 research.

## Phase 6 §6 (Runtime View) note

State-machine row: **n/a — append-only sink**. No domain aggregate, no transitions. §6 will instead document the two sequence diagrams (Record write path, ListEvents query path).
