-- 0318_capability_bindings_schema_backfill.sql
-- ADR 0092 (grant-model unification) D1, issue #89/A8, slice A8.1.
--
-- ============================================================================
-- WHY
-- ============================================================================
--   Tier-1 (CapabilityService.CanDo) and tier-2 (authz.Require) read disjoint
--   grant tables: iam_user_roles (5 roles, no CHECK-mirrored Go set) vs
--   user_process_areas (7 roles). A principal holding only an area membership
--   is refused at tier-1 before tier-2 ever runs -- the area model is
--   unreachable through HTTP except to narrow someone tier-1 already admitted
--   (TestNoDeclaredOperationIsUnreachable, red by design until this program
--   lands). ADR 0092 D1 collapses iam_user_roles, user_process_areas, and
--   iam_group_roles into one binding relation carrying
--   (subject_kind, subject_id, role_code, scope_kind, scope_ref): the two
--   tiers become two predicates over one source (tier-1: granted at ANY
--   scope; tier-2: granted at tenant OR the resource's area), making
--   "tier-1 is a strict relaxation of tier-2" a testable property instead of
--   an accident of two tables agreeing by luck.
--
--   iam_group_members is membership (who is in a group), not a grant, and is
--   explicitly NOT one of the three source relations folded in here.
--
-- ============================================================================
-- WHAT THIS MIGRATION DOES (A8.1 scope only -- schema + reversible backfill)
-- ============================================================================
--   1. Promotes/creates the unique keys the new table's subject FKs need:
--      metaldocs.iam_users(tenant_id, user_id) and
--      metaldocs.iam_groups(tenant_id, id).
--   2. Creates metaldocs.roles -- a hand-seeded role catalog (TRANSITIONAL:
--      A8.3 "generated role catalogs + tripwire repoint" replaces the hand
--      seed with a Go-registry-driven codegen; this table and its seed are
--      the CLAUDE.md Global-Maximum-rule-labelled local maximum until then).
--   3. Creates metaldocs.capability_bindings -- the ADR 0092 D1 relation.
--   4. Backfills all rows from iam_user_roles, user_process_areas (full
--      history, not just active), and iam_group_roles, stamping
--      source_relation for provenance.
--
--   Deliberately NOT in this migration (A8.2/A8.3/A8.4, out of A8.1 scope):
--   no query builder reads capability_bindings yet; no tier reads from it;
--   the old three relations are untouched and remain the live read path.
--   The DB tripwire (enforce_capability_asserted) extension for this table
--   ships as the immediately-following migration 0319 (machine-generated
--   from internal/platform/tripwire, kept out of this hand-written file per
--   the TRIPWIRE-ARM-PARITY golden-generation contract).
--
-- ============================================================================
-- NAMED DEVIATION FROM THE ADR'S LITERAL COLUMN LIST
-- ============================================================================
--   ADR 0092 D1 names a single logical `subject_id`. Postgres cannot express
--   one FK column conditionally pointing at two different tables/types
--   (iam_users.user_id is TEXT; iam_groups.id is UUID). `subject_id` is
--   split into physical subject_user_id (text, FK -> iam_users) and
--   subject_group_id (uuid, FK -> iam_groups), discriminated by
--   subject_kind. This is a stronger reading of the ADR's own requirement
--   (referentially valid, unrepresentable-by-construction) than a single
--   polymorphic column could deliver -- a dangling subject_id is not just
--   application-validated, it is a FOREIGN KEY VIOLATION.
--
-- ============================================================================
-- REVERSIBILITY (no down-migration tooling exists; forward-only)
-- ============================================================================
--   This migration only INSERTs into the new relation; it never touches
--   iam_user_roles, user_process_areas, or iam_group_roles. Per ADR 0092
--   Consequences ("migration is reversible until the old relations are
--   dropped"), rollback is a manually-executed compensating forward
--   migration, valid for as long as that remains true:
--
--     -- partial: undo only the backfilled rows (leaves any A8.2+ native
--     -- grants, source_relation IS NULL, untouched). This DELETE is itself
--     -- tripwire-guarded (migration 0319 arm #21, same as any other write to
--     -- this table), so it needs the same capability GUC an application
--     -- write would carry -- proven end to end (seed -> partial delete ->
--     -- native row survives) against a throwaway database on 2026-08-11:
--     BEGIN;
--     SELECT set_config('metaldocs.asserted_caps', '[{"cap":"user.manage"}]', true);
--     DELETE FROM metaldocs.capability_bindings
--       WHERE source_relation IN ('iam_user_roles', 'user_process_areas', 'iam_group_roles');
--     COMMIT;
--
--     -- full: undo this migration's schema too (also revert migration 0319's
--     -- tripwire arms first -- dropping capability_bindings/roles drops their
--     -- attached trigger with them, so no separate step is required there):
--     DROP TABLE metaldocs.capability_bindings;
--     DROP TABLE metaldocs.roles;
--     ALTER TABLE metaldocs.iam_groups DROP CONSTRAINT iam_groups_tenant_id_id_uk;
--
--   NAMED LIMIT ON THE "FULL" FORM -- proven, not assumed, against a
--   throwaway database on 2026-08-11: the fourth statement,
--   `ALTER TABLE metaldocs.iam_users DROP CONSTRAINT iam_users_tenant_user_uk`,
--   FAILS without CASCADE. Four FK constraints that predate this migration --
--   approval_instances_submitted_by_tenant_fkey,
--   approval_signoffs_actor_tenant_fkey,
--   user_process_areas_granted_by_same_tenant,
--   user_process_areas_revoked_by_same_tenant (db/baseline/0001_current_schema.sql
--   lines 4246, 4278, 4550, 4558) -- already reference
--   iam_users(tenant_id, user_id) and already depended on the bare index
--   `ux_iam_users_tenant_user` for that uniqueness before this migration ever
--   ran. `ADD CONSTRAINT ... UNIQUE USING INDEX` renamed that same physical
--   index in place; it did not create a new dependency, it made a pre-existing
--   one visible. Dropping the promoted constraint therefore requires CASCADE,
--   which would silently drop those four unrelated FKs too -- out of scope
--   for A8.1 to fix and not a regression this migration introduced. Treat the
--   iam_users promotion as effectively permanent; only the iam_groups
--   promotion and the two new tables are freely reversible.
--
--   All forms are safe only while the three source relations still exist and
--   still are the live read path -- exactly the A8.1 end state.
--
-- ============================================================================
-- REPLAY SAFETY (this migration must be safe to execute a second time against
-- a database where it already succeeded -- partial restore / operator replay,
-- independent of whether public.schema_migrations' own ledger row would have
-- skipped a framework-driven re-run)
-- ============================================================================
--   Every DDL statement below is guarded so a second execution is a no-op,
--   not a duplicate_object error:
--     - the two ALTER TABLE ... ADD CONSTRAINT promotions have no native
--       IF NOT EXISTS form in Postgres, so each is wrapped in a DO block that
--       checks pg_constraint by (conname, conrelid) first;
--     - both CREATE TABLE statements use IF NOT EXISTS;
--     - the roles seed INSERT uses ON CONFLICT (code) DO NOTHING;
--     - all three CREATE INDEX statements use IF NOT EXISTS;
--     - ENABLE/FORCE ROW LEVEL SECURITY are already idempotent in Postgres
--       (no error re-enabling/re-forcing what is already on) -- no guard
--       needed;
--     - CREATE POLICY has no IF NOT EXISTS form, so it is preceded by
--       DROP POLICY IF EXISTS on the same name.
--
--   The three backfill INSERT ... SELECT statements are the one place a
--   naive guard would be wrong: capability_bindings has no uniqueness
--   constraint over historical (effective_to IS NOT NULL) rows at all --
--   ux_capability_bindings_active_identity only covers the active slice --
--   so replaying the bare INSERTs would silently duplicate revoked history,
--   not raise an error a replay operator could see. Each is instead guarded
--   by `WHERE NOT EXISTS (... WHERE source_relation = '<this source>')`:
--   once any row has been stamped with a given source_relation, that
--   source's SELECT produces zero rows on every subsequent run. This is
--   coarse (all-or-nothing per source, not per-row) but correct for this
--   migration's actual write pattern -- each source is backfilled in one
--   all-or-nothing statement, so "some rows already exist for this source"
--   and "this source is fully backfilled" are the same state in practice.
--   It also self-heals a partial failure: if source 2 fails after source 1
--   committed within the same migration transaction this is moot (the whole
--   BEGIN/COMMIT rolls back together), but it keeps the statements correct
--   in isolation if this file is ever run outside that wrapping transaction.
--
--   Proof this holds: tests/integration/migrations/migration_0318_test.go
--   runs this file's statements twice against the same database and asserts
--   the second run succeeds with capability_bindings row counts unchanged.
--
-- ============================================================================

BEGIN;

-- ── prerequisite unique keys (FK targets for capability_bindings) ──────────

-- ux_iam_users_tenant_user is a unique INDEX, not a unique CONSTRAINT; a bare
-- unique index cannot be an FK target. Promote it in place (no rebuild).
-- Postgres has no `ADD CONSTRAINT IF NOT EXISTS`, so replay safety is a
-- manual pg_constraint existence check (see REPLAY SAFETY header note).
DO $guard$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'iam_users_tenant_user_uk'
          AND conrelid = 'metaldocs.iam_users'::regclass
    ) THEN
        ALTER TABLE metaldocs.iam_users
          ADD CONSTRAINT iam_users_tenant_user_uk UNIQUE USING INDEX ux_iam_users_tenant_user;
    END IF;
