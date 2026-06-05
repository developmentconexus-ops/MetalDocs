-- migrations/0186_role_capabilities_typed.sql
-- Plan 4 (2026-05-11): collapse capability namespace.
-- The five doc.* values seeded by 0165 + 0169 are renamed to document.*
-- so they match the typed iamdomain.Capability consts in
-- internal/modules/iam/domain/model.go. Idempotent: re-runs are no-ops
-- because the WHERE clauses no longer match after first apply.

BEGIN;

UPDATE metaldocs.role_capabilities SET capability = 'document.view'    WHERE capability = 'doc.view';
UPDATE metaldocs.role_capabilities SET capability = 'document.create'  WHERE capability = 'doc.create';
UPDATE metaldocs.role_capabilities SET capability = 'document.edit'    WHERE capability = 'doc.edit';
UPDATE metaldocs.role_capabilities SET capability = 'document.submit'  WHERE capability = 'doc.submit';
UPDATE metaldocs.role_capabilities SET capability = 'document.signoff' WHERE capability = 'doc.signoff';

INSERT INTO public.schema_migrations (version, description)
VALUES ('0186', 'Plan 4: rename doc.* role_capabilities rows to document.*')
ON CONFLICT (version) DO NOTHING;

COMMIT;
