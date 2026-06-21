# Feature F4.2 — Evidence

> **Milestone:** 4 — search consumes published visibility contracts  ·  **Feature:** `f4.2-documents-search-projection`  ·  **Closed:** 2026-06-21
> **Contract:** `spec.md` (consumer-first: one published view `metaldocs.v_document_search_facts` = pure 1:1 projection of `public.documents` to the 14 columns the search consumer reads; no WHERE, no COALESCE).

## What was implemented

- **NEW** `db/migrations/0244_documents_search_projection.sql` — creates `metaldocs.v_document_search_facts`
  as a pure column projection of `public.documents` (the 14 columns search reads: `tenant_id, id,
  controlled_document_id, name, status, profile_code_snapshot, process_area_code_snapshot, created_by, code,
  revision_number, effective_from, effective_to, created_at, archived_at`). No `WHERE`, no `COALESCE` —
  `archived_at` exposed as a column (search keeps its own `archived_at IS NULL` filter in F4.3), nullable
  snapshot columns pass through. `COMMENT ON VIEW` records it as the ADR-0039 D3(a) published contract +
  mandatory `schema_migrations` row (`ON CONFLICT DO NOTHING`). No `security_invoker` — matches 0242/0243.
- **NEW** `internal/modules/documents/repository/document_search_facts_parity_integration_test.go` — the
  view-vs-base parity gate (producer proves itself before F4.3 repoints the consumer). Seeds 4 discriminator
  docs (standalone / CD-linked / archived / NULL-snapshot) and asserts NULL-safe per-column row equality,
  equal row counts, and archived-doc presence in the view.
- **EDIT** `wiki/decisions/0039-cross-module-base-table-read-boundary.md` — "Related code" annotated with
  migration 0244 + `v_document_search_facts` (no decision change; instance of M0/F0.1 ADR-0039 D3(a)/D4).
- **No Go production code, no search change, no documents repository change** — F4.2 only *publishes* the view;
  the search consume is F4.3.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing test first, then green | `go test -tags integration -run TestDocumentSearchFacts_ParityWithBaseTable ./internal/modules/documents/repository/` | **RED** before migration: `relation "metaldocs.v_document_search_facts" does not exist (SQLSTATE 42P01)`; **GREEN** after: `ok …/documents/repository 3.416s` | real (PG :5434) |
| View == base across {standalone, CD-linked, archived, NULL-snapshot}, NULL-safe; row counts equal; archived row PRESENT (no hidden filter) | same test (per-column equality + count + archived-presence assertions) | **PASS** (0.25s) | real (PG :5434) |
| Migration 0244 applies in full bootstrap | `go test -tags integration ./tests/integration/testdb/...` | `ok …/testdb 3.838s` (baseline + all `db/migrations` incl. 0244 apply clean) | real (PG :5434) |
| Static build | `go build ./...` | `BUILD-OK` (exit 0) | — |
| Guard unchanged (no F4.2 ledger edit — view published, no raw read removed yet) | `go run ./tools/cilint ./...` | `cilint-exit=0`; `hgcrossmodule.go` not modified | real |
| Cilint unit suite unaffected | `go test ./tools/cilint/...` | `ok …/internal/analyzers (cached)` (PendingBaseline still valid — no C4 row drained until F4.3) | real |

> The archived + NULL-snapshot docs are the anti-drift discriminators: archived presence proves no hidden
> `WHERE archived_at IS NULL`; NULL-snapshot equality (scanned via `sql.Null*`) proves no COALESCE baked into
> the view. The seed inserts raw `public.documents` rows under a session-level `document.create` capability
> assertion (`testdb.SetCapsOnDB`, conn pinned to 1) — the table has a capability-assertion trigger.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Migration 0244 applies cleanly on test PG | yes | testdb bootstrap row above |
| `v_document_search_facts` == base projected to the 14 columns, 1 row/doc, NULL-safe, incl. archived | yes | `TestDocumentSearchFacts_ParityWithBaseTable` GREEN (equality + count + archived presence) |
| ADR-0039 references the new view | yes | "Related code" note now lists 0244 + `v_document_search_facts` |
| Build unaffected | yes | `go build ./...` BUILD-OK |
| Guard unchanged (no F4.2 ledger edit) | yes | `cilint-exit=0`, no `hgcrossmodule.go` diff |

## Review disposition

- Spec-compliance review: **PASS** — view = consumer-derived 1:1 projection (no filter, no COALESCE,
  `archived_at` exposed); exactly the 14 columns search reads (least exposure); no consumer/documents-repository
  touched; no policy baked into the contract.
- Code-quality review: **PASS** — migration forward-only, idempotent (`ON CONFLICT DO NOTHING`), transactional,
  ADR-commented; test asserts per-column NULL-safe equality + count + archived presence (not a bare count) and
  exercises 4 discriminator shapes. Seed constraints discovered empirically during RED (capability-assertion
  trigger → `SetCapsOnDB` + 1 conn; `documents_status_check` → lowercase `'draft'`) — test-fixture fixes, not
  contract changes.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| none | — | — |