END;
$guard$;

-- No prior uniqueness on (tenant_id, id) exists for iam_groups at all
-- (id alone is already PK-unique); cheap fresh add.
DO $guard$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'iam_groups_tenant_id_id_uk'
          AND conrelid = 'metaldocs.iam_groups'::regclass
    ) THEN
        ALTER TABLE metaldocs.iam_groups
          ADD CONSTRAINT iam_groups_tenant_id_id_uk UNIQUE (tenant_id, id);
    END IF;
END;
$guard$;

-- ── metaldocs.roles: hand-seeded catalog (TRANSITIONAL, see header) ────────

CREATE TABLE IF NOT EXISTS metaldocs.roles (
    code text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT roles_pkey PRIMARY KEY (code),
    CONSTRAINT chk_roles_code_format CHECK ((code ~ '^[a-z][a-z0-9_]*$'::text))
);

-- The union of every role-declaration surface known at A8.1 time (ADR 0092
-- Context: iamtypes.validRoles, iamtypes.areaRoles, OpenAPI UserRole/AreaRole,
-- the user_process_areas CHECK (7), the iam_user_roles CHECK (5)). A8.3
-- repoints this seed at a Go-registry-driven generator; until then this INSERT
-- is the single hand-maintained catalog source.
INSERT INTO metaldocs.roles (code, description) VALUES
    ('system_admin', 'Tenant-wide administrator; every capability (ADR 0092 D2 target -- not yet an ordinary grant in A8.1)'),
    ('approver',     'Tenant-scoped approver role'),
    ('author',       'Tenant-scoped author role'),
    ('editor',       'Tenant-scoped editor role'),
    ('viewer',       'Tenant-scoped viewer role'),
    ('signer',       'Area-scoped signer role'),
    ('area_admin',   'Area-scoped administrator role'),
    ('qms_admin',    'Area-scoped QMS administrator role')
