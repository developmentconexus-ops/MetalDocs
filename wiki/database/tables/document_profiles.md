# metaldocs.document_profiles

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** taxonomy
> **Last verified:** 2026-06-12 (migration 0236 — `is_active` dropped as dead schema, superseded by `archived_at`; baseline file unchanged, the column is removed by the forward migration)

## Purpose
Current curated-baseline table owned by `taxonomy`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `code` | `text` | no | Baseline column. |
| `family_code` | `text` | no | Baseline column. |
| `name` | `text` | no | Baseline column. |
| `description` | `text` | no | Baseline column. |
| `review_interval_days` | `integer` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `alias` | `text` | no | Baseline column. |
| `tenant_id` | `uuid` | no | Baseline column. |
| `default_template_version_id` | `uuid` | yes | Baseline column. |
| `owner_user_id` | `text` | yes | Baseline column. |
| `editable_by_role` | `text` | no | Baseline column. |
| `archived_at` | `timestamp with time zone` | yes | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.document_profiles (
code text NOT NULL,
    family_code text NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    review_interval_days integer NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    alias text NOT NULL,
    tenant_id uuid DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid NOT NULL,
    default_template_version_id uuid,
    owner_user_id text,
    editable_by_role text DEFAULT 'admin'::text NOT NULL,
    archived_at timestamp with time zone,
    CONSTRAINT chk_document_profiles_alias_length CHECK (((char_length(alias) >= 1) AND (char_length(alias) <= 24))),
    CONSTRAINT document_profiles_review_interval_days_check CHECK ((review_interval_days > 0)),
    CONSTRAINT profile_code_format CHECK ((code ~ '^[a-z][a-z0-9_-]{1,63}$'::text))
);
```

## Runtime Usage

Use `rg -n "document_profiles" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.

**PK left as-is (2026-07-02, DB-08/T-015 audit):** PK remains `code` alone (cross-tenant unique), NOT promoted to `(tenant_id, code)` despite the redundant-looking `(tenant_id, code)` unique index (`ux_document_profiles_tenant_code`). Reason: `document_profiles(code)` has 4 inbound single-column FKs (`document_profile_governance`, `document_profile_schema_versions`, `document_sequences`, and formerly `template_drafts`) from tables that have **no `tenant_id` column and no RLS** — they are global registries keyed purely by `profile_code`, not per-tenant data. Promoting the PK would break those FKs and require a much larger redesign (adding `tenant_id` to 4 more tables, backfilling, adding RLS, deciding whether profile governance/schema-versions/sequence-counters become per-tenant concepts). See `wiki/modules/taxonomy-tech-debt.md` T-015 design note for the full inventory and recommendation. Contrast with `document_process_areas` (same T-015 finding), which WAS promoted — see `wiki/database/tables/document_process_areas.md`.
