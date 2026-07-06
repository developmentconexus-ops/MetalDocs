# Feature F7.4 — Spec: RLS-truth sweep

> **Milestone:** 7 — Tenant Lifecycle Kernel · **Folder:** `f7.4-rls-truth-sweep`
> **Status:** Approved (pre-code) 2026-07-05
> **Design of record:** `../validation-contract.md` §4 (binding) + §0.4 (DB-role facts) + mission T-003 / Finding 15 (backend-canon). Carried from M6 as F7.4.

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | All shape decisions locked upstream? | Yes — contract §4.1–4.5 fully specifies role attributes, non-owner requirement, census+predicate fixes, `approval_signoffs` disposition, negative+positive proof. No new operator questions. |
| 2 | Runtime-truth delta 1 — role mechanism | §4.2 says "table ownership is **reassigned** to a distinct owner role (e.g. metaldocs_owner/postgres) while the app role connects as a **non-owner**." Runtime truth: dev `metaldocs_app` is SUPERUSER+BYPASSRLS and **owns all 63 tables**; it is also the role the API/migrations/janitors/bootstrap run as. Making it NOSUPERUSER or reassigning all 63 tables' ownership would break dev bootstrap (CREATE DATABASE/TEMPLATE, DDL, GUC bypass seeding) — a schema-wide owner migration = **HS-2 risk**. **Decision (deviation-from-literal, same end-state):** introduce a dedicated `metaldocs_ci` role — NOSUPERUSER + NOBYPASSRLS, **owns nothing** (non-owner by construction), DML-only grants. The integration suite's RLS proof connects as `metaldocs_ci`; `metaldocs_app` remains the owner/bootstrap identity (mirroring prod, where a separate migration identity owns and the app role is a non-owner — §0.4 "all Owner: -"). This satisfies §4.2's actual requirement (app-connection-role ≠ owner, NOBYPASSRLS) without a reassignment migration. Prod-parity: `metaldocs_ci` == the prod app-role constraint. Flagged for HS-1 operator review. |
| 3 | Runtime-truth delta 2 — `approval_signoffs` column + schema | Table lives in schema **`public`** (not `metaldocs`); keyed by `actor_tenant_id uuid` (not `tenant_id`) — the reason it was missed by the FORCE-RLS census. Policy added on `actor_tenant_id` (§4.4 default expectation), using the canonical null-GUC-bypass idiom (below). No cross-tenant co-sign semantics found (single-tenant signoff), so a strict `tenant_isolation` policy is correct — no ADR carve-out needed. |
| 4 | Runtime-truth delta 3 — policy semantics (reconciles §4.2 + §4.5 "no-GUC → 0 rows") | **CONTRACT-VS-RUNTIME CONFLICT — surfaced, not patched.** §4.2/§4.5 state the check as "no tenant GUC → 0 rows (policy active)." Runtime truth: every tenant `tenant_isolation` policy is `(NULLIF(current_setting('metaldocs.tenant_id',true),'') IS NULL) OR (<col> = (NULLIF(...))::uuid)` — the **unset/empty GUC branch deliberately passes ALL rows**. This null-GUC escape hatch is **ratified M3 design (ADR 0027 amendment)**: the leader-elected janitors (audit-integrity-validator, lease-reaper, idempotency-janitor) and bootstrap scan cross-tenant with **no** tenant GUC; M3's TxRunner auto-seeds the GUC on every app request so real reads are always scoped. **Removing the null branch from all 34 policies to make "no-GUC → 0 rows" literally true would break those janitors = a schema-wide RLS-model redesign, outside F7.4's boundary (HS-2).** F7.4 therefore proves the *actual security property* — cross-tenant **isolation** — via the strictly-stronger **wrong-GUC-blocks + bypass-contrast**: GUC=tenantB → tenantA's row = **0** under the NOBYPASSRLS CI role, while the *same wrong-GUC query* under BYPASSRLS `metaldocs_app` still returns tenantA's row (the pre-F7.4 false-green). The null-GUC=all-rows branch is asserted/pinned (case d) as the deliberate escape hatch, not treated as a leak. **Ratify at HS-1.** |

