-- Repair ledger rows for recent migrations that predated the self-recording
-- migration runner contract. Effects are idempotent and safe to replay.

BEGIN;

INSERT INTO metaldocs.role_capabilities (role, capability, created_at)
VALUES ('system_admin', 'audit.read', NOW())
ON CONFLICT (role, capability) DO NOTHING;

ALTER TABLE metaldocs.audit_events
    ADD COLUMN IF NOT EXISTS tenant_id TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_audit_events_tenant_id
    ON metaldocs.audit_events (tenant_id);

ALTER TABLE public.document_placeholder_values
  DROP CONSTRAINT IF EXISTS document_placeholder_values_revision_id_fkey;

ALTER TABLE public.document_placeholder_values
  ADD CONSTRAINT document_placeholder_values_revision_id_fkey
  FOREIGN KEY (revision_id) REFERENCES public.document_revisions(id) ON DELETE CASCADE;

INSERT INTO metaldocs.role_capabilities (role, capability, created_at)
VALUES
  ('approver', 'template.review', NOW()),
  ('system_admin', 'template.review', NOW()),
  ('qms_admin', 'template.review', NOW())
ON CONFLICT (role, capability) DO NOTHING;

INSERT INTO public.schema_migrations (version, description)
VALUES
  ('0189', 'seed audit.read capability for system_admin'),
  ('0190', 'add tenant isolation to audit_events'),
  ('0191', 'point document_placeholder_values revision FK at document_revisions'),
  ('0192', 'seed review capability for template lifecycle'),
  ('0197', 'repair recent migration ledger rows')
ON CONFLICT (version) DO NOTHING;

COMMIT;
