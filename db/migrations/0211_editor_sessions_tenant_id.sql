-- 0211_editor_sessions_tenant_id.sql
BEGIN;

ALTER TABLE public.editor_sessions
    ADD COLUMN IF NOT EXISTS tenant_id UUID;

UPDATE public.editor_sessions es
   SET tenant_id = d.tenant_id
  FROM public.documents d
 WHERE es.document_id = d.id
   AND (es.tenant_id IS NULL OR es.tenant_id = 'ffffffff-ffff-ffff-ffff-ffffffffffff');

ALTER TABLE public.editor_sessions
    ALTER COLUMN tenant_id DROP DEFAULT;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM public.editor_sessions
         WHERE tenant_id IS NULL
            OR tenant_id = 'ffffffff-ffff-ffff-ffff-ffffffffffff'
    ) THEN
        RAISE EXCEPTION '0211_editor_sessions_tenant_id: tenant_id backfill incomplete';
    END IF;
END $$;

ALTER TABLE public.editor_sessions
    ALTER COLUMN tenant_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_editor_sessions_tenant_id
    ON public.editor_sessions (tenant_id);

INSERT INTO public.schema_migrations (version, description)
VALUES ('0211', 'add tenant_id to editor_sessions for document tenant isolation')
ON CONFLICT (version) DO NOTHING;

COMMIT;
