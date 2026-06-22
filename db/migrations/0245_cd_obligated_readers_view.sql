-- 0245: controlleddocuments publishes metaldocs.v_cd_obligated_readers — the
-- denominator read contract for the distribution module (mission frontend-
-- screen-completion, M2/F2.1a; ADR-0040; ADR-0039 D3a/D4 inventory).
--
-- ADR-0039 D3a/D4: the distribution module (internal/modules/distribution, built in
-- F2.1c/F2.2) reads THIS view (not CD's base tables) to derive the obligated-reader
-- set for /api/v1/documents/:id/distribution*. Three legs UNION'd, DISTINCT BY
-- (tenant_id, controlled_document_id, user_id) with source precedence
-- user_grant > area_grant > company_scope (most-specific wins). For area_grant
-- rows on a user with multiple granting areas: lowest area_code wins
-- (deterministic).
--
-- Why a new sibling view (not extending v_cd_grantee):
--   v_cd_grantee is restricted-only by design (migration 0243 COMMENT + the
--   WHERE visibility_scope='restricted' gate) — that gate is the search-semantic
--   contract making search's EXISTS predicate (search/.../v2documents/reader.go)
--   correct-by-construction. Mutating it forces search to carry distribution-domain
--   knowledge → module-boundary leak. There is also zero DROP/ALTER VIEW precedent
--   across 244 prior migrations (wiki/database/migration-policy.md is forward-only).
--   New sibling view = clean. See ADR-0040.
--
-- Reads (compliant per ADR-0039):
--   own base tables: public.controlled_documents,
--                    public.controlled_document_user_grants,
--                    public.controlled_document_area_grants
--   published views: metaldocs.v_active_user_areas (iam, D3a — encodes
--                    effective_to IS NULL, ADR 0037 D1),
--                    metaldocs.v_cd_search_facts (CD-owned, is_company scalar)
--
-- Security posture matches the underlying tables (no security_invoker), identical
-- to 0242/0243.

BEGIN;

CREATE VIEW metaldocs.v_cd_obligated_readers AS
WITH legs AS (
  -- Leg 1: direct user-grant. source_rank=1 (highest precedence).
  SELECT cdug.tenant_id,
         cdug.controlled_document_id,
         cdug.user_id,
         NULL::text          AS area_code,
         'user_grant'::text  AS source,
         1                   AS source_rank
    FROM public.controlled_document_user_grants cdug

  UNION ALL

  -- Leg 2: area-grant ⋈ active area membership. source_rank=2.
  SELECT cdag.tenant_id,
         cdag.controlled_document_id,
         upa.user_id,
         upa.area_code       AS area_code,
         'area_grant'::text  AS source,
         2                   AS source_rank
    FROM public.controlled_document_area_grants cdag
    JOIN metaldocs.v_active_user_areas upa
      ON upa.tenant_id = cdag.tenant_id
     AND upa.area_code = cdag.area_code

  UNION ALL

  -- Leg 3: company-scope CDs × DISTINCT active tenant users.
  -- Consumes CD's own v_cd_search_facts.is_company (1:1 over controlled_documents)
  -- — no hardcoded scope literal.
  SELECT f.tenant_id,
         f.controlled_document_id,
         tu.user_id,
         NULL::text             AS area_code,
         'company_scope'::text  AS source,
         3                      AS source_rank
    FROM metaldocs.v_cd_search_facts f
    JOIN (SELECT DISTINCT tenant_id, user_id FROM metaldocs.v_active_user_areas) tu
      ON tu.tenant_id = f.tenant_id
   WHERE f.is_company
)
SELECT DISTINCT ON (tenant_id, controlled_document_id, user_id)
       tenant_id,
       controlled_document_id,
       user_id,
       area_code,
       source
  FROM legs
 ORDER BY tenant_id, controlled_document_id, user_id, source_rank, area_code NULLS LAST;

COMMENT ON VIEW metaldocs.v_cd_obligated_readers IS
  'Published CD obligated-reader read contract (ADR-0040; ADR-0039 D3a/D4): one (tenant_id, controlled_document_id, user_id, area_code NULL, source) row per user obligated to read a CD. Three legs UNION''d (user_grant ∪ active area_grant member ∪ company-scope active tenant user) and DISTINCT BY (tenant_id, cd, user_id) with source precedence user_grant > area_grant > company_scope. Non-owner modules (distribution) read THIS view, never CD''s grant base tables. v_cd_grantee remains restricted-only by design (search semantics, migration 0243). Mission frontend-screen-completion M2/F2.1a.';

INSERT INTO public.schema_migrations (version, description)
VALUES ('0245', 'controlleddocuments publishes metaldocs.v_cd_obligated_readers obligated-reader view (M2/F2.1a, ADR-0040)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
