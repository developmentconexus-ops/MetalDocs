-- 0246: taxonomy publishes metaldocs.v_process_area_name — the per-area
-- human-label read contract for the distribution module (mission frontend-
-- screen-completion, M2/F2.1b; ADR-0041; ADR-0039 D3a/D4 inventory).
--
-- ADR-0039 D3a/D4: the distribution module (internal/modules/distribution, built
-- in F2.1c/F2.2) joins THIS view to F2.1a's metaldocs.v_cd_obligated_readers on
-- (tenant_id, area_code) to populate DistributionRecipient.area_name and
-- DistributionAreaCoverage.area_name. The base table metaldocs.document_process_areas
-- is taxonomy-owned and must not be read directly by non-taxonomy modules
-- (hgcrossmodule analyzer; ADR-0039).
--
-- Why a 1:1 projection (no is_active / archived_at filter):
--   The existing taxonomy AreaCatalogReader port
--   (internal/modules/taxonomy/infrastructure/area_catalog_reader.go:32) reads the
--   base table without any active/archived filter — names resolve for archived
--   areas too. The view preserves that semantic. Adding a filter now would change
--   contract on the existing port's behavior and is YAGNI (spec.md Q5: minimal
--   contract, additive later if a consumer actually needs it).
--
-- Renames (code → area_code, name → area_name):
--   align the published shape with F2.1a's v_cd_obligated_readers.area_code so the
--   F2.2 join is natural.
--
-- Reads (compliant per ADR-0039):
--   own base tables: metaldocs.document_process_areas (taxonomy-owned)
--
-- Security posture matches the underlying table (no security_invoker), identical
-- to 0242 / 0243 / 0245.

BEGIN;

CREATE VIEW metaldocs.v_process_area_name AS
SELECT tenant_id,
       code AS area_code,
       name AS area_name
  FROM metaldocs.document_process_areas;

COMMENT ON VIEW metaldocs.v_process_area_name IS
  'Published taxonomy per-area human-label read contract (ADR-0041; ADR-0039 D3a/D4): one (tenant_id, area_code, area_name) row per (tenant_id, area_code), 1:1 projection of metaldocs.document_process_areas (code → area_code, name → area_name). No is_active / archived_at filter — labels resolve for archived areas (parity with internal/modules/taxonomy/infrastructure/area_catalog_reader.go). Sole consumer: distribution module (F2.2), joining on (tenant_id, area_code) to F2.1a''s v_cd_obligated_readers. Mission frontend-screen-completion M2/F2.1b.';

INSERT INTO public.schema_migrations (version, description)
VALUES ('0246', 'taxonomy publishes metaldocs.v_process_area_name per-area label view (M2/F2.1b, ADR-0041)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
