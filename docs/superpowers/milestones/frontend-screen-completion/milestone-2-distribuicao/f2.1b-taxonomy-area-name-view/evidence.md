# Feature F2.1b — Evidence

> **Milestone:** 2 — Distribuição coverage-scope  ·  **Feature:** `f2.1b-taxonomy-area-name-view`  ·  **Closed:** 2026-06-22
> **Contract:** `spec.md` (consumer contract + Validation Gate this proves against).
> A feature is closed only when every row below is filled with **real, honestly-labeled** output —
> not "done" / "green" / "looks good", and not a fixture passed off as the real provider.

## What was implemented

- **Migration `0246`** — `db/migrations/0246_taxonomy_process_area_name_view.sql` publishes `metaldocs.v_process_area_name (tenant_id uuid, area_code text, area_name text)` as a 1:1 `CREATE VIEW` over `metaldocs.document_process_areas` (renames: `code → area_code`, `name → area_name`). No `is_active`/`archived_at` filter, matching the existing `AreaCatalogReader` port semantic. Idempotent via the `schema_migrations` ledger (`ON CONFLICT DO NOTHING`). No Go code.
- **Integration test** — `internal/modules/taxonomy/infrastructure/v_process_area_name_integration_test.go` asserts view shape + 1:1 per-tenant projection (3 seeded areas → 3 view rows) + cross-tenant isolation. Build tag `integration`.
- **ADR-0041** — `wiki/decisions/0041-taxonomy-process-area-name-view.md` records the decision (D1 minimal view, D2 shape, D3 no-filter semantic, D4 no extra columns, D5 security posture).
- **ADR-0039 inventory** — `wiki/decisions/0039-cross-module-base-table-read-boundary.md` updated with `v_process_area_name` inventory bullet (line 16); `Last verified` flipped to 2026-06-21.

Producer matches consumer contract in `spec.md`: view shape is exactly `(tenant_id uuid, area_code text, area_name text)`, 1:1 projection, ADR-0041 recorded, ADR-0039 inventory updated, `hgcrossmodule` = 0 H-G.

**Commits (chronological):**
- `4db2c5d5 test(M2/F2.1b): failing test for v_process_area_name — shape + 1:1 + isolation`
- `33c82cfc feat(M2/F2.1b): publish metaldocs.v_process_area_name view (ADR-0041)`
- `c0b57bbd test(M2/F2.1b): relax v_process_area_name shape nullability to match PG`
- `2aff58e7 docs(adr): ADR-0041 v_process_area_name + ADR-0039 inventory (M2/F2.1b)`
- `2bcbd892 docs(adr): flip ADR-0039 front-matter Last verified to 2026-06-21 (M2/F2.1b)`
- `08486ad3 docs(M2/F2.1b): commit feature plan (writing-plans output)`

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| Migration ledger row exists exactly once | `psql -c "SELECT version FROM public.schema_migrations WHERE version='0246'"` | ` version` / `---------` / ` 0246` / `(1 linha)` — exactly one row | real |
| View shape — 3 columns, correct types | `psql -c "\d+ metaldocs.v_process_area_name"` | `tenant_id uuid`, `area_code text`, `area_name text` — 3 columns; `information_schema.columns` confirms `is_nullable=YES` for all three (PG view nullability behaviour; test file relaxed in `c0b57bbd`) | real |
| `information_schema` column-level shape | `SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_schema='metaldocs' AND table_name='v_process_area_name' ORDER BY ordinal_position` | `tenant_id uuid YES`, `area_code text YES`, `area_name text YES` | real |
| 1:1 per-tenant projection — `TestProcessAreaName_OneToOneProjection_PerTenant` | `go test -tags=integration -run "TestProcessAreaName" ./internal/modules/taxonomy/infrastructure/... -v` | `--- PASS: TestProcessAreaName_OneToOneProjection_PerTenant (218.25s)` | real (live PG, isolated DB clone) |
| Cross-tenant isolation — `TestProcessAreaName_CrossTenantIsolation` | same command | `--- PASS: TestProcessAreaName_CrossTenantIsolation (34.65s)` | real |
| View shape test — `TestProcessAreaName_ViewShape` | same command | `--- PASS: TestProcessAreaName_ViewShape (2.41s)` | real |
| Full integration test suite result | same command | `PASS` / `ok metaldocs/internal/modules/taxonomy/infrastructure 257.898s` / `go test exit: 0` | real |
| Taxonomy runtime + sibling F2.1a migration untouched | `git diff 6ff506d5 -- internal/modules/taxonomy ':(exclude)internal/modules/taxonomy/infrastructure/v_process_area_name_integration_test.go'` and `git diff 6ff506d5 -- db/migrations/0245_cd_obligated_readers_view.sql` | Both diffs: empty (exit 0) — no runtime Go code changed, F2.1a migration untouched | real |
| ADR-0041 recorded | `ls wiki/decisions/0041-taxonomy-process-area-name-view.md` | `0041-taxonomy-process-area-name-view.md  21/06/2026 23:43:51` — file exists | real |
| ADR-0039 inventory updated | `grep -n "v_process_area_name" wiki/decisions/0039-cross-module-base-table-read-boundary.md` | `Line 16: > - db/migrations/0246_taxonomy_process_area_name_view.sql — ...` — ≥1 hit | real |
| `hgcrossmodule` analyzer — 0 H-G findings | `go run ./tools/cilint/...` | `cilint exit: 0` — no output, no H-G findings | real |
| `go build ./...` | `go build ./...` | `build exit: 0` | real |
| `go vet ./...` | `go vet ./...` | `vet exit: 0` | real |
| `go test ./...` (non-integration regression) | `go test ./...` | All packages green; `test exit: 0` — full output below | real |

