# Phase 4 Persistence Map — Audit Module

Module path: `internal/modules/audit`

## §1. Owned Tables

Owned table set: `metaldocs.audit_events`.

Evidence:
- `migrations/0004_init_audit_events.sql:1` — `CREATE TABLE IF NOT EXISTS metaldocs.audit_events (`
- No `public.audit_events` table creation found in `migrations/`.

Created in migration: `migrations/0004_init_audit_events.sql`.

| Table | Created in (migration filename) | Notes |
|---|---|---|
| `metaldocs.audit_events` | `migrations/0004_init_audit_events.sql` | `id` is `PRIMARY KEY`; no idempotency-key column distinct from `id`. |

| Column | Type | Constraints (NOT NULL, FK, default) |
|---|---|---|
| `id` | `TEXT` | `PRIMARY KEY` (`migrations/0004_init_audit_events.sql:2`) |
| `occurred_at` | `TIMESTAMPTZ` | `NOT NULL` (`migrations/0004_init_audit_events.sql:3`) |
| `actor_id` | `TEXT` | `NOT NULL` (`migrations/0004_init_audit_events.sql:4`) |
| `action` | `TEXT` | `NOT NULL` (`migrations/0004_init_audit_events.sql:5`) |
| `resource_type` | `TEXT` | `NOT NULL` (`migrations/0004_init_audit_events.sql:6`) |
| `resource_id` | `TEXT` | `NOT NULL` (`migrations/0004_init_audit_events.sql:7`) |
| `payload` | `JSONB` | `NOT NULL DEFAULT '{}'::jsonb` (`migrations/0004_init_audit_events.sql:8`) |
| `trace_id` | `TEXT` | `NOT NULL` (`migrations/0004_init_audit_events.sql:9`) |

Event id generation fact:
- `internal/modules/iam/delivery/http/admin_handler.go:458` sets event id with exact format string: `"20060102150405.000000000"` in:
  - `"evt_" + strings.ReplaceAll(time.Now().UTC().Format("20060102150405.000000000"), ".", "")`
- Fact: timestamp-formatted id generation is collision-prone at high concurrency.

Payload size constraint fact:
- No JSONB size limit found for `payload` in migration or audit module code.
- `migrations/0004_init_audit_events.sql:8` defines `payload JSONB NOT NULL DEFAULT '{}'::jsonb` only.
- Result: none — JSONB has no size constraint in `0004`.

## §2. Tables Read/Written but Not Owned

Audit module writes:
- `internal/modules/audit/infrastructure/postgres/writer.go:22-26` writes only `INSERT INTO metaldocs.audit_events`.

Audit module reads:
- `internal/modules/audit/infrastructure/postgres/writer.go:51-57` reads only `FROM metaldocs.audit_events`.

Inspection result across `internal/modules/audit/` and `migrations/`:
- Confirmed fact: audit module does not write any table other than `metaldocs.audit_events`.

| Table | Owner module | Read / Write | Operations using it |
|---|---|---|---|
| none | n/a | n/a | none |

## §3. Triggers / GUCs / Functions

Search scope: `migrations/` for trigger/function/RLS/policy/GUC entries touching `audit_events`.

Matches for `metaldocs.audit_events`:
- `migrations/0004_init_audit_events.sql:1` (table create)
- `migrations/0005_grant_workflow_audit_privileges.sql:2` (grant insert)

No trigger/function/RLS policy/GUC definitions touching `metaldocs.audit_events` found.

| Object | Kind (trigger / function / GUC) | File:line | Purpose |
|---|---|---|---|
| none | none | none | none |

## §4. Indexes

Indexes from `migrations/0004_init_audit_events.sql`:

| Index | Table | Columns | Unique? | Purpose |
|---|---|---|---|---|
| `idx_audit_events_occurred_at` | `metaldocs.audit_events` | `occurred_at DESC` | No | Time-ordered event listing (`migrations/0004_init_audit_events.sql:12`) |
| `idx_audit_events_actor_time` | `metaldocs.audit_events` | `actor_id, occurred_at DESC` | No | Actor + time filtering/order (`migrations/0004_init_audit_events.sql:13`) |
| `idx_audit_events_resource_time` | `metaldocs.audit_events` | `resource_type, resource_id, occurred_at DESC` | No | Resource + time filtering/order (`migrations/0004_init_audit_events.sql:14`) |

## §5. Tripwire Pairing Audit

Repo mutation methods in `internal/modules/audit/infrastructure/postgres/writer.go`:

| Method (file:line) | Authz.Require called? | Cap + area arg | SQL verb | Table |
|---|---|---|---|---|
| `Record` (`internal/modules/audit/infrastructure/postgres/writer.go:20`) | NO | n/a (side-effect sink, not capability-scoped) | INSERT | `metaldocs.audit_events` |

Footnote: side-effect sinks are out of the tripwire model; tripwire enforces caller-side capability; audit module records what already happened.

## §6. Migration History

Chronological migrations affecting audit persistence:

| Order | Filename | Verb summary | Date (from filename or commit) |
|---|---|---|---|
| 0004 | `0004_init_audit_events.sql` | `CREATE TABLE metaldocs.audit_events`; create 3 indexes | from filename order |
| 0005 | `0005_grant_workflow_audit_privileges.sql` | `GRANT INSERT ON TABLE metaldocs.audit_events TO metaldocs_app;` | from filename order |

Grant verification for `SELECT` on `metaldocs.audit_events`:
- `migrations/0005_grant_workflow_audit_privileges.sql:2` exact SQL: `GRANT INSERT ON TABLE metaldocs.audit_events TO metaldocs_app;`
- No explicit `GRANT SELECT ON TABLE metaldocs.audit_events TO metaldocs_app` found in `migrations/`.
- No `GRANT ... TO PUBLIC` for `metaldocs.audit_events` found in `migrations/`.
- No `ALTER DEFAULT PRIVILEGES ... GRANT ... ON TABLES` found in `migrations/` for this table.
- Source of `SELECT` for `metaldocs_app` via migration files: none found.

Other ALTERs on `audit_events` (columns/retention/hash chain):
- Search in `migrations/` for `ALTER TABLE` touching `audit_events`: none.
- Search in `migrations/` for retention/partition automation touching `audit_events` (`PARTITION`, `DROP PARTITION`, `pg_cron`, `retention`, `TTL`): none.
- Result: none — table grows monotonically.
