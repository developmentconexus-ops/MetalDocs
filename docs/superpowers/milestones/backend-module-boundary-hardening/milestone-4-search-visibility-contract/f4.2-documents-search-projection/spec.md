# Feature F4.2 — Spec: documents publishes a search projection read contract

> **Milestone:** 4 — search consumes published visibility contracts  ·  **Folder:** `f4.2-documents-search-projection`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-21 / leandrotca (operator ratified the 1:1 passthrough projection shape via the F4.2 consumer-contract interview) — *no implementation begins until this line is filled.*

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

Engine: consumer-contract-first dialog seeded with milestone.md F4.2 row. The decisive question (HS-2 risk
item) was the view *shape* — whether the contract bakes in policy (active-only) or stays a pure projection.

| # | Question | Answer |
|---|----------|--------|
| 1 | What does the search consumer read from `public.documents`? | The 14 columns `reader.go` projects/filters: `tenant_id, id, controlled_document_id, name, status, profile_code_snapshot, process_area_code_snapshot, created_by, code, revision_number, effective_from, effective_to, created_at, archived_at`. (Projection cols 1-13 + the `archived_at IS NULL` and visibility-leg filters.) |
| 2 | One published view or several? | One — `metaldocs.v_document_search_facts`, 1 row per document. The documents↔CD LEFT JOIN stays in search (F4.3 joins this view to `v_cd_search_facts`); F4.2 only publishes the documents leg. |
| 3 | Does the view filter (`archived_at IS NULL`) or COALESCE? | **No (operator-ratified 1:1 passthrough).** `archived_at` is exposed as a column; search keeps its own `WHERE d.archived_at IS NULL` and all `COALESCE` in F4.3. Behavior stays with the consumer that decides it; the contract promises columns, not policy. |
| 4 | Why passthrough over active-only? | (a) The F4.3 seam becomes a pure rename `FROM public.documents` → `FROM metaldocs.v_document_search_facts` — every filter/COALESCE/ORDER BY/LIMIT untouched, so parity is trivial total row-equality. (b) Mirrors F4.1's `v_cd_search_facts` (one consistent "facts view = projection" rule). (c) No policy baked into the contract → reusable by any future read-only consumer without re-versioning. |
| 5 | Owner / which module's migration? | documents — the view reads only the documents-owned `public.documents` base table (compliant D3a producer). Migration 0244. |

## Consumer contract (FIRST — before any producer)

- **Consumer:** `internal/modules/search/infrastructure/v2documents/reader.go` `ListDocuments`. The only F4.2
  consumer (F4.3 wires it). Today it reads `FROM public.documents d` directly (C4a violation).
- **Contract — one published, versioned view (ADR-0039 D3a), schema `metaldocs`:**

  **`metaldocs.v_document_search_facts`** — exactly 1 row per `public.documents` row, the columns search reads:
  | column | type | source (`public.documents.*`) | role in search |
  |--------|------|-------------------------------|----------------|
  | `tenant_id` | uuid | `tenant_id` | tenant filter + join key |
  | `id` | uuid | `id` | projection + ORDER BY tiebreak |
  | `controlled_document_id` | uuid (nullable) | `controlled_document_id` | LEFT JOIN key to `v_cd_search_facts`; NULL ⇒ standalone-doc visibility leg |
  | `name` | text | `name` | projection + text filter |
  | `status` | text | `status` | projection + status filter |
  | `profile_code_snapshot` | text (nullable) | `profile_code_snapshot` | projection + profile/family filter + family join key |
  | `process_area_code_snapshot` | text (nullable) | `process_area_code_snapshot` | projection + process-area filter |
  | `created_by` | text | `created_by` | projection + owner filter + standalone visibility (`created_by = $13`) |
  | `code` | text (nullable) | `code` | projection (`COALESCE(cd.code, d.code, '')`) |
  | `revision_number` | int | `revision_number` | projection (`COALESCE(cd.sequence_num, d.revision_number, 0)`) |
  | `effective_from` | timestamptz (nullable) | `effective_from` | projection |
  | `effective_to` | timestamptz (nullable) | `effective_to` | projection + expiry-before/after filters |
  | `created_at` | timestamptz | `created_at` | projection + ORDER BY |
  | `archived_at` | timestamptz (nullable) | `archived_at` | search's `archived_at IS NULL` active filter (exposed, not pre-applied) |

- **Source of truth for the contract:** `reader.go:54-121` (the `public.documents` column references). ADR-0039 D3a/D4.

## What this feature implements

A single migration (`db/migrations/0244_*.sql`) creating the `metaldocs.v_document_search_facts` view above
over the documents-owned `public.documents` base table — pure column projection, no `WHERE`, no `COALESCE`,
1 row/document — with a `COMMENT ON VIEW` documenting it as a published ADR-0039 D3a read contract and the
mandatory `schema_migrations` row. ADR-0039's `Related code` note is extended to reference it. No Go code, no
change to documents' own repository, no change to search (that is F4.3).

## Non-goals (mandatory)

- **No search-side change** — `reader.go` is untouched in F4.2 (F4.3 consumes the view).
- **No CD projection** — the CD leg (C4b/c/e) is F4.1's `v_cd_search_facts`/`v_cd_grantee`, already published.
- **No filter or COALESCE baked into the view** — passthrough only; `archived_at` exposed as a column.
- **No change to documents' own repository** — F4.2 only *publishes* a view.
- **No new columns beyond the 14 search reads** — least-exposure contract.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Migration 0244 applies cleanly on real PG (test DB :5434) | `go test -tags integration ./tests/integration/testdb/...` (full bootstrap incl. 0244) + the parity test below | real |
| `v_document_search_facts` == `public.documents` projected to the 14 columns, **1 row/doc, NULL-safe, incl. archived rows** | new integration parity test `document_search_facts_parity_integration_test.go` — view row == base row across {standalone, CD-linked, archived, NULL-snapshot} docs; row count equal; archived row PRESENT in the view (passthrough, not filtered) | real |
| ADR-0039 references the new view | grep ADR-0039 for `v_document_search_facts` | real |

> TDD: write the failing parity test (against the not-yet-created view) first, then add the migration to
> green. The archived-doc + NULL-snapshot rows are the anti-drift discriminators (prove no hidden filter, no
> COALESCE). **HS-3:** if test PG :5434 is down, these steps are **not-run**, never false-green.

## ADR needed?

- [x] No *new* ADR — instance of the already-Accepted **ADR-0039 D3a** mechanism (published view). Action:
  extend ADR-0039's `Related code` note to list migration 0244 and `v_document_search_facts`. No new decision.