**`go test ./...` full output (last run, non-integration):**
```
ok  	metaldocs/apps/api/cmd/metaldocs-api	24.051s
ok  	metaldocs/apps/api/internal/wiring	3.739s
ok  	metaldocs/apps/worker/cmd/metaldocs-worker	23.303s
ok  	metaldocs/db/migrations	(cached)
ok  	metaldocs/internal/modules/audit/application	(cached)
ok  	metaldocs/internal/modules/audit/delivery/http	5.790s
ok  	metaldocs/internal/modules/audit/infrastructure/postgres	(cached)
ok  	metaldocs/internal/modules/auth/application	15.701s
ok  	metaldocs/internal/modules/auth/delivery/http	21.060s
ok  	metaldocs/internal/modules/auth/domain	(cached)
ok  	metaldocs/internal/modules/auth/infrastructure/memory	(cached)
ok  	metaldocs/internal/modules/auth/infrastructure/postgres	(cached)
ok  	metaldocs/internal/modules/controlleddocuments/application	4.105s
ok  	metaldocs/internal/modules/controlleddocuments/delivery/http	5.661s
ok  	metaldocs/internal/modules/controlleddocuments/domain	2.142s
ok  	metaldocs/internal/modules/controlleddocuments/infrastructure	2.643s
ok  	metaldocs/internal/modules/documents	5.457s
ok  	metaldocs/internal/modules/documents/application	4.433s
ok  	metaldocs/internal/modules/documents/approval/application	4.677s
ok  	metaldocs/internal/modules/documents/approval/domain	(cached)
ok  	metaldocs/internal/modules/documents/approval/http	5.730s
ok  	metaldocs/internal/modules/documents/approval/http/contracts	(cached)
ok  	metaldocs/internal/modules/documents/approval/infrastructure	3.097s
ok  	metaldocs/internal/modules/documents/approval/infrastructure/signature	(cached)
ok  	metaldocs/internal/modules/documents/approval/repository	(cached)
ok  	metaldocs/internal/modules/documents/delivery/http	5.639s
ok  	metaldocs/internal/modules/documents/domain	2.038s
ok  	metaldocs/internal/modules/documents/repository	4.207s
ok  	metaldocs/internal/modules/iam/application	(cached)
ok  	metaldocs/internal/modules/iam/authz	(cached)
ok  	metaldocs/internal/modules/iam/delivery/http	5.539s
ok  	metaldocs/internal/modules/iam/domain	(cached)
ok  	metaldocs/internal/modules/iam/infrastructure/memory	(cached)
ok  	metaldocs/internal/modules/iam/infrastructure/postgres	4.032s
ok  	metaldocs/internal/modules/iam/presence	6.694s
ok  	metaldocs/internal/modules/jobs/audit_integrity_validator	(cached)
ok  	metaldocs/internal/modules/jobs/idempotency_janitor	(cached)
ok  	metaldocs/internal/modules/jobs/scheduler	(cached)
ok  	metaldocs/internal/modules/jobs/stuck_instance_watchdog	4.446s
ok  	metaldocs/internal/modules/render/fanout	(cached)
ok  	metaldocs/internal/modules/render/resolvers	(cached)
ok  	metaldocs/internal/modules/search/application	(cached)
ok  	metaldocs/internal/modules/search/delivery/http	3.928s
ok  	metaldocs/internal/modules/search/infrastructure/v2documents	4.361s
ok  	metaldocs/internal/modules/security/application	(cached)
ok  	metaldocs/internal/modules/taxonomy	5.328s
ok  	metaldocs/internal/modules/taxonomy/application	3.353s
ok  	metaldocs/internal/modules/taxonomy/delivery/http	(cached)
ok  	metaldocs/internal/modules/taxonomy/domain	2.003s
ok  	metaldocs/internal/modules/taxonomy/infrastructure	4.160s
ok  	metaldocs/internal/modules/templates/application	(cached)
ok  	metaldocs/internal/modules/templates/delivery/http	(cached)
ok  	metaldocs/internal/modules/templates/domain	(cached)
ok  	metaldocs/internal/modules/templates/infrastructure	(cached)
ok  	metaldocs/internal/modules/templates/repository	(cached)
ok  	metaldocs/internal/platform/authn	(cached)
ok  	metaldocs/internal/platform/bootstrap	10.603s
ok  	metaldocs/internal/platform/config	(cached)
ok  	metaldocs/internal/platform/db	(cached)
ok  	metaldocs/internal/platform/db/postgres	(cached)
ok  	metaldocs/internal/platform/docgenv2	3.247s
ok  	metaldocs/internal/platform/featureflags	(cached)
ok  	metaldocs/internal/platform/httpclient	(cached)
ok  	metaldocs/internal/platform/httpresponse	(cached)
ok  	metaldocs/internal/platform/middleware	2.100s
ok  	metaldocs/internal/platform/migrate	(cached)
ok  	metaldocs/internal/platform/objectstore	3.249s
ok  	metaldocs/internal/platform/observability	11.952s
ok  	metaldocs/internal/platform/pagination	(cached)
ok  	metaldocs/internal/platform/problem	(cached)
ok  	metaldocs/internal/platform/ratelimit	(cached)
ok  	metaldocs/internal/platform/render/gotenberg	(cached)
ok  	metaldocs/internal/platform/requesttrace	(cached)
ok  	metaldocs/internal/platform/security	(cached)
ok  	metaldocs/internal/platform/servicebus	(cached)
ok  	metaldocs/internal/platform/sqlescape	(cached)
ok  	metaldocs/internal/platform/tenant	(cached)
ok  	metaldocs/internal/platform/useragent	(cached)
ok  	metaldocs/internal/platform/worker	(cached)
ok  	metaldocs/scripts/api-lint	9.746s
ok  	metaldocs/tests/docx_v2	(cached)
ok  	metaldocs/tests/unit	11.575s
ok  	metaldocs/tests/unit/iam_memberships	2.484s
ok  	metaldocs/tests/unit/iam_people	2.485s
ok  	metaldocs/tools/cilint/internal/analyzers	4.859s
```

