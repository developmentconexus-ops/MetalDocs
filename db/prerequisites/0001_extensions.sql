-- MetalDocs database prerequisites.
-- This file is applied before product schema, reference data, dev seeds, and tail migrations.

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;
