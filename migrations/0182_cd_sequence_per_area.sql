-- 0182_cd_sequence_per_area.sql
-- WARNING: dev-only DB. Drops smoke fixtures + per-profile counters.
-- Confirm no prod tenants have allocated codes before applying.
-- Revert path: restore profile_sequence_counters from git + re-seed controlled_documents.

BEGIN;

DROP TABLE IF EXISTS profile_sequence_counters;
DELETE FROM controlled_documents;

CREATE TABLE IF NOT EXISTS cd_sequence_counters (
  tenant_id          UUID NOT NULL,
  profile_code       TEXT NOT NULL,
  process_area_code  TEXT NOT NULL,
  next_seq           INT  NOT NULL DEFAULT 1,
  PRIMARY KEY (tenant_id, profile_code, process_area_code),
  FOREIGN KEY (tenant_id, profile_code)
    REFERENCES metaldocs.document_profiles (tenant_id, code),
  FOREIGN KEY (tenant_id, process_area_code)
    REFERENCES metaldocs.document_process_areas (tenant_id, code)
);

GRANT SELECT, INSERT, UPDATE ON TABLE cd_sequence_counters TO metaldocs_app;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0182', 'cd_sequence_counters per (profile, area)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
