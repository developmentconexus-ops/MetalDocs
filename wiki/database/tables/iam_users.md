# metaldocs.iam_users

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** iam

## Purpose
Current curated-baseline table owned by `iam`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `user_id` | `text` | no | Baseline column. |
| `display_name` | `text` | no | Baseline column. |
| `is_active` | `boolean` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `updated_at` | `timestamp with time zone` | no | Baseline column. |
| `tenant_id` | `uuid` | no | Baseline column. |
| `deactivated_at` | `timestamp with time zone` | yes | Baseline column. |
| `last_login_ip` | `text` | yes | Client IP on most recent successful login (PR-4, migration 0219). |
| `last_login_user_agent` | `text` | yes | Raw User-Agent on most recent successful login (PR-4, migration 0219). |
| `last_login_device_label` | `text` | yes | Derived device label (browser + OS). Populated by PR-7 once UA parsing lands; nullable until then. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.iam_users (
user_id text NOT NULL,
    display_name text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    tenant_id uuid DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid NOT NULL,
    deactivated_at timestamp with time zone,
    CONSTRAINT iam_users_deactivated_after_created CHECK (((deactivated_at IS NULL) OR (deactivated_at >= created_at)))
);
```

## Runtime Usage

Use `rg -n "iam_users" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.

### Migrations applied since baseline
- `0219_iam_users_last_login_context.sql` (PR-4) — adds `last_login_ip`, `last_login_user_agent`, `last_login_device_label` so the People-tab "Last login" drawer can surface IP/UA/device alongside the timestamp on `auth_identities`. Forward-only, idempotent (`ADD COLUMN IF NOT EXISTS`).

_Last verified: 2026-06-02 (PR-4)._
