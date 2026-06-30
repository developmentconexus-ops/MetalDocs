-- 0221_tenant_plans.sql
-- PR-8 of the 12-PR IAM Admin Center rebuild.
--
-- Adds the metaldocs.tenant_plans table that backs the read-only Usage
-- card on the Admin Center Overview ("you have used X of Y seats / storage
-- / api calls" + tier label).
--
-- This PR is READ-ONLY against this table. Mutation flows (upgrade /
-- downgrade / overage billing) belong to the Tier-A platform-owner surface
-- and live outside the tenant-admin observability scope — tracked in
-- wiki/modules/iam-tech-debt.md.
--
-- Defaults reflect the v1 "Pro" plan envelope: 100 seats, 50 GiB storage,
-- 1M api calls / month. Every existing tenant is backfilled at the same
-- envelope so the FE consumer can render a non-null "of" denominator.
--
-- Forward-only per wiki/database/migration-policy.md. Idempotent via
-- CREATE TABLE IF NOT EXISTS + INSERT … ON CONFLICT DO NOTHING.

BEGIN;

CREATE TABLE IF NOT EXISTS metaldocs.tenant_plans (
    tenant_id               uuid PRIMARY KEY,
    plan_tier               text NOT NULL DEFAULT 'pro'
                                CHECK (plan_tier IN ('free', 'pro', 'enterprise')),
    seats_allocated         integer NOT NULL DEFAULT 100
                                CHECK (seats_allocated >= 0),
    storage_allocated_bytes bigint  NOT NULL DEFAULT 53687091200
                                CHECK (storage_allocated_bytes >= 0),
    api_calls_allocated     integer NOT NULL DEFAULT 1000000
                                CHECK (api_calls_allocated >= 0),
    created_at              timestamptz NOT NULL DEFAULT now(),
    updated_at              timestamptz NOT NULL DEFAULT now()
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
          FROM pg_constraint
         WHERE conrelid = 'metaldocs.tenant_plans'::regclass
           AND conname = 'tenant_plans_tenant_id_fkey'
    ) THEN
        EXECUTE 'ALTER TABLE ONLY metaldocs.tenant_plans '
             || 'ADD CONSTRAINT tenant_plans_tenant_id_fkey '
             || 'FOREIGN KEY (tenant_id) REFERENCES metaldocs.tenants(id) ON DELETE CASCADE';
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_tenant_plans_plan_tier
    ON metaldocs.tenant_plans (plan_tier);

INSERT INTO metaldocs.tenant_plans (tenant_id)
SELECT id FROM metaldocs.tenants
ON CONFLICT (tenant_id) DO NOTHING;

COMMENT ON TABLE  metaldocs.tenant_plans                         IS 'PR-8: read-only plan envelope per tenant (seats/storage/api_calls allocated, tier label). Mutation owned by Tier-A platform-owner surface.';
COMMENT ON COLUMN metaldocs.tenant_plans.plan_tier               IS 'free | pro | enterprise — display label only at Tier-B.';
COMMENT ON COLUMN metaldocs.tenant_plans.seats_allocated         IS 'Seat quota; consumed by COUNT(iam_users WHERE is_active).';
COMMENT ON COLUMN metaldocs.tenant_plans.storage_allocated_bytes IS 'Storage quota in bytes; consumed by attachment/blob aggregation (tech-debt: real consumption pending).';
COMMENT ON COLUMN metaldocs.tenant_plans.api_calls_allocated     IS 'API call quota per billing window; consumed by audit/request-log aggregation (tech-debt: source not wired).';

INSERT INTO public.schema_migrations (version, description)
VALUES ('0221', 'PR-8: tenant_plans table for Admin Center Usage card (read-only at Tier-B)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