ON CONFLICT (code) DO NOTHING;

-- ── metaldocs.capability_bindings: ADR 0092 D1 relation ────────────────────

CREATE TABLE IF NOT EXISTS metaldocs.capability_bindings (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    tenant_id uuid DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid NOT NULL,
    subject_kind text NOT NULL,
    subject_user_id text,
    subject_group_id uuid,
    role_code text NOT NULL,
    scope_kind text NOT NULL,
    scope_ref text,
    effective_from timestamp with time zone DEFAULT now() NOT NULL,
    effective_to timestamp with time zone,
    granted_by text,
    revoked_by text,
    source_relation text,
    CONSTRAINT capability_bindings_pkey PRIMARY KEY (id),
    CONSTRAINT chk_capability_bindings_subject_kind
        CHECK ((subject_kind = ANY (ARRAY['user'::text, 'group'::text]))),
    -- Unrepresentable-by-construction: exactly one physical subject column is
    -- set, and it must match subject_kind -- see the NAMED DEVIATION note
    -- above for why subject_id could not be a single polymorphic column.
    CONSTRAINT chk_capability_bindings_subject_shape CHECK (
        ((subject_kind = 'user'::text) AND (subject_user_id IS NOT NULL) AND (subject_group_id IS NULL))
        OR
        ((subject_kind = 'group'::text) AND (subject_group_id IS NOT NULL) AND (subject_user_id IS NULL))
    ),
    CONSTRAINT chk_capability_bindings_scope_kind
        CHECK ((scope_kind = ANY (ARRAY['tenant'::text, 'area'::text]))),
    -- Unrepresentable-by-construction: a tenant-scope binding cannot name an
    -- area, and an area-scope binding must name one.
    CONSTRAINT chk_capability_bindings_scope_shape CHECK (
        ((scope_kind = 'tenant'::text) AND (scope_ref IS NULL))
        OR
        ((scope_kind = 'area'::text) AND (scope_ref IS NOT NULL))
    ),
    -- Mirrors user_process_areas' temporal-model CHECKs (ADR 0037) exactly:
    -- active iff effective_to IS NULL; revoke stamps a past tombstone +
    -- revoked_by; no future-dated effective_to is ever written.
    CONSTRAINT chk_capability_bindings_effective_interval
        CHECK (((effective_to IS NULL) OR (effective_to > effective_from))),
    CONSTRAINT chk_capability_bindings_revoked_by_required CHECK (
        (((effective_to IS NULL) AND (revoked_by IS NULL))
         OR ((effective_to IS NOT NULL) AND (revoked_by IS NOT NULL)))
    ),
    CONSTRAINT chk_capability_bindings_source_relation CHECK (
        (source_relation IS NULL)
        OR (source_relation = ANY (ARRAY['iam_user_roles'::text, 'user_process_areas'::text, 'iam_group_roles'::text]))
    ),
    CONSTRAINT fk_capability_bindings_tenant
        FOREIGN KEY (tenant_id) REFERENCES metaldocs.tenants(id),
    CONSTRAINT fk_capability_bindings_role
        FOREIGN KEY (role_code) REFERENCES metaldocs.roles(code),
    -- MATCH SIMPLE (default): a NULL component skips enforcement, so a
    -- 'group' subject row (subject_user_id NULL) is correctly exempt from
    -- this FK, and vice versa -- the discriminated-union shape falls out of
    -- Postgres's own FK semantics, not app-level branching.
    CONSTRAINT fk_capability_bindings_subject_user
        FOREIGN KEY (tenant_id, subject_user_id) REFERENCES metaldocs.iam_users(tenant_id, user_id),
    CONSTRAINT fk_capability_bindings_subject_group
        FOREIGN KEY (tenant_id, subject_group_id) REFERENCES metaldocs.iam_groups(tenant_id, id),
    CONSTRAINT fk_capability_bindings_scope_area
        FOREIGN KEY (tenant_id, scope_ref) REFERENCES metaldocs.document_process_areas(tenant_id, code)
);