## Consumer contract (FIRST)

- **Consumers:**
  1. **A CI/prod operator** — the integration suite runs green under a role that is NOSUPERUSER + NOBYPASSRLS + non-owner. RLS is genuinely enforced, so a passing suite is real proof of tenant isolation (not a superuser false-green).
  2. **A future engineer** — a lint fails the build if a new tenant-scoped read/write relies **solely** on RLS (no explicit `tenant_id`/`actor_tenant_id` predicate). Correct-by-construction: the guard, not vigilance.
  3. **A data subject / auditor** — a tenant-scoped query **leaks another tenant's row under the wrong GUC and is blocked (0 rows) under the right GUC**, proven for real under the non-bypassing role — RLS is active, not bypassed.
- **Source of truth:** `../validation-contract.md` §4; the running Postgres role catalog + `pg_policy`.

## What this feature implements

1. **`metaldocs_ci` role (prod-parity CI identity).** A migration (`db/migrations/NNNN_*.sql`, expand-only, idempotent `DO`-guarded `CREATE ROLE IF NOT EXISTS`) creates `metaldocs_ci` NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE LOGIN, and grants it: `USAGE` on schemas `metaldocs`/`public`, `SELECT/INSERT/UPDATE/DELETE` on all tenant-scoped tables + `USAGE` on sequences, plus `ALTER DEFAULT PRIVILEGES` so template-cloned per-test DBs inherit the grants. Password from env (`METALDOCS_CI_DB_PASSWORD`, default dev value documented in `.env.example`; never printed). Roles are cluster-global; grants are per-DB and inherited by `CREATE DATABASE … TEMPLATE`.
2. **`approval_signoffs` RLS closure.** Migration `0284`: `ALTER TABLE public.approval_signoffs ENABLE ROW LEVEL SECURITY; ALTER TABLE ONLY public.approval_signoffs FORCE ROW LEVEL SECURITY;` + `CREATE POLICY tenant_isolation ON public.approval_signoffs USING ((NULLIF(current_setting('metaldocs.tenant_id',true),'') IS NULL) OR (actor_tenant_id = (NULLIF(current_setting('metaldocs.tenant_id',true),''))::uuid))` — **verbatim** idiom from `0278`/`0258` (null-GUC-bypass branch preserved for parity). Note `public` schema, `actor_tenant_id` column. Down-migration (`DROP POLICY` + `NO FORCE`/`DISABLE`) provided.
3. **testdb harness: non-owner connection seam.** `tests/integration/testdb`: add `OpenAsCIRole(t, dbName) *sql.DB` (connects to the per-test DB as `metaldocs_ci`) and ensure the template DB grants `metaldocs_ci` DML (so clones inherit). Setup/seed paths stay on `metaldocs_app` (owner) — only the isolation-proof reads/writes go through the CI role.
4. **Negative+positive RLS proof (the §4.5 bar, corrected for the null-GUC-bypass idiom).** New integration test `TestRLSTruth_NonOwnerRoleEnforcesIsolation`: seed two tenants' rows in a FORCE-RLS table (owner conn); under the `metaldocs_ci` conn — (a) **positive:** GUC=tenantA → only A's rows; (b) **negative/isolation:** GUC=tenantB → A's row **blocked (0)**, only B's rows; (c) **bypass false-green contrast (the crux):** the *same wrong-GUC (=tenantB) query* on the `metaldocs_app` (BYPASSRLS) conn still returns A's row — proving the pre-F7.4 suite was false-green; (d) **null-GUC escape-hatch documented:** no GUC under the CI role returns all rows (the deliberate janitor/bootstrap branch) — asserted so the idiom is pinned, not mistaken for a leak. Also assert `public.approval_signoffs` isolates identically under the CI role (wrong-GUC blocks).
5. **Census sweep + sole-RLS lint.** Run tenant-scoped integration reads/writes under the CI role; enumerate every query whose isolation assertion flips RED and fix each with an explicit `tenant_id`/`actor_tenant_id` predicate (M6 F6.4 idiom). Add an api-lint guard (`scripts/api-lint`, M2/M3 pattern) that flags a new tenant-scoped read/write lacking an explicit tenant predicate; negative proof: a synthetic sole-RLS query fails the lint. The set of surfaced queries is enumerated in `evidence.md` (the suite IS the census).

