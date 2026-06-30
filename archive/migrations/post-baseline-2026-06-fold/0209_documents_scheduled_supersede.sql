BEGIN;

ALTER TABLE public.documents
  ADD COLUMN IF NOT EXISTS superseded_document_id uuid
  REFERENCES public.documents(id);

CREATE INDEX IF NOT EXISTS idx_documents_superseded_document_id
  ON public.documents (superseded_document_id)
  WHERE superseded_document_id IS NOT NULL;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0209', 'add scheduled supersede target on documents')
ON CONFLICT (version) DO NOTHING;

COMMIT;
