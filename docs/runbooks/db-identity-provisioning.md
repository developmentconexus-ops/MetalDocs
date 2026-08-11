# Runbook — Database identity provisioning (three-identity split)

**Binary:** `apps/dbprovision/cmd/metaldocs-dbprovision`
**Owner:** ops / whoever operates the compose stack or `scripts/start-api.ps1`
**Context:** A6.1 re-cut (issue #88). `metaldocs-api`/`worker`/`jobs` boot-fatal
via `pgdb.AssertSafeIdentity` if the DB identity they connect as is
`SUPERUSER` or `BYPASSRLS` (both make RLS and the REVOKE-based audit_events
hardening inert). That gate only works if something else has already created
a safe, DML-only identity for them to use, and that something can no longer be
the app binaries themselves — they are the thing being gated. This runbook
documents the one-shot step that does that job, and why it moved out of
application startup.

## What changed operationally

Before A6.1, `metaldocs-api` applied `db/grants/*.sql` (role creation +
grants) and ran forward migrations itself, at every startup, over the same
pool it then served requests from. That created a deadlock once the identity
gate landed: the gate refuses to build API dependencies under an unsafe
identity, but the code path that *creates* the safe identity ran inside
`BuildAPIDependencies`, after the gate. An existing (already-provisioned)
volume could self-heal via `docker-entrypoint-initdb.d`, but that only fires
on a *fresh* Postgres data volume — an existing volume had no path to ever
gain the safe role.

Provisioning is now a separate one-shot binary, `metaldocs-dbprovision`, that:

1. Connects as the **bootstrap superuser** (`metaldocs_app` in dev/compose) —
   never through `bootstrap.Build*Dependencies`, never through the code path
   `AssertSafeIdentity` gates.
2. Applies `db/prerequisites/*.sql` and `db/grants/*.sql` (idempotent,
   guarded — see the header comments in `db/grants/0000_identity_roles.sql`
   and `db/grants/0001_role_grants.sql`), which creates/repairs the three
   identities described below; rotates `metaldocs_runtime`'s password to
   `METALDOCS_RUNTIME_DB_PASSWORD` (same var api/worker/jobs authenticate
   with — see `.env.example`); and, if `METALDOCS_JOBS_RIVER_SCHEMA` names a
   schema other than `public`, creates/re-owns that schema for
   `metaldocs_owner` (the two `db/grants` files only ever cover
   `public`/`metaldocs`, since a schema named only by a runtime env var can't
   be baked into hand-authored SQL).
3. Opens a SEPARATE connection pool pinned to `metaldocs_owner`
   (`postgres.OpenAsRole` — every physical connection it hands out issues
   `SET ROLE metaldocs_owner` before it is usable, and re-asserts that on
   every pooled reuse) and applies forward migrations
   (`internal/platform/migrate.Apply`) plus the River queue schema migration
   over that pool — DDL never runs over the identity that later serves
   requests, and never risks silently falling back to the bootstrap
   superuser across a pool reconnect (see `postgres.OpenAsRole`'s doc
   comment for why a single shared `*sql.DB` plus one `SET ROLE` statement
   does not actually guarantee that). For a custom River schema, grants
   `metaldocs_runtime` DML access on it once River's migrator has created
   its tables.
4. Exits. It does not open a long-lived pool and is not itself subject to
   `AssertSafeIdentity`.

`metaldocs-api`, `metaldocs-worker`, and `metaldocs-jobs` no longer create
roles, grant privileges, or run migrations. They only ever connect as
`metaldocs_runtime`.

## Identity map

| Identity | Attributes | Used by | Purpose |
|---|---|---|---|
| bootstrap superuser (`metaldocs_app` in dev) | `SUPERUSER` | `metaldocs-dbprovision` only, once per deploy | Create roles/extensions; the only identity allowed to be unsafe |
| `metaldocs_owner` | `NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT NOLOGIN` | `metaldocs-dbprovision` (via `SET ROLE`), never connects directly | Owns the `metaldocs`/`public` schemas and every table; runs DDL |
| `metaldocs_runtime` | `NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT LOGIN` | `metaldocs-api`, `metaldocs-worker`, `metaldocs-jobs` | DML only, RLS-subject; the only identity `AssertSafeIdentity` permits |

`metaldocs_owner` is deliberately `NOLOGIN`: nothing can connect *as* it
directly, only reach it via `SET ROLE` from a superuser session, which keeps
the DDL-capable identity out of reach of anything driven by request traffic.

`metaldocs_runtime` never owns a table. Postgres RLS does not apply to a
table's owner unless `FORCE ROW LEVEL SECURITY` is set; since the serving
identity must always be RLS-subject, it must never be an owner. Ownership
lives entirely with `metaldocs_owner`, established in
`db/grants/0000_identity_roles.sql` by `ALTER SCHEMA ... OWNER TO
metaldocs_owner` plus a scoped, per-object-kind `ALTER TABLE/SEQUENCE/
FUNCTION/PROCEDURE/TYPE ... OWNER TO metaldocs_owner` loop over
`pg_class`/`pg_proc`/`pg_type` — **not** a blanket `REASSIGN OWNED BY
CURRENT_USER`. That statement was tried and reverted: `CURRENT_USER`, under
the bootstrap superuser this file requires, is the actual cluster `initdb`
role, which also structurally "owns" `pg_catalog`/`information_schema`.
`REASSIGN OWNED BY` has no per-schema filter, so it always tries to move
those two system schemas too and always fails with `SQLSTATE 2BP01`
("cannot reassign ownership of objects owned by role ... because they are
required by the database system"), rolling back the entire transaction —
including the `CREATE ROLE` statements — on every run. The scoped loop only
ever touches `public`/`metaldocs` (plus, as of the custom-River-schema fix
above, whatever `METALDOCS_JOBS_RIVER_SCHEMA` names), so it never hits
`pg_catalog`/`information_schema` and cannot hit `2BP01`. See that file's own
header comment for the full empirical account.

## When to run

Before any of `metaldocs-api`, `metaldocs-worker`, `metaldocs-jobs` start —
on every boot of the stack, not just the first. It is idempotent: on an
already-provisioned volume every stage is a guarded no-op except forward
migrations, which are ledgered (`schema_migrations`) and only apply what is
new.

## Procedure

**Compose:** the `db-provision` service in
`deploy/compose/docker-compose.yml` runs automatically. `api`, `worker`, and
`jobs` declare `depends_on: { db-provision: { condition:
service_completed_successfully } }`, so compose will not start them until it
exits 0.

**`scripts/start-api.ps1`** (runs compiled binaries directly, not through
compose): the script builds/runs `metaldocs-dbprovision.exe` as a step before
launching `metaldocs-api.exe`, temporarily swapping `PGUSER`/`PGPASSWORD` to
the bootstrap superuser (`POSTGRES_USER`/`POSTGRES_PASSWORD`) for that one
step only, then restoring the runtime credentials for the app binary itself.

Manual/ad-hoc invocation (e.g. an environment without compose):

```
$env:PGUSER = $env:POSTGRES_USER
$env:PGPASSWORD = $env:POSTGRES_PASSWORD
./metaldocs-dbprovision.exe
```

## Verification

Role attributes and ownership are the load-bearing facts — confirm them with
plain SQL against the target database:

```sql
SELECT rolname, rolsuper, rolbypassrls, rolcreaterole, rolcreatedb, rolcanlogin
FROM pg_roles WHERE rolname IN ('metaldocs_owner', 'metaldocs_runtime');
-- metaldocs_owner:   f, f, f, f, f (NOLOGIN)
-- metaldocs_runtime: f, f, f, f, t

SELECT tableowner, count(*) FROM pg_tables WHERE schemaname = 'metaldocs' GROUP BY tableowner;
-- must show metaldocs_owner only; metaldocs_runtime must never appear here
```

Boot logs: `metaldocs-dbprovision` logs each grants-stage file and each
applied migration version, then `"db provisioning complete"`. The app
binaries log nothing about grants/migrations; if `AssertSafeIdentity` fires,
they log a single fatal `"refusing to boot: db identity ... is unsafe"` and
exit non-zero before opening any listener.

## Rollback / troubleshooting

- **Gate fires on a service that should be safe:** the connecting role has
  `SUPERUSER` or `BYPASSRLS` set. Fix the role's attributes
  (`ALTER ROLE ... NOSUPERUSER NOBYPASSRLS`) or point the service's
  `PGUSER`/`PGPASSWORD` at `metaldocs_runtime`. Do not add an escape hatch to
  the gate itself.
- **`db-provision` fails on a fresh volume with `schema "metaldocs" does not
  exist`:** the baseline schema (`db/baseline/0001_current_schema.sql`) has
  not been applied yet. In compose this runs via
  `docker-entrypoint-initdb.d` before `db-provision` starts; outside compose,
  run `scripts/dev-bootstrap-baseline.ps1` (or the equivalent baseline load)
  first.
- **Re-running is always safe.** Every stage `db-provision` runs is either
  guarded (`IF NOT EXISTS`/`IF EXISTS` checks in the grants SQL) or ledgered
  (`schema_migrations`); there is no destructive step.
