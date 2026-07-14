# metaldocs.tenants

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** auth/platform

## Purpose
Canonical tenant master table. Stores the human-readable company/tenant identity that backs UUID tenant ownership across auth and IAM.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Canonical tenant identifier used by tenant-scoped operational tables. |
| `name` | `text` | no | Human-readable tenant/company name. |
| `slug` | `text` | no | Stable tenant slug for admin/UI use. |
| `created_at` | `timestamp with time zone` | no | Row creation timestamp. |
| `updated_at` | `timestamp with time zone` | no | Last update timestamp. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.tenants (
    id uuid NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT tenants_name_not_blank CHECK ((length(btrim(name)) > 0)),
    CONSTRAINT tenants_slug_not_blank CHECK ((length(btrim(slug)) > 0))
);
```

## Constraints

- Primary key: `tenants_pkey (id)`
- Unique: `tenants_slug_key (slug)`

## Runtime Usage

Use `rg -n "GetTenantByID|metaldocs\\.tenants|tenantName" internal frontend/apps/web/src` and the auth module wiki to verify readers before changing this table.

## Seed or Reference Data

`db/reference-data/0001_product_reference_data.sql` seeds the canonical system tenant row:

- `id = ffffffff-ffff-ffff-ffff-ffffffffffff`
- `name = System Tenant`
- `slug = system`

## Notes and Debt

- Existing runtime tables still carry their own `tenant_id` columns; this table is the canonical identity source, not yet a universal foreign-key backstop for every tenant-scoped table.
- Upgrade migration `0214_tenants_master_table.sql` backfills tenant rows from auth/IAM tenant UUIDs. Non-system rows without an existing name source are initialized with deterministic placeholder names until a real tenant-admin workflow exists.
