# Feature F7.4 — Plan: RLS-truth sweep

> Input: `spec.md` (approved). Engine: `superpowers:subagent-driven-development` (TDD; sonnet impl+review, haiku mechanical, never fable; main orchestrates+reviews+commits). Migrations expand-only, guarded. Commits local.

## Runtime-truth anchors (verified this session)
- Next migration numbers: **0284**, **0285** (latest on disk = `0283_tripwire_delete_return_old.sql`; 0280 already skipped historically).
- Role/grant idiom: guarded `DO $$ … IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname='…') …$$` (mirrors `0266`). `metaldocs_app` has `rolcreaterole=t` (restricted-role creation permitted; non-elevating).
- RLS policy idiom (verbatim from `0278`/`0258`): `(NULLIF(current_setting('metaldocs.tenant_id',true),'') IS NULL) OR (<col> = (NULLIF(...))::uuid)`. Null-GUC branch = pass-all (janitor/bootstrap escape hatch, ADR 0027).
- `public.approval_signoffs` — schema `public`, `actor_tenant_id uuid NOT NULL`, 5 BEFORE-INSERT triggers (cap-asserted, SoD, eligibility, tenant-consistent, immutable-on-update). Direct row seeding is trigger-heavy → row-level leak/block proof runs on a low-trigger FORCE-RLS table (`metaldocs.tenant_lifecycle_jobs`, 0278); `approval_signoffs` §4.4 proven at catalog level (FORCE + policy on `actor_tenant_id`).
- Harness: `tests/integration/testdb/db.go` — `Open()` clones per-test DB from template (owner `metaldocs_app`); `openDBWithDatabase(dsn,dbName)` = `pgx.ParseConfig` → override `cfg.Database` → `stdlib.OpenDB`. RLS applies to a non-owner under plain `ENABLE` (FORCE only needed for the owner) → `metaldocs_ci` genuinely filters.
- Lint host: `scripts/api-lint/` — mirror `async_tenant_seed_rule.go` (AST scan of async handler roots + `async-tenant-tables.txt` table set + allowlist `.txt`, registered in `RunCodeRules`).

## Tasks (ordered)

