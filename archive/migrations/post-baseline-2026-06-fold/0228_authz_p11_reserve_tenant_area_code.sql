-- 0228_authz_p11_reserve_tenant_area_code.sql
-- ADR 0022 Phase 11 (F7) — reserve the area_code value 'tenant'.
--
-- The literal "tenant" is the load-bearing tier-2 area-filter sentinel: in
-- internal/modules/iam/authz/authz.go the area predicate is
--   ($2 = 'tenant' OR upa.area_code = $2)
-- so passing areaCode = "tenant" intentionally disables the area filter
-- (degenerating tier-2 to a tier-1-equivalent check). Nothing previously stopped
-- a real process area from being created with code = 'tenant'; such a row would
-- silently switch OFF area scoping for every area-grade capability resolved
-- against it (a privilege-escalation footgun). The existing area_code_format
-- CHECK forces lowercase, so the exact lowercase 'tenant' is the only collision.
--
-- Reserve it with a CHECK constraint so a real area can never shadow the sentinel.
-- Forward-only, idempotent (DROP IF EXISTS + ADD re-runs cleanly). The ADD
-- validates existing rows: if any environment already holds a 'tenant' area the
-- migration fails LOUDLY — that is the correct surfacing of a latent scoping
-- defect, not something to silence. Mirrored inline in the curated baseline
-- db/baseline/0001_current_schema.sql (CONSTRAINT area_code_not_tenant).

BEGIN;

ALTER TABLE metaldocs.document_process_areas
    DROP CONSTRAINT IF EXISTS area_code_not_tenant;
ALTER TABLE metaldocs.document_process_areas
    ADD CONSTRAINT area_code_not_tenant CHECK (code <> 'tenant');

INSERT INTO public.schema_migrations (version, description)
VALUES ('0228', 'ADR 0022 Phase 11 (F7): reserve area_code=tenant — a real process area can no longer shadow the tier-2 area-filter sentinel')
ON CONFLICT (version) DO NOTHING;

COMMIT;
