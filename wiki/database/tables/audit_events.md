# metaldocs.audit_events

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** audit

## Purpose
Current curated-baseline table owned by `audit`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `text` | no | Baseline column. |
| `occurred_at` | `timestamp with time zone` | no | Baseline column. |
| `actor_id` | `text` | no | Baseline column. |
| `action` | `text` | no | Baseline column. |
| `resource_type` | `text` | no | Baseline column. |
| `resource_id` | `text` | no | Baseline column. |
| `payload` | `jsonb` | no | Constrained by `audit_events_payload_size_cap CHECK (octet_length(payload::text) <= 65536)` (64 KiB), added by migration `0266` (2026-07-02, DB-09/T-010). |
| `trace_id` | `text` | no | Baseline column. |
| `tenant_id` | `text` | no | Baseline column. |
| `audit_sequence` | `bigint` | no | Baseline column. |
| `prev_hash` | `text` | no | Baseline column. |
| `row_hash` | `text` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.audit_events (
id text NOT NULL,
    occurred_at timestamp with time zone NOT NULL,
    actor_id text NOT NULL,
    action text NOT NULL,
    resource_type text NOT NULL,
    resource_id text NOT NULL,
    payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    trace_id text NOT NULL,
    tenant_id text DEFAULT ''::text NOT NULL,
    audit_sequence bigint NOT NULL,
    prev_hash text DEFAULT ''::text NOT NULL,
    row_hash text DEFAULT ''::text NOT NULL,
    CONSTRAINT audit_events_payload_size_cap CHECK ((octet_length((payload)::text) <= 65536))
);
```

**Post-baseline note (migration `0266`, 2026-07-02, DB-09):** `audit_events_payload_size_cap` added (64 KiB ceiling on `octet_length(payload::text)`), added `NOT VALID` + `VALIDATE CONSTRAINT` for low-downtime rollout, pre-checked by a RAISE-on-violation DO block. Same migration also durably hardens (does not change) the application role's grant posture: `metaldocs_app` has only ever had `INSERT, SELECT` on this table (never `UPDATE`/`DELETE`) — 0266 adds an explicit, idempotent `REVOKE UPDATE, DELETE, TRUNCATE ... FROM metaldocs_app` so that posture is self-documenting and survives an accidental future blanket-GRANT. See `wiki/modules/audit-tech-debt.md` T-013 for the full role inventory and the deferred items (partitioning, `pg_cron` retention, ULID/UUIDv7 ids).

## Runtime Usage

Use `rg -n "audit_events" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Grants

Application role `metaldocs_app` (the single connection role used by all 4 binaries — no distinct audit-writer role exists): `INSERT, SELECT` only (`archive/migrations/0005_grant_workflow_audit_privileges.sql`, widened by `archive/migrations/0193_audit_events_hash_chain.sql`). `UPDATE`/`DELETE`/`TRUNCATE` explicitly revoked (never granted; hardened as a no-op-but-durable statement by migration `0266`). This is the append-only enforcement mechanism referenced by the (closed) tamper-evidence hardening — see `wiki/modules/audit-tech-debt.md` T-004.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema. See `wiki/modules/audit-tech-debt.md` for the tech-debt register (T-010 payload cap closed 2026-07-02; T-013 documents the DB-09 hardening pass and deferred items).