-- Active-identity dedup, mirroring user_process_areas' single-active-row
-- pattern (ADR 0037): at most one ACTIVE (effective_to IS NULL) binding per
-- (tenant, subject, role, scope). COALESCE is required because Postgres
-- unique indexes treat NULL as distinct-from-itself, which would otherwise
-- defeat dedup across the discriminated-union subject/scope columns.
CREATE UNIQUE INDEX IF NOT EXISTS ux_capability_bindings_active_identity ON metaldocs.capability_bindings (
    tenant_id,
    subject_kind,
    COALESCE(subject_user_id, ''::text),
    COALESCE(subject_group_id::text, ''::text),
    role_code,
    scope_kind,
    COALESCE(scope_ref, ''::text)
) WHERE (effective_to IS NULL);

CREATE INDEX IF NOT EXISTS ix_capability_bindings_subject_active ON metaldocs.capability_bindings (
    tenant_id, subject_kind, subject_user_id, subject_group_id
) WHERE (effective_to IS NULL);

CREATE INDEX IF NOT EXISTS ix_capability_bindings_tenant_role_active ON metaldocs.capability_bindings (
    tenant_id, role_code, scope_kind, scope_ref
) WHERE (effective_to IS NULL);

-- Tenant isolation: FORCE + policy, identical idiom to every sibling grant
-- table (iam_user_roles, user_process_areas, iam_groups, iam_users). The
-- null-GUC branch is the deliberate janitor/bootstrap/migration escape
-- hatch, not a leak (matches every other tenant_isolation policy in the
-- baseline).
ALTER TABLE metaldocs.capability_bindings ENABLE ROW LEVEL SECURITY;
ALTER TABLE ONLY metaldocs.capability_bindings FORCE ROW LEVEL SECURITY;

