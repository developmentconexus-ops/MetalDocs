# Feature F4.1 — Evidence

> **Milestone:** 4 — search consumes published visibility contracts  ·  **Feature:** `f4.1-cd-visibility-contract`  ·  **Closed:** 2026-06-21
> **Contract:** `spec.md` (consumer-first: two published `metaldocs` views derived from the search consumer `v2documents/reader.go` — `v_cd_search_facts` 1 row/CD projection+scalar legs, `v_cd_grantee` bounded restricted-CD edges). Shape 2, operator-ratified.

## What was implemented

- **NEW** `db/migrations/0243_cd_search_visibility_contract.sql` — creates two ADR-0039 D3(a)/D4 views over
  CD-owned base tables + iam's published `metaldocs.v_active_user_areas`:
  - `metaldocs.v_cd_search_facts` — exactly 1 row per `controlled_documents` row: projection cols
    (`code, department_code, profile_code, sequence_num`) + `is_company` (= `visibility_scope = 'company'`,
    so the consumer never names CD's scope enum) + `owner_user_id`. The projection LEFT JOIN target;
    strictly 1:1 so search's projection join cannot fan out document rows.
  - `metaldocs.v_cd_grantee` — bounded visibility edges, **restricted CDs only**: `controlled_document_area_grants`
    ⋈ `metaldocs.v_active_user_areas` (active-now, `effective_to IS NULL` already encoded by the iam view —
    revoked members excluded) UNION direct `controlled_document_user_grants`. Company-scope CDs contribute no
    edges, so the unbounded (cd × actor) cross-product is never materialized (HS-2 avoided).
  - Both with `COMMENT ON VIEW` recording the published-contract role + the mandatory
    `public.schema_migrations` row (`ON CONFLICT DO NOTHING`). No `security_invoker` — matches 0242's posture.
- **NEW** `internal/modules/controlleddocuments/infrastructure/cd_visibility_contract_parity_integration_test.go`
  — the view-vs-raw parity gate (producer proves itself before F4.3 repoints the consumer). Reuses M3's
  `seedCDVisibility`/`cdScenario` (company + restricted CDs; owner/areaMember/revokedMem/userGrant/none).
- **EDIT** `wiki/decisions/0039-cross-module-base-table-read-boundary.md` — "Related code" annotated with
  migration 0243 + the two view names (no decision change; instance of M0/F0.1 ADR-0039 D3(a)/D4).
- **No Go production code, no search change, no CD repository change** — F4.1 only *publishes* the views;
  the search consume is F4.3, the documents projection (C4a) is F4.2.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| Facts view 1 row/CD; projection + `is_company` + owner == base table (NULL-safe) | `go test -tags integration -run TestCDSearchFacts_ParityWithBaseTable ./internal/modules/controlleddocuments/infrastructure/` | **PASS** (0.24s) | real (PG :5434) |
| Grantee set == active grant-leg set; revoked member ∉, ungranted ∉, company CD has 0 edges | `…-run TestCDGrantee_BoundedSetExcludesRevokedAndUngranted …` | **PASS** (0.07s) | real (PG :5434) |
| Composed decision `is_company OR owner=$13 OR EXISTS(grantee=$13)` == verbatim raw `reader.go:89-118` predicate, 5 actors × {company,restricted} | `…-run TestCDVisibilityContract_ComposedDecisionParityWithRaw …` | **PASS** (0.08s) | real (PG :5434) |
| Migration 0243 applies in full bootstrap | `go test -tags integration ./tests/integration/testdb/...` | `ok …/testdb 3.893s` | real (PG :5434) |
| Static build | `go build ./...` | `BUILD-OK` (exit 0) | — |
| Guard unchanged (no F4.1 ledger edit — views published, no raw read removed yet) | `go run ./tools/cilint ./...` | `cilint-exit=0`; `hgcrossmodule.go` not modified | real |
| Cilint unit suite unaffected | `go test ./tools/cilint/...` | `ok …/internal/analyzers (cached)` (PendingBaseline still valid — no C4 row drained until F4.3) | real |

> The revoked-member + ungranted-user rows are the anti-drift discriminators. The facts projection columns
> are compared NULL-safely (`sql.NullString`/`sql.NullInt64`): the view passes `department_code` etc. through
> verbatim (search applies its own `COALESCE` in F4.3); an earlier plain-`string` scan tripped on a NULL
> `department_code` — a test-scan-type fix, not a view change.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Migration 0243 applies cleanly on real PG (test DB :5434) | yes | testdb bootstrap row above |
| `v_cd_search_facts` 1 row/CD with correct projection + `is_company` + owner == base | yes | `TestCDSearchFacts_ParityWithBaseTable` PASS |
| `v_cd_grantee` set == inline grant-leg set, excluding revoked + ungranted | yes | `TestCDGrantee_BoundedSetExcludesRevokedAndUngranted` PASS |
| Composed decision == verbatim pre-M4 inline predicate, all 5 actor scopes | yes | `TestCDVisibilityContract_ComposedDecisionParityWithRaw` PASS |
| ADR-0039 references the two new views | yes | "Related code" note now lists 0243 + `v_cd_search_facts`/`v_cd_grantee` |

## Review disposition

- Spec-compliance review: **PASS** — views = consumer-derived contract (Shape 2); `is_company` replaces the
  `'company'` literal; no (cd × actor) cross-product; no consumer/CD-repository touched; no new authz semantics
  (seam-only — composed decision proven set-identical to the raw predicate).
- Code-quality review: **PASS** — migration forward-only, idempotent, transactional, ADR-commented; test
  asserts set/row equality + drift exclusions (not counts) and runs the raw predicate verbatim from a copy of
  `reader.go:89-118` so any future predicate divergence fails the gate.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| none | — | — |
