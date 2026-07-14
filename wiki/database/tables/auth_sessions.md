# metaldocs.auth_sessions

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** auth

## Purpose
Current curated-baseline table owned by `auth`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `session_id` | `text` | no | Baseline column. |
| `user_id` | `text` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `expires_at` | `timestamp with time zone` | no | Baseline column. |
| `revoked_at` | `timestamp with time zone` | yes | Baseline column. |
| `ip_address` | `text` | yes | Baseline column. |
| `user_agent` | `text` | yes | Baseline column. |
| `last_seen_at` | `timestamp with time zone` | no | Baseline column. |
| `tenant_id` | `uuid` | no | Session-bound tenant selected at login; FK to `metaldocs.tenants(id)`. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.auth_sessions (
session_id text NOT NULL,
    user_id text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    revoked_at timestamp with time zone,
    ip_address text,
    user_agent text,
    last_seen_at timestamp with time zone DEFAULT now() NOT NULL,
    tenant_id uuid NOT NULL
);
```

## Runtime Usage

Use `rg -n "auth_sessions" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.

- `tenant_id` is the authoritative active-tenant binding for the authenticated session.
- Since migration `0214_tenants_master_table.sql`, the session tenant must resolve through `metaldocs.tenants` so auth responses can surface the human-readable tenant name.