### Task A — migration 0284: `metaldocs_ci` non-owner role + grants  *(sonnet)*
- `db/migrations/0284_ci_rls_role.sql` (+ `_down.sql`): guarded `DO` block —
  `CREATE ROLE metaldocs_ci NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE LOGIN PASSWORD '<dev>'` only `IF NOT EXISTS`. Then:
  `GRANT USAGE ON SCHEMA metaldocs, public TO metaldocs_ci;`
  `GRANT SELECT,INSERT,UPDATE,DELETE ON ALL TABLES IN SCHEMA metaldocs, public TO metaldocs_ci;`
  `GRANT USAGE,SELECT ON ALL SEQUENCES IN SCHEMA metaldocs, public TO metaldocs_ci;`
  `ALTER DEFAULT PRIVILEGES FOR ROLE metaldocs_app IN SCHEMA metaldocs, public GRANT SELECT,INSERT,UPDATE,DELETE ON TABLES TO metaldocs_ci;` (+ sequences) — future tables inherit.
  Header comment: dev password is a **non-secret fixture** for a DML-only non-super role; prod overrides via `ALTER ROLE … PASSWORD` from a secret (never `.env`-committed). Grants baked into template → inherited by every `CREATE DATABASE … TEMPLATE` clone. Append `schema_migrations` version row (mirror 0266's INSERT).
- **Test:** none standalone (proven live by Task D's catalog assertions + role-attribute query). Bootstrap-apply is exercised when the integration template rebuilds.

### Task B — migration 0285: `approval_signoffs` FORCE RLS + policy  *(sonnet)*
- `db/migrations/0285_approval_signoffs_rls.sql` (+ `_down.sql`):
  `ALTER TABLE public.approval_signoffs ENABLE ROW LEVEL SECURITY;`
  `ALTER TABLE ONLY public.approval_signoffs FORCE ROW LEVEL SECURITY;`
  `DROP POLICY IF EXISTS tenant_isolation ON public.approval_signoffs;`
  `CREATE POLICY tenant_isolation ON public.approval_signoffs USING ((NULLIF(current_setting('metaldocs.tenant_id',true),'') IS NULL) OR (actor_tenant_id = (NULLIF(current_setting('metaldocs.tenant_id',true),''))::uuid));`
  Down: `DROP POLICY` + `NO FORCE` + `DISABLE`. `schema_migrations` row.
- **Test:** catalog assertion in Task D (`relforcerowsecurity`=t + `pg_policy` USING references `actor_tenant_id`).

### Task C — testdb CI-role seam  *(sonnet)*
- `tests/integration/testdb/ci_role.go` (build tag `integration`): `OpenAsCIRole(t *testing.T, dbName string) *sql.DB` — `pgx.ParseConfig(DSN(t))`, set `cfg.User="metaldocs_ci"`, `cfg.Password = os.Getenv("METALDOCS_CI_DB_PASSWORD")` else dev default matching the migration, `cfg.Database=dbName`; `stdlib.OpenDB`; ping; `t.Cleanup(close)`. Doc: setup/seed stays on the owner conn; only isolation-proof reads go through this.
- **Test:** exercised by Task D.

### Task D — RLS-truth proof test (§4.5 bar)  *(sonnet, TDD — write RED first)*
- `tests/integration/security/rls_truth_test.go` (or `tests/integration/iam/`), `TestRLSTruth_NonOwnerRoleEnforcesIsolation`:
  1. `db, dbName := testdb.Open(t)` (owner). Seed two `tenants` rows (A,B) + one `tenant_lifecycle_jobs` row each (owner conn, no GUC — bypass). 
  2. Role-attribute assert: `SELECT rolsuper,rolbypassrls FROM pg_roles WHERE rolname='metaldocs_ci'` = f,f; owns 0 tables (`pg_class.relowner`→ not `metaldocs_ci`).
  3. `ci := testdb.OpenAsCIRole(t, dbName)`. In one tx each (`SET LOCAL metaldocs.tenant_id`):
     - (a) positive: GUC=A → job rows = {A only}.
     - (b) isolation: GUC=B → A's job row count = **0**.
     - (c) bypass contrast: **owner** conn, GUC=B, same query → A's row still returned (documents pre-F7.4 false-green). 
     - (d) null-GUC pin: `ci`, no GUC → both rows (deliberate escape hatch; asserted so idiom is pinned).
  4. `approval_signoffs` §4.4: catalog assert FORCE+policy-on-`actor_tenant_id`; plus a CI-role wrong-GUC SELECT returns 0 (empty table is fine — proves policy is queryable/effective under the non-owner role, no trigger-heavy seed needed).
- Fail-first: run before 0284/0285 land (or with policy absent) → RED; then green.

### Task E — sole-RLS async-read lint  *(sonnet)*
- `scripts/api-lint/sole_rls_rule.go` + `_test.go` + `sole-rls-allowlist.txt`; register in `RunCodeRules` beside `async_tenant_seed_rule`. Scope (non-contradictory with M3): AST-scan the **same async handler roots** as `async_tenant_seed_rule`; flag a `Query/QueryContext/QueryRow*` whose **string-literal SELECT** targets a table in `async-tenant-tables.txt` and whose SQL text contains **no** `tenant_id`/`actor_tenant_id` token, unless allowlisted. Rationale in header: M3 covered async *writes* (must seed GUC); this covers async *reads* (belt-and-suspenders explicit predicate, the exact class M6 F6.4 fixed on review-due ports — null-GUC branch means an unseeded async read would pass-all without a predicate).
- **Negative proof:** `testdata/` synthetic sole-RLS async SELECT → rule reports it; add-predicate → clean. `code_rules`-style unit test.
- Run `api-lint` over the tree → 0 violations (fix any surfaced real async read with an explicit predicate, M6 idiom; enumerate in evidence).

### Task F — evidence.md  *(main)*
- Census enumeration (async-read surface the lint covers + result), role-attribute proof, negative+positive capture, **both §4.2 deviations disclosed**, fixture-vs-real labels, review disposition, bounded defers (full-suite-under-CI-role not a blanket harness swap — setup needs owner; census = lint static coverage + live proof; literal ownership-reassignment + deny-by-default policy model = named follow-ons).

## Ordering & gates
A+B (migrations, parallel) → template rebuild picks them up → C (seam) → D (proof, depends A/B/C) → E (lint, independent, parallel with D) → `go build ./...` + `go vet -tags=integration ./tests/...` + api-lint 0 → real-DB run of D under loaded `.env` → F. Then milestone close.

## Non-goals (from spec)
No `metaldocs_app` NOSUPERUSER flip; no 63-table ownership reassignment; no RLS-policy-model redesign (null-GUC branch stays); no new route/cap/contract; no speculative predicate retrofit beyond census-surfaced.