## Non-goals (mandatory)

- **No** flipping `metaldocs_app` itself to NOSUPERUSER, and **no** schema-wide ownership reassignment of the 63 tables (HS-2 — would break dev bootstrap; the dedicated `metaldocs_ci` non-owner role achieves the same isolation property).
- **No** rewrite of the RLS policy model (the 18+`approval_signoffs` `tenant_isolation` policies are the model; F7.4 makes them *effective under test*, it does not redesign them).
- **No** new capability, route, or contract change (this is infra/test-truth, not API surface).
- **No** retrofitting explicit predicates into queries the census does **not** surface (correct-by-construction from the census, not a speculative grep-and-touch).

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof | Real vs fixture |
|---|---|---|
| `metaldocs_ci` is NOSUPERUSER+NOBYPASSRLS+non-owner | `SELECT rolsuper,rolbypassrls` = f,f; owns 0 tables (`pg_class.relowner`) | real (Postgres) |
| Positive: right-GUC query returns exactly the tenant's rows | `TestRLSTruth_NonOwnerRoleEnforcesIsolation` case (a) | real |
| Negative/isolation: wrong-GUC query blocked (0 rows) under CI role | case (b) | real |
| Bypass false-green: same wrong-GUC query leaks under BYPASSRLS owner | case (c) — A's row returned under `metaldocs_app` | real |
| Null-GUC escape-hatch pinned (all rows, deliberate) | case (d) | real |
| `approval_signoffs` FORCE RLS + `tenant_isolation` policy on `actor_tenant_id` | `pg_class.relforcerowsecurity`=t + `pg_policy`; CI-role wrong-GUC-blocks assert | real |
| Every flip-surfaced query carries an explicit tenant predicate | census enumerated in `evidence.md`; each fix linked | real |
| Sole-RLS lint blocks a new bare tenant-scoped query | `scripts/api-lint` guard + synthetic-query negative test; `go run ./scripts/api-lint -strict …` 0 violations on the tree | real |
| Full targeted integration suite green under the new role path | `go test -tags=integration ./tests/integration/...` (isolation proofs via CI role) | real |
| No regression across M0–M6 | full `go test ./...` + `-tags=integration` targeted suites green | real |
| Live/system-runnable unaffected | `.\scripts\check-system-runnable.ps1`; API still boots as `metaldocs_app` | real |

TDD: failing proof/lint first, then green. Evidence per contract §6.3 (census table, negative+positive capture, deviation disclosure, fixture-vs-real labeled).

## ADR needed? / Deviations to ratify at HS-1

Two documented deviations-from-literal-contract, both **same-security-property, lower-blast-radius**, surfaced (not patched around):

1. **§4.2 role mechanism** — dedicated non-owner `metaldocs_ci` role vs literal ownership-reassignment of all 63 tables. Upholds the invariant (app-connection-role ≠ owner, NOBYPASSRLS); avoids a dev-bootstrap-breaking owner migration (HS-2). Interview delta 1.
2. **§4.2/§4.5 "no-GUC → 0 rows" check** — reconciled to **wrong-GUC-blocks + bypass-contrast** because the ratified M3 null-GUC-bypass policy idiom (ADR 0027) makes literal "no-GUC → 0 rows" false-by-design and un-achievable without a janitor-breaking 34-policy redesign (HS-2). The stronger isolation property is proven. Interview delta 3.

Neither is a MUST-deviation from the *target architecture* (both uphold "app connects as a non-owner under NOBYPASSRLS; RLS genuinely filters tenant reads"). **No new ADR authored** unless HS-1 directs one; if the operator requires literal ownership-reassignment or a deny-by-default policy model, each becomes its own follow-on milestone with its own migration + janitor-impact analysis.