-- ENABLE/FORCE above are already idempotent (no error re-asserting what is
-- already on). CREATE POLICY is not -- Postgres has no
-- `CREATE POLICY IF NOT EXISTS` -- so replay safety is DROP IF EXISTS then
-- recreate, same idiom as the guarded ADD CONSTRAINT statements above.
DROP POLICY IF EXISTS tenant_isolation ON metaldocs.capability_bindings;

CREATE POLICY tenant_isolation ON metaldocs.capability_bindings USING (
    ((NULLIF(current_setting('metaldocs.tenant_id'::text, true), ''::text) IS NULL)
     OR (tenant_id = (NULLIF(current_setting('metaldocs.tenant_id'::text, true), ''::text))::uuid))
);

-- ── backfill ─────────────────────────────────────────────────────────────
-- Runs before migration 0319 attaches the capability-assertion trigger to
-- this table, so no metaldocs.asserted_caps GUC needs to be set here -- there
-- is nothing yet to assert against. RLS is inert for the migration-owner
-- role exactly as it is for every other bulk migration sweep in this repo
-- (e.g. 0317); no tenant_id GUC is set, admitting every tenant's rows in one
-- pass.

-- iam_group_roles carries no CHECK on `role` at all (unlike iam_user_roles
-- and user_process_areas, whose role values are guaranteed subsets of the
-- 8-role catalog by their own CHECKs). Fail loudly, before the FK-enforced
-- INSERT below would reject it anyway, with a diagnostic that names the
-- offending value.
DO $migration$
DECLARE
    v_bad_role text;
BEGIN
    SELECT gr.role INTO v_bad_role
    FROM metaldocs.iam_group_roles gr
    WHERE NOT EXISTS (SELECT 1 FROM metaldocs.roles r WHERE r.code = gr.role)
    LIMIT 1;

    IF v_bad_role IS NOT NULL THEN
        RAISE EXCEPTION 'capability_bindings backfill: iam_group_roles.role value % is not present in metaldocs.roles', v_bad_role
            USING ERRCODE = 'P0001';
    END IF;
END;
$migration$;

-- Source 1: iam_user_roles -- tenant-scoped user grants.
-- Replay guard: capability_bindings has no uniqueness constraint over
-- historical (effective_to IS NOT NULL) rows, so a bare replay of this
-- INSERT would silently duplicate history rather than error. Once any row
-- is stamped source_relation = 'iam_user_roles', this source is considered
-- fully backfilled and the SELECT yields nothing on every later run (see
-- REPLAY SAFETY header note).
INSERT INTO metaldocs.capability_bindings
    (tenant_id, subject_kind, subject_user_id, subject_group_id, role_code,
     scope_kind, scope_ref, effective_from, effective_to, granted_by, revoked_by, source_relation)
