# documents

This dictionary page covers same-name tables in multiple schemas. Keep schema qualification explicit in runtime SQL.

## metaldocs.documents

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `metaldocs`
> **Owner:** documents

## Purpose
Current curated-baseline table owned by `documents`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `text` | no | Baseline column. |
| `title` | `text` | no | Baseline column. |
| `owner_id` | `text` | no | Baseline column. |
| `classification` | `text` | no | Baseline column. |
| `status` | `text` | no | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `updated_at` | `timestamp with time zone` | no | Baseline column. |
| `document_type_code` | `text` | no | Baseline column. |
| `business_unit` | `text` | no | Baseline column. |
| `department` | `text` | no | Baseline column. |
| `tags` | `jsonb` | no | Baseline column. |
| `effective_at` | `timestamp with time zone` | yes | Baseline column. |
| `expiry_at` | `timestamp with time zone` | yes | Baseline column. |
| `metadata_json` | `jsonb` | no | Baseline column. |
| `document_profile_code` | `text` | no | Baseline column. |
| `document_family_code` | `text` | no | Baseline column. |
| `process_area_code` | `text` | yes | Baseline column. |
| `subject_code` | `text` | yes | Baseline column. |
| `profile_schema_version` | `integer` | no | Baseline column. |
| `document_sequence` | `integer` | no | Baseline column. |
| `document_code` | `text` | no | Baseline column. |
| `document_type_key` | `text` | no | Baseline column. |
| `document_type_version` | `integer` | no | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE metaldocs.documents (
id text NOT NULL,
    title text NOT NULL,
    owner_id text NOT NULL,
    classification text NOT NULL,
    status text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    document_type_code text NOT NULL,
    business_unit text NOT NULL,
    department text NOT NULL,
    tags jsonb DEFAULT '[]'::jsonb NOT NULL,
    effective_at timestamp with time zone,
    expiry_at timestamp with time zone,
    metadata_json jsonb DEFAULT '{}'::jsonb NOT NULL,
    document_profile_code text NOT NULL,
    document_family_code text NOT NULL,
    process_area_code text,
    subject_code text,
    profile_schema_version integer DEFAULT 1 NOT NULL,
    document_sequence integer NOT NULL,
    document_code text NOT NULL,
    document_type_key text DEFAULT ''::text NOT NULL,
    document_type_version integer DEFAULT 1 NOT NULL
);
```

## Runtime Usage

Use `rg -n "documents" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Curated baseline table in the `metaldocs` schema.

## public.documents

> **Source:** `db/baseline/0001_current_schema.sql`
> **Schema:** `public`
> **Owner:** documents

## Purpose
Current curated-baseline table owned by `documents`. See the owning module wiki and runtime repositories for business behavior.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|
| `id` | `uuid` | no | Baseline column. |
| `tenant_id` | `uuid` | no | Baseline column. |
| `template_version_id` | `uuid` | no | Baseline column. |
| `name` | `text` | no | Baseline column. |
| `status` | `text` | no | Baseline column. |
| `form_data_json` | `jsonb` | no | Baseline column. |
| `current_revision_id` | `uuid` | yes | Baseline column. |
| `active_session_id` | `uuid` | yes | Baseline column. |
| `finalized_at` | `timestamp with time zone` | yes | Baseline column. |
| `archived_at` | `timestamp with time zone` | yes | Baseline column. |
| `created_at` | `timestamp with time zone` | no | Baseline column. |
| `updated_at` | `timestamp with time zone` | no | Baseline column. |
| `created_by` | `text` | no | Baseline column. |
| `effective_from` | `timestamp with time zone` | yes | Baseline column. |
| `effective_to` | `timestamp with time zone` | yes | Baseline column. |
| `revision_number` | `integer` | no | Baseline column. |
| `revision_version` | `integer` | no | Baseline column. |
| `content_hash_at_submit` | `text` | yes | Baseline column. |
| `placeholder_schema_snapshot` | `jsonb` | yes | Baseline column. |
| `placeholder_schema_hash` | `bytea` | yes | Baseline column. |
| `composition_config_snapshot` | `jsonb` | yes | Baseline column. |
| `composition_config_hash` | `bytea` | yes | Baseline column. |
| `body_docx_snapshot_s3_key` | `text` | yes | Baseline column. |
| `body_docx_hash` | `bytea` | yes | Baseline column. |
| `values_frozen_at` | `timestamp with time zone` | yes | Baseline column. |
| `values_hash` | `bytea` | yes | Baseline column. |
| `final_docx_s3_key` | `text` | yes | Baseline column. |
| `content_hash` | `bytea` | yes | Baseline column. |
| `final_pdf_s3_key` | `text` | yes | Baseline column. |
| `pdf_hash` | `bytea` | yes | Baseline column. |
| `pdf_generated_at` | `timestamp with time zone` | yes | Baseline column. |
| `reconstruction_attempts` | `jsonb` | no | Baseline column. |
| `controlled_document_id` | `uuid` | yes | Baseline column. |
| `profile_code_snapshot` | `text` | yes | Baseline column. |
| `process_area_code_snapshot` | `text` | yes | Baseline column. |
| `code` | `text` | yes | Baseline column. |
| `created_by_display_name_snapshot` | `text` | yes | Baseline column. |
| `area_name_snapshot` | `text` | yes | Baseline column. |

