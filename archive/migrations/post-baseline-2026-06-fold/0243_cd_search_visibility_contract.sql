-- 0243: controlleddocuments publishes its search visibility + projection READ CONTRACT
-- (mission backend-module-boundary-hardening, M4/F4.1).
--
-- ADR-0039 D3a/D4: the search module (internal/modules/search/infrastructure/
-- v2documents/reader.go) reads THESE published views, never CD's base tables
-- (public.controlled_documents / controlled_document_area_grants /
-- controlled_document_user_grants). Two views, by job:
--
--   metaldocs.v_cd_search_facts — EXACTLY 1 row per controlled document. The
--     projection LEFT JOIN target (code/department/profile/sequence) PLUS the two
--     unbounded visibility legs as scalar columns: is_company (= visibility_scope =
--     'company', so the consumer never hardcodes CD's scope enum) and owner_user_id.
--     Strictly 1:1 so search's projection join cannot fan out document rows.
--
--   metaldocs.v_cd_grantee — the BOUNDED visibility edges: one (cd, actor) row per
--     actor who may see a RESTRICTED cd via (a) an ACTIVE area-grant membership —
--     controlled_document_area_grants joined to iam's published
--     metaldocs.v_active_user_areas (which encodes effective_to IS NULL, ADR 0037 D1,
--     so revoked members are already excluded) — or (b) a direct user-grant. Gated on
--     visibility_scope = 'restricted' to mirror the inline predicate EXACTLY (grants
--     only confer visibility on restricted CDs). Company-scope CDs contribute no edges
--     (visibility is via v_cd_search_facts.is_company), so the unbounded (cd × actor)
--     cross-product is never materialized.
--
-- The search visibility decision the consumer composes from these (== the verbatim
-- pre-M4 inline predicate, reader.go:89-118 CD branch):
--   f.is_company OR f.owner_user_id = $actor
--   OR EXISTS (SELECT 1 FROM metaldocs.v_cd_grantee g
--              WHERE g.tenant_id = f.tenant_id
--                AND g.controlled_document_id = f.controlled_document_id
--                AND g.grantee_user_id = $actor)
--
-- Both views read only CD-owned base tables + iam's PUBLISHED v_active_user_areas
-- (compliant D3a) — no third module's base table. Security posture matches the
-- underlying tables (no security_invoker), identical to 0242's v_active_user_areas.

BEGIN;

CREATE VIEW metaldocs.v_cd_search_facts AS
  SELECT cd.tenant_id,
         cd.id AS controlled_document_id,
         cd.code,
         cd.department_code,
         cd.profile_code,
         cd.sequence_num,
         (cd.visibility_scope = 'company') AS is_company,
         cd.owner_user_id
    FROM public.controlled_documents cd;

COMMENT ON VIEW metaldocs.v_cd_search_facts IS
  'Published CD search projection + scalar visibility legs (ADR-0039 D3a/D4). 1 row/CD: projection cols (code, department_code, profile_code, sequence_num) + is_company (= visibility_scope = ''company'') + owner_user_id. Non-owner modules (search) read THIS view, never public.controlled_documents. Mission backend-module-boundary-hardening M4/F4.1.';

CREATE VIEW metaldocs.v_cd_grantee AS
  SELECT cd.tenant_id,
         cd.id AS controlled_document_id,
         upa.user_id AS grantee_user_id
    FROM public.controlled_documents cd
    JOIN public.controlled_document_area_grants cdag
      ON cdag.tenant_id = cd.tenant_id
     AND cdag.controlled_document_id = cd.id
    JOIN metaldocs.v_active_user_areas upa
      ON upa.tenant_id = cd.tenant_id
     AND upa.area_code = cdag.area_code
   WHERE cd.visibility_scope = 'restricted'
  UNION
  SELECT cd.tenant_id,
         cd.id AS controlled_document_id,
         cdug.user_id AS grantee_user_id
    FROM public.controlled_documents cd
    JOIN public.controlled_document_user_grants cdug
      ON cdug.tenant_id = cd.tenant_id
     AND cdug.controlled_document_id = cd.id
   WHERE cd.visibility_scope = 'restricted';

COMMENT ON VIEW metaldocs.v_cd_grantee IS
  'Published CD bounded visibility edges (ADR-0039 D3a/D4): one (tenant_id, controlled_document_id, grantee_user_id) row per actor who may see a RESTRICTED CD via an ACTIVE area-grant membership (metaldocs.v_active_user_areas, effective_to IS NULL, ADR 0037 D1) or a direct user-grant. Gated on visibility_scope = ''restricted'' to mirror the inline search predicate exactly; revoked members and company-scope cross-products are excluded. Non-owner modules (search) read THIS view, never CD''s grant base tables. Mission backend-module-boundary-hardening M4/F4.1.';

INSERT INTO public.schema_migrations (version, description)
VALUES ('0243', 'controlleddocuments publishes v_cd_search_facts + v_cd_grantee search visibility contract (M4/F4.1, ADR-0039 D3a)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
