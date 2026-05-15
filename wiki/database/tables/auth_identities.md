# metaldocs.auth_identities

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** auth

## Purpose
Current curated-baseline table owned by `auth`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `user_id` | `text` | no | Baseline column. |
| `username` | `text` | no | Baseline column. |
| `email` | `text` | yes | Baseline column. |
| `password_hash` | `text` | no | Baseline column. |
| `password_algo` | `text` | no | Baseline column. |
| `must_change_password` | `boolean` | no | Baseline column. |
| `last_login_at` | `timestamp with time zone` | yes | Baseline column. |
| `failed_login_attempts` | `integer` | no | Baseline column. |
| `locked_until` | `timestamp with time zone` | yes | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `updated_at` | `timestamp with time zone` | no | Baseline column. |
| `display_name` | `text` | no | Baseline column. |
| `is_active` | `boolean` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.auth_identities (
user_id text NOT NULL,
    username text NOT NULL,
    email text,
    password_hash text NOT NULL,
    password_algo text NOT NULL,
    must_change_password boolean DEFAULT false NOT NULL,
    last_login_at timestamp with time zone,
    failed_login_attempts integer DEFAULT 0 NOT NULL,
    locked_until timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    display_name text NOT NULL,
    is_active boolean DEFAULT true NOT NULL
);
```

## Runtime Usage

Use `rg -n "auth_identities" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
