-- MetalDocs database prerequisites.
-- This file is applied before product schema, reference data, dev seeds, and tail migrations.

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;
-- pg_trgm backs the controlled-document trigram search indexes. It was introduced
-- by the (now folded) migration 0239_cd_trgm_search.sql; on the baseline squash its
-- extension ownership moves here so the curated baseline stays extension-free and
-- prerequisites remains the single home for extension setup.
CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;
