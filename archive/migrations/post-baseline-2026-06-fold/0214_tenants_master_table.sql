-- 0214_tenants_master_table.sql
BEGIN;

CREATE TABLE IF NOT EXISTS metaldocs.tenants (
    id uuid NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT tenants_name_not_blank CHECK ((length(btrim(name)) > 0)),
    CONSTRAINT tenants_slug_not_blank CHECK ((length(btrim(slug)) > 0))
);

DO $$
DECLARE
    auth_sessions_tenant_type text;
    invalid_auth_sessions text;
BEGIN
    SELECT data_type
      INTO auth_sessions_tenant_type
      FROM information_schema.columns
     WHERE table_schema = 'metaldocs'
       AND table_name = 'auth_sessions'
       AND column_name = 'tenant_id';

    IF auth_sessions_tenant_type = 'text' THEN
        SELECT string_agg(tenant_id, ', ' ORDER BY tenant_id)
          INTO invalid_auth_sessions
          FROM (
            SELECT DISTINCT tenant_id
              FROM metaldocs.auth_sessions
             WHERE tenant_id IS NULL
                OR tenant_id !~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
          ) invalid_values;

        IF invalid_auth_sessions IS NOT NULL THEN
            RAISE EXCEPTION
                '0214_tenants_master_table: backfill required for metaldocs.auth_sessions.tenant_id values: %',
                invalid_auth_sessions;
        END IF;

        EXECUTE $alter$
            ALTER TABLE metaldocs.auth_sessions
                ALTER COLUMN tenant_id TYPE UUID
                USING tenant_id::uuid
        $alter$;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
          FROM pg_constraint
         WHERE conrelid = 'metaldocs.tenants'::regclass
           AND conname = 'tenants_pkey'
    ) THEN
        EXECUTE 'ALTER TABLE ONLY metaldocs.tenants ADD CONSTRAINT tenants_pkey PRIMARY KEY (id)';
    END IF;

    IF NOT EXISTS (
        SELECT 1
          FROM pg_constraint
         WHERE conrelid = 'metaldocs.tenants'::regclass
           AND conname = 'tenants_slug_key'
    ) THEN
        EXECUTE 'ALTER TABLE ONLY metaldocs.tenants ADD CONSTRAINT tenants_slug_key UNIQUE (slug)';
    END IF;
END $$;

INSERT INTO metaldocs.tenants (id, name, slug)
SELECT tenant_id,
       CASE
           WHEN tenant_id = 'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid THEN 'System Tenant'
           ELSE 'Tenant ' || tenant_id::text
       END,
       CASE
           WHEN tenant_id = 'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid THEN 'system'
           ELSE lower(replace(tenant_id::text, '-', ''))
       END
  FROM (
    SELECT tenant_id FROM metaldocs.auth_sessions
    UNION
    SELECT tenant_id FROM metaldocs.iam_groups
    UNION
    SELECT tenant_id FROM metaldocs.iam_user_roles
    UNION
    SELECT tenant_id FROM metaldocs.iam_users
  ) tenant_scope
ON CONFLICT (id) DO NOTHING;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
          FROM pg_constraint
         WHERE conrelid = 'metaldocs.auth_sessions'::regclass
           AND conname = 'fk_auth_sessions_tenant'
    ) THEN
        EXECUTE 'ALTER TABLE ONLY metaldocs.auth_sessions ADD CONSTRAINT fk_auth_sessions_tenant FOREIGN KEY (tenant_id) REFERENCES metaldocs.tenants(id)';
    END IF;

    IF NOT EXISTS (
        SELECT 1
          FROM pg_constraint
         WHERE conrelid = 'metaldocs.iam_groups'::regclass
           AND conname = 'iam_groups_tenant_id_fkey'
    ) THEN
        EXECUTE 'ALTER TABLE ONLY metaldocs.iam_groups ADD CONSTRAINT iam_groups_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES metaldocs.tenants(id)';
    END IF;

    IF NOT EXISTS (
        SELECT 1
          FROM pg_constraint
         WHERE conrelid = 'metaldocs.iam_user_roles'::regclass
           AND conname = 'iam_user_roles_tenant_id_fkey'
    ) THEN
        EXECUTE 'ALTER TABLE ONLY metaldocs.iam_user_roles ADD CONSTRAINT iam_user_roles_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES metaldocs.tenants(id)';
    END IF;

    IF NOT EXISTS (
        SELECT 1
          FROM pg_constraint
         WHERE conrelid = 'metaldocs.iam_users'::regclass
           AND conname = 'iam_users_tenant_id_fkey'
    ) THEN
        EXECUTE 'ALTER TABLE ONLY metaldocs.iam_users ADD CONSTRAINT iam_users_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES metaldocs.tenants(id)';
    END IF;
END $$;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0214', 'add canonical tenants master table and convert auth_sessions tenant_id to uuid')
ON CONFLICT (version) DO NOTHING;

COMMIT;
