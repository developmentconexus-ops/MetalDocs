-- MetalDocs database prerequisites.
-- This file is applied before product schema, reference data, dev seeds, and tail migrations.
--
-- Requires superuser (CREATE EXTENSION is a superuser-only operation absent a
-- pre-trusted extension allowlist), so it is applied by whichever identity
-- runs the bootstrap-superuser provisioning phase: compose initdb.d,
-- scripts/dev-bootstrap-baseline.ps1, and, as of A6.1 (issue #88), the
-- metaldocs-dbprovision one-shot binary via internal/platform/migrate.ApplyGrants
-- (which enforces the BEGIN/COMMIT guard below on every file it applies).

BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;
-- pg_trgm backs the controlled-document trigram search indexes. It was introduced
-- by the (now folded) migration 0239_cd_trgm_search.sql; on the baseline squash its
-- extension ownership moves here so the curated baseline stays extension-free and
-- prerequisites remains the single home for extension setup.
CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;

COMMIT;