## Baseline Definition

```sql
CREATE TABLE public.documents (
id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    template_version_id uuid NOT NULL,
    name text NOT NULL,
    status text NOT NULL,
    form_data_json jsonb NOT NULL,
    current_revision_id uuid,
    active_session_id uuid,
    finalized_at timestamp with time zone,
    archived_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by text NOT NULL,
    effective_from timestamp with time zone,
    effective_to timestamp with time zone,
    revision_number integer DEFAULT 1 NOT NULL,
    revision_version integer DEFAULT 0 NOT NULL,
    content_hash_at_submit text,
    placeholder_schema_snapshot jsonb,
    placeholder_schema_hash bytea,
    composition_config_snapshot jsonb,
    composition_config_hash bytea,
    body_docx_snapshot_s3_key text,
    body_docx_hash bytea,
    values_frozen_at timestamp with time zone,
    values_hash bytea,
    final_docx_s3_key text,
    content_hash bytea,
    final_pdf_s3_key text,
    pdf_hash bytea,
    pdf_generated_at timestamp with time zone,
    reconstruction_attempts jsonb DEFAULT '[]'::jsonb NOT NULL,
    controlled_document_id uuid,
    profile_code_snapshot text,
    process_area_code_snapshot text,
    code text,
    created_by_display_name_snapshot text,
    area_name_snapshot text,
    CONSTRAINT documents_body_docx_hash_len CHECK (((body_docx_hash IS NULL) OR (octet_length(body_docx_hash) = 32))),
    CONSTRAINT documents_composition_config_hash_len CHECK (((composition_config_hash IS NULL) OR (octet_length(composition_config_hash) = 32))),
    CONSTRAINT documents_content_hash_len CHECK (((content_hash IS NULL) OR (octet_length(content_hash) = 32))),
    CONSTRAINT documents_name_not_empty CHECK ((length(TRIM(BOTH FROM name)) > 0)),
    CONSTRAINT documents_pdf_hash_len CHECK (((pdf_hash IS NULL) OR (octet_length(pdf_hash) = 32))),
    CONSTRAINT documents_placeholder_schema_hash_len CHECK (((placeholder_schema_hash IS NULL) OR (octet_length(placeholder_schema_hash) = 32))),
    CONSTRAINT documents_reconstruction_attempts_is_array CHECK ((jsonb_typeof(reconstruction_attempts) = 'array'::text)),
    CONSTRAINT documents_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'finalized'::text, 'archived'::text, 'under_review'::text, 'approved'::text, 'rejected'::text, 'scheduled'::text, 'published'::text, 'superseded'::text, 'obsolete'::text]))),
    CONSTRAINT documents_values_hash_len CHECK (((values_hash IS NULL) OR (octet_length(values_hash) = 32)))
);
```

## Runtime Usage

Use `rg -n "documents" internal apps` and the owning module wiki to verify readers/writers before changing this table.

## Seed or Reference Data

Check `db/reference-data/0001_product_reference_data.sql` and `db/dev-seeds/0001_local_dev_seed.sql` before adding rows.

## Notes and Debt

Retained in `public` because current runtime/baseline truth still uses it. Do not move schemas without an approved migration plan.
