# metaldocs.document_access_policies

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** documents

## Purpose
Current curated-baseline table owned by `documents`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `bigint` | no | Baseline column. |
| `subject_type` | `text` | no | Baseline column. |
| `subject_id` | `text` | no | Baseline column. |
| `resource_scope` | `text` | no | Baseline column. |
| `resource_id` | `text` | no | Baseline column. |
| `capability` | `text` | no | Baseline column. |
| `effect` | `text` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.document_access_policies (
id bigint NOT NULL,
    subject_type text NOT NULL,
    subject_id text NOT NULL,
    resource_scope text NOT NULL,
    resource_id text NOT NULL,
    capability text NOT NULL,
    effect text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT chk_document_access_policies_capability CHECK ((capability = ANY (ARRAY['document.create'::text, 'document.view'::text, 'document.edit'::text, 'document.upload_attachment'::text, 'document.change_workflow'::text, 'document.manage_permissions'::text]))),
    CONSTRAINT chk_document_access_policies_effect CHECK ((effect = ANY (ARRAY['allow'::text, 'deny'::text]))),
    CONSTRAINT chk_document_access_policies_resource_scope CHECK ((resource_scope = ANY (ARRAY['document'::text, 'document_type'::text, 'area'::text]))),
    CONSTRAINT chk_document_access_policies_subject_type CHECK ((subject_type = ANY (ARRAY['user'::text, 'role'::text, 'group'::text])))
);
```

## Runtime Usage

Use `rg -n "document_access_policies" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.
