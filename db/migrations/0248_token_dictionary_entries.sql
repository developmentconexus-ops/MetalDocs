-- Token dictionary (SP-1): per-tenant author-defined name -> value constants.
-- New module `tokens`. Capability-governed (token.view / token_dictionary.manage),
-- audited. ADR superseding 0008. RLS uses the 0247 NULL-permissive
-- tenant_isolation pattern verbatim. The name CHECK is anti-corruption storage
-- hygiene (rejects {}, ., -, whitespace, unicode) -- NOT the token grammar, which
-- stays Node-owned in @metaldocs/shared-tokens. Forward-only, idempotent.
--
-- NOTE on tenant_id: no FK to a tenants table. This MATCHES house style -- the
-- pooled multi-tenant model isolates by tenant_id + RLS, not by FK (confirmed:
-- 0247 tenant tables carry tenant_id NOT NULL with no FK). Spec's "FK ->
-- tenants(id)" is aspirational notation; the enforced contract is NOT NULL +
-- the tenant_isolation RLS policy below.

BEGIN;

CREATE TABLE IF NOT EXISTS metaldocs.token_dictionary_entries (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid        NOT NULL,
    name        text        NOT NULL
                            CHECK (name ~ '^[A-Za-z0-9_]+$')
                            CHECK (char_length(name) BETWEEN 1 AND 64),
    value       text        NOT NULL
                            CHECK (char_length(value) BETWEEN 1 AND 4096),
    label       text        NOT NULL
                            CHECK (char_length(label) BETWEEN 1 AND 256),
    description text        CHECK (description IS NULL OR char_length(description) <= 1024),
    -- Actor identity is TEXT system-wide: iam_users.user_id (PK) is text, and
    -- every actor column references it as text (audit.actor_id, notifications.
    -- recipient_user_id, idempotency_keys.actor_user_id, grant_area_membership
    -- _user_id text, UserIDFromContext -> string). These FK the same identity, so
    -- they MUST be text too -- a uuid column cannot hold a user_id (e.g. "admin").
    created_by  text        NOT NULL,
    updated_by  text        NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- (tenant_id, name) unique; the index also serves GetByName and (leading column)
-- List(tenant_id) -- no separate tenant_id index needed at SP-1.
CREATE UNIQUE INDEX IF NOT EXISTS uq_token_dictionary_tenant_name
    ON metaldocs.token_dictionary_entries (tenant_id, name);

-- RLS: 0247 pattern verbatim (ENABLE + FORCE + one NULL-permissive policy).
ALTER TABLE metaldocs.token_dictionary_entries ENABLE ROW LEVEL SECURITY;
ALTER TABLE metaldocs.token_dictionary_entries FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON metaldocs.token_dictionary_entries;
CREATE POLICY tenant_isolation ON metaldocs.token_dictionary_entries
  USING (
    NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL
    OR tenant_id = NULLIF(current_setting('metaldocs.tenant_id', true), '')::uuid
  );

INSERT INTO public.schema_migrations (version, description)
VALUES ('0248', 'token dictionary per-tenant name->value constants table (SP-1)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
