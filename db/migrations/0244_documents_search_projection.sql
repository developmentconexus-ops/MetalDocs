-- 0244: documents publishes its search projection READ CONTRACT
-- (mission backend-module-boundary-hardening, M4/F4.2).
--
-- ADR-0039 D3a/D4: the search module (internal/modules/search/infrastructure/
-- v2documents/reader.go) reads THIS published view, never the documents-owned
-- public.documents base table (Category C4a). The view is a PURE 1:1 column
-- projection — exactly the 14 columns search consumes — with NO WHERE and NO
-- COALESCE. archived_at is exposed as a column (search keeps its own
-- `archived_at IS NULL` active filter in F4.3); nullable snapshot columns pass
-- through verbatim (search keeps its own COALESCE). The contract promises
-- columns, not policy, so any future read-only consumer can reuse it without
-- re-versioning, and the F4.3 seam is a pure rename (FROM public.documents ->
-- FROM metaldocs.v_document_search_facts) with a trivially total parity proof.
--
-- Mirrors M4/F4.1's metaldocs.v_cd_search_facts (facts view = projection;
-- consumer owns its null/active logic). Reads only the documents-owned base
-- table — no third module's base table. Security posture matches the underlying
-- table (no security_invoker), identical to 0242/0243.

BEGIN;

CREATE VIEW metaldocs.v_document_search_facts AS
  SELECT d.tenant_id,
         d.id,
         d.controlled_document_id,
         d.name,
         d.status,
         d.profile_code_snapshot,
         d.process_area_code_snapshot,
         d.created_by,
         d.code,
         d.revision_number,
         d.effective_from,
         d.effective_to,
         d.created_at,
         d.archived_at
    FROM public.documents d;

COMMENT ON VIEW metaldocs.v_document_search_facts IS
  'Published documents search projection (ADR-0039 D3a/D4). Pure 1:1 projection of public.documents: the 14 columns the search consumer reads (tenant_id, id, controlled_document_id, name, status, profile_code_snapshot, process_area_code_snapshot, created_by, code, revision_number, effective_from, effective_to, created_at, archived_at). No WHERE, no COALESCE: archived_at exposed (consumer applies its own active filter), nullable cols pass through. Non-owner modules (search) read THIS view, never public.documents. Mission backend-module-boundary-hardening M4/F4.2.';

INSERT INTO public.schema_migrations (version, description)
VALUES ('0244', 'documents publishes v_document_search_facts search projection contract (M4/F4.2, ADR-0039 D3a)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
