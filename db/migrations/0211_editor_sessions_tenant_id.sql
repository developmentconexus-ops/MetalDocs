-- 0211_editor_sessions_tenant_id.sql
BEGIN;

ALTER TABLE public.editor_sessions
    ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff';

UPDATE public.editor_sessions es
   SET tenant_id = d.tenant_id
  FROM public.documents d
 WHERE es.document_id = d.id
   AND es.tenant_id = 'ffffffff-ffff-ffff-ffff-ffffffffffff';

CREATE INDEX IF NOT EXISTS idx_editor_sessions_tenant_id
    ON public.editor_sessions (tenant_id);

INSERT INTO public.schema_migrations (version, description)
VALUES ('0211', 'add tenant_id to editor_sessions for document tenant isolation')
ON CONFLICT (version) DO NOTHING;

COMMIT;
