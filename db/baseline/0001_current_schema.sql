-- MetalDocs curated current-state schema baseline.
-- This file contains product schema only. Product reference data belongs in
-- db/reference-data/0001_product_reference_data.sql. Local-only seed data belongs
-- in db/dev-seeds/0001_local_dev_seed.sql.

BEGIN;

CREATE SCHEMA IF NOT EXISTS metaldocs;

CREATE TABLE IF NOT EXISTS public.schema_migrations (
  version TEXT PRIMARY KEY,
  description TEXT NOT NULL DEFAULT '',
  applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMIT;
