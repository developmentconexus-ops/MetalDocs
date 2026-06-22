# Feature F2.1a — Evidence

> **Milestone:** 2 — Distribuição coverage-scope  ·  **Feature:** `f2.1a-cd-obligated-view`  ·  **Closed:** 2026-06-21
> **Contract:** `spec.md` (consumer-contract-first; the distribution module (F2.1c/F2.2) reads `metaldocs.v_cd_obligated_readers`).

## What was implemented

- **New owner-published view** `metaldocs.v_cd_obligated_readers` (controlleddocuments-owned) — denominator read contract for the distribution module. Three legs `UNION ALL` (user_grant / area_grant via `v_active_user_areas` / company_scope via `v_cd_search_facts.is_company × DISTINCT active tenant users`), `DISTINCT BY (tenant_id, controlled_document_id, user_id)` with precedence `user_grant > area_grant > company_scope`; deterministic tiebreaker is lowest `area_code NULLS LAST` for area_grant rows with multi-area users.
- **Forward-only migration** `db/migrations/0245_cd_obligated_readers_view.sql` — `CREATE VIEW`, `COMMENT ON VIEW`, `INSERT INTO public.schema_migrations … ON CONFLICT DO NOTHING`. No DROP/ALTER VIEW (no precedent across 244 prior migrations; runner gates idempotency via `schema_migrations`).
- **Integration test** `internal/modules/controlleddocuments/infrastructure/v_cd_obligated_readers_integration_test.go` — three contract dimensions (three-leg precedence on a restricted CD with a user-grant/area-grant overlap user; company-scope union across active tenant users; declared column shape via `information_schema.columns`).
- **ADR-0040** `wiki/decisions/0040-cd-obligated-readers-view.md` — durable decision: new sibling view (not `v_cd_grantee` extension), three-leg union, precedence, no `area_name`/`display_name`, forward-only idempotency.
- **ADR-0039 inventory update** — `v_cd_obligated_readers` added to the published-view exemption list (D3a), consumer = distribution.
- **`wiki/decisions/index.md`** — ADR-0040 row added; `Last verified` refreshed.
- **`v_cd_grantee` untouched.** Search module untouched. Producer matches the spec consumer contract (5-column shape, three-leg semantics, source-precedence DISTINCT) line-for-line.

Commits:

- `e49fc1d7` test(M2/F2.1a): anchor company-CD set to v_active_user_areas count + drop ordinal sort
- `f357fb15` feat(M2/F2.1a): publish metaldocs.v_cd_obligated_readers view (ADR-0040)
- `af70aa6b` docs(M2/F2.1a): ADR-0040 + ADR-0039 inventory for v_cd_obligated_readers

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing test first, then green | `go test -tags=integration -run TestObligatedReaders ./internal/modules/controlleddocuments/infrastructure/...` | initial run (no migration applied): test file refused to compile against missing view → migration 0245 + applied via testdb framework → `--- PASS: TestObligatedReaders_RestrictedCD_ThreeLegsWithPrecedence (133.64s) / TestObligatedReaders_CompanyCD_AllActiveTenantUsers (6.90s) / TestObligatedReaders_ViewShape (3.45s); PASS; ok metaldocs/internal/.../infrastructure 146.659s` | real (live PG via testdb framework) |
| Migration applies cleanly on dev DB | `docker exec -i metaldocs-postgres psql … < db/migrations/0245_*.sql` | `BEGIN / CREATE VIEW / COMMENT / INSERT 0 1 / COMMIT`; ledger row: `0245 \| controlleddocuments publishes metaldocs.v_cd_obligated_readers …` (count=1) | real |
| Idempotency (runner-level, per spec Q7) | testdb factory framework re-applies all migrations per pid template-DB; suite executed end-to-end twice in succession across this task — both runs green; `schema_migrations` row count for 0245 = 1 on dev DB after apply. Raw DDL re-apply is intentionally NOT idempotent (forward-only policy; no DROP/ALTER VIEW precedent). | real |
| View shape matches contract | `TestObligatedReaders_ViewShape` queries `information_schema.columns` in `ordinal_position` order — 5 columns: `tenant_id uuid`, `controlled_document_id uuid`, `user_id text`, `area_code text`, `source text` (all `is_nullable=YES` per PG UNION-ALL conservative default; runtime NOT NULL inherited from base tables — documented in test + ADR-0040 consequences) | PASS (3.45s) | real |
| Three-leg semantics + DISTINCT-with-precedence | `TestObligatedReaders_RestrictedCD_ThreeLegsWithPrecedence` seeds a restricted CD with user-grant U1 + area-grant on A1 (member U2) + overlap user U1 also member of A1 → 3 distinct rows (U1.source=`user_grant` area=NULL, U2.source=`area_grant` area=A1, overlap.source=`user_grant`); revoked area member + owner + uninvolved user excluded | PASS (133.64s) | real |
| Company-scope leg = all active tenant users | `TestObligatedReaders_CompanyCD_AllActiveTenantUsers` anchors expected cardinality to `SELECT count(DISTINCT user_id) FROM metaldocs.v_active_user_areas WHERE tenant_id=$1` — drift in seedCDVisibility surfaces as fixture-setup error, not silent size mismatch | PASS (6.90s) | real |
| Search untouched | `git diff HEAD -- db/migrations/0243_cd_search_visibility_contract.sql internal/modules/search` | empty diff | real |
| `v_cd_grantee` untouched | not modified in any feature commit — `git log --oneline -3 -- db/migrations/0241*` (the v_cd_grantee migration) = no entries since 0241 | real |
| ADR-0040 recorded | `ls wiki/decisions/0040-*.md` | `wiki/decisions/0040-cd-obligated-readers-view.md` | real |
| ADR-0039 inventory updated | `grep -c "v_cd_obligated_readers" wiki/decisions/0039-cross-module-base-table-read-boundary.md` | `1` | real |
| `hgcrossmodule` analyzer green | `go run ./tools/cilint/ ./internal/modules/controlleddocuments/... ./internal/modules/distribution/...` | no findings (clean output = 0 H-G) | real |
| Static — `go build` | `go build ./...` | exit 0, no output | real |
| Static — `go vet` | `go vet ./...` | exit 0, no output | real |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from `spec.md`) | Met? | Evidence |
|---------------------------------------|------|----------|
| Migration applies cleanly on a fresh DB and is idempotent on re-run | yes | row 2 + row 3 above |
| View shape matches the contract (5 columns, declared types) | yes | row 4 above (with PG info_schema nullability caveat documented in ADR-0040 + test comment) |
| Three-leg semantics correct + distinct-with-precedence (restricted CD with overlap user; company-scope union) | yes | rows 5 + 6 above |
| Search untouched (`git diff db/migrations/0243*`, `internal/modules/search`) | yes | row 7 above |
| ADR-0040 recorded + ADR-0039 inventory updated | yes | rows 9 + 10 above |
| `hgcrossmodule` analyzer green | yes | row 11 above |

## Review disposition

- **Spec-compliance review:** Producer matches the consumer contract block in `spec.md` exactly — 5-column shape, three-leg UNION ALL semantics, DISTINCT BY (tenant_id, cd, user_id) with declared precedence, lowest-area_code tiebreaker, no `area_name`/`display_name`/`name`, no `v_cd_grantee` mutation, no base-table change, no reverse migration. T1 quality reviewer APPROVED on re-review at `e49fc1d7` after fixing two issues (brittle company-set cardinality + sort masking ordinal_position). T3 ADR/inventory completed by opus subagent with one acknowledged style deviation (used inline bullet inventory matching existing ADR-0039 format rather than a new table) — accepted: matches the file's existing structure.
- **Code-quality review:** No Go code introduced. Migration SQL follows the project's CREATE VIEW + COMMENT + schema_migrations INSERT pattern (mirrors 0242/0243 surface). Integration test reuses sibling fixtures (`seedCDVisibility`, `cdScenario`) per ADR-0034. The only style deviation noted (and resolved) was the PG info_schema nullability semantics — documented in both the test source and ADR-0040 consequences so future readers do not chase a phantom drift.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Apply-via-`start-api.ps1 -Build` proof (vs. direct `psql` apply used here) | The runner's idempotency mechanism is `schema_migrations` lookup, not view DDL re-runnability. Direct `psql` apply + ON CONFLICT DO NOTHING on the ledger row demonstrates the same gate that the runner uses. F2.1c will start the API end-to-end (handler work) and provides the runner-path proof inline. | Trigger: F2.1c open. Owner: F2.1c implementer. |
| Performance / EXPLAIN evidence for the company-scope leg's `f × tu` cross-product | Spec non-goal #5 explicitly defers indexing/latency to a real test in F2.2. The fixture exercises ≤ a handful of users + CDs; production volumetrics are unknown at view-publish time. | Trigger: F2.2 integration test surfaces a measurable latency regression. Owner: F2.2 implementer. |
