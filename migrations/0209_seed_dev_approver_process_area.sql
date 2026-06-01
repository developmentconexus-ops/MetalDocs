-- Bind the dev 'approver' user (seeded by 0159) to a tenant process area so the
-- authz tier-1 capability check (internal/modules/iam/authz/authz.go Require)
-- can resolve template.approve / template.review / template.view via the
-- role_capabilities × user_process_areas join.
--
-- Without this row the approver user's /auth/me capabilities projection still
-- lists template.approve (it derives from role_capabilities directly) but the
-- tx-scoped authz.Require returns ErrCapDenied because the
-- upa.role = rc.role join produces zero rows. Local QA on PR #41 surfaced the
-- gap (500 → 403 after the templates handler error-mapping patch). This
-- migration closes 0159's residual seed.
--
-- Idempotent.

BEGIN;

INSERT INTO public.user_process_areas (user_id, tenant_id, area_code, role, effective_from, granted_by)
VALUES ('approver', 'ffffffff-ffff-ffff-ffff-ffffffffffff', 'quality', 'approver', now(), 'admin')
ON CONFLICT DO NOTHING;

COMMIT;