> Observable changes must be runtime-verified here by us — never deferred to the operator.
> Fixture-only proof is labeled as such; it is not end-to-end proof of the real provider. A
> suite-level "all green" without per-criterion mapping below is **not** acceptance.

## Acceptance vs spec Validation Gate

Restate this feature's Validation Gate from `spec.md` (which traces to its `milestone.md` row) and
mark each criterion met/not-met against the evidence above.

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Migration applies cleanly + idempotent (`SELECT version ... WHERE version='0246'` returns one row) | yes | Migration ledger row check above — `(1 linha)` with `0246` |
| View shape matches the contract (`\d+` shows exactly 3 columns with the declared types + nullability) | yes | `\d+` shows 3 columns; `information_schema` confirms `tenant_id uuid`, `area_code text`, `area_name text` (nullability is `YES` per PG view semantics — test file relaxed in `c0b57bbd` per accepted T1 finding) |
| 1:1 row count with the taxonomy base table per tenant (integration test seeds N areas, asserts count = N) | yes | `TestProcessAreaName_OneToOneProjection_PerTenant` PASS (seeds 3 areas, asserts 3 rows); `TestProcessAreaName_CrossTenantIsolation` PASS; `TestProcessAreaName_ViewShape` PASS |
| ADR-0041 recorded + ADR-0039 inventory updated | yes | `0041-taxonomy-process-area-name-view.md` exists; `grep` returns line 16 hit in ADR-0039 |
| `hgcrossmodule` analyzer green (`go run ./tools/cilint/...` = 0 H-G) | yes | `cilint exit: 0` — no output, no H-G findings |

## Review disposition

- **Spec-compliance review:** PASS at every task (T1–T3). Minor finding in T3 (outer `Last verified` date flip on ADR-0039 front-matter) already fixed in commit `2bcbd892` before gate run.
- **Code-quality review:** T1 finding (nullability `NO` vs `YES` in test expectations) accepted + fixed in T2 via commit `c0b57bbd` — PG views do not propagate `NOT NULL` from base tables unless explicitly cast; test relaxed to `YES` which matches actual DB output. T2 APPROVED. T3 findings (forward ADR-0042 reference in ADR-0041, D3(a) notation) rejected — consistent with sibling ADR-0040 precedent (validator-passed in F2.1a).

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| none | | |
