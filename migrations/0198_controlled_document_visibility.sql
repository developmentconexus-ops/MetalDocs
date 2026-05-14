ALTER TABLE controlled_documents
  ADD COLUMN IF NOT EXISTS visibility_scope TEXT NOT NULL DEFAULT 'company',
  DROP CONSTRAINT IF EXISTS controlled_documents_visibility_scope_check,
  ADD CONSTRAINT controlled_documents_visibility_scope_check
    CHECK (visibility_scope IN ('company', 'restricted'));

CREATE TABLE IF NOT EXISTS controlled_document_area_grants (
  tenant_id UUID NOT NULL,
  controlled_document_id UUID NOT NULL,
  area_code TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, controlled_document_id, area_code),
  FOREIGN KEY (controlled_document_id)
    REFERENCES controlled_documents (id)
    ON DELETE CASCADE,
  FOREIGN KEY (tenant_id, area_code)
    REFERENCES metaldocs.document_process_areas (tenant_id, code)
);

CREATE TABLE IF NOT EXISTS controlled_document_user_grants (
  tenant_id UUID NOT NULL,
  controlled_document_id UUID NOT NULL,
  user_id TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, controlled_document_id, user_id),
  FOREIGN KEY (controlled_document_id)
    REFERENCES controlled_documents (id)
    ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS ix_cd_area_grants_tenant_area
  ON controlled_document_area_grants (tenant_id, area_code, controlled_document_id);

CREATE INDEX IF NOT EXISTS ix_cd_user_grants_tenant_user
  ON controlled_document_user_grants (tenant_id, user_id, controlled_document_id);