SELECT
    tenant_id, 'user', user_id, NULL, role_code,
    'tenant', NULL, assigned_at, NULL, assigned_by, NULL, 'iam_user_roles'
FROM metaldocs.iam_user_roles
WHERE NOT EXISTS (
    SELECT 1 FROM metaldocs.capability_bindings WHERE source_relation = 'iam_user_roles'
);

-- Source 2: user_process_areas -- area-scoped user grants. ALL rows,
-- active and already-revoked, to preserve full grant history (backfill must
-- not flatten or lose provenance). Same replay guard as Source 1.
INSERT INTO metaldocs.capability_bindings
    (tenant_id, subject_kind, subject_user_id, subject_group_id, role_code,
     scope_kind, scope_ref, effective_from, effective_to, granted_by, revoked_by, source_relation)
SELECT
    tenant_id, 'user', user_id, NULL, role,
    'area', area_code, effective_from, effective_to, granted_by, revoked_by, 'user_process_areas'
FROM public.user_process_areas
WHERE NOT EXISTS (
    SELECT 1 FROM metaldocs.capability_bindings WHERE source_relation = 'user_process_areas'
);

-- Source 3: iam_group_roles -- tenant-scoped group grants. This table has no
-- timestamp/provenance columns of its own; effective_from anchors on the
-- owning group's created_at (the best available provenance) and granted_by
-- is honestly NULL rather than fabricated.
--
-- CORRECTED (this migration's original comment here was wrong and is
-- replaced, not merely amended, per the mechanism actually verified):
-- metaldocs.iam_group_roles DOES carry a composite PRIMARY KEY
-- (group_id, role) -- iam_group_roles_pkey, added via a separate
-- ALTER TABLE ... ADD CONSTRAINT in db/baseline/0001_current_schema.sql
-- (the CREATE TABLE's own column list does not show it, which is what the
-- original comment missed: pg_dump-style baselines add primary keys as a
-- later ALTER TABLE, not inline). That PK already makes a duplicate
-- (group_id, role) source row impossible, so this SELECT DISTINCT can never
-- actually collapse anything through this join: two source rows can only
-- differ by group_id or role, and every column this SELECT emits for a
-- given (group_id, role) pair is a deterministic function of group_id
-- (g.tenant_id, g.created_at) or a literal, so it cannot produce two
-- distinct output rows that further collapse either. DISTINCT is kept as
-- defense-in-depth against a future migration weakening or dropping that
-- PK, not because a real duplicate hazard exists today -- if that ever
-- changes, this comment (not a guess) is the reason the safety net is here.
INSERT INTO metaldocs.capability_bindings
    (tenant_id, subject_kind, subject_user_id, subject_group_id, role_code,
     scope_kind, scope_ref, effective_from, effective_to, granted_by, revoked_by, source_relation)
SELECT DISTINCT
    g.tenant_id, 'group', NULL::text, gr.group_id, gr.role,
    'tenant', NULL::text, g.created_at, NULL::timestamptz, NULL::text, NULL::text, 'iam_group_roles'
FROM metaldocs.iam_group_roles gr
JOIN metaldocs.iam_groups g ON g.id = gr.group_id
WHERE NOT EXISTS (
    SELECT 1 FROM metaldocs.capability_bindings WHERE source_relation = 'iam_group_roles'
);

-- ── schema_migrations ledger ────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0318', 'ADR 0092 D1 / issue #89 A8.1: metaldocs.roles catalog + metaldocs.capability_bindings relation (schema, constraints, indexes, RLS) and a reversible backfill from iam_user_roles, user_process_areas, and iam_group_roles. Old relations untouched and remain the live read path; nothing reads capability_bindings yet.')
ON CONFLICT (version) DO NOTHING;

COMMIT;
