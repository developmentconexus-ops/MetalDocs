# Module: search

> **Last verified:** 2026-06-11 (Wave 1)
> **Status:** active (limited surface)
> **Maturity:** L2
> **Scope:** Cross-module search across templates, controlled documents, document versions.
> **Out of scope:** Full-text search of frozen PDFs (TBD, separate concern).
> **Key files:**
> - `internal/modules/search/` — backend module root
> - `internal/modules/search/application/service.go` — service: extracts actor from context, delegates single call to reader; no post-fetch authz filter
> - `internal/modules/search/domain/model.go` — `Document`, `Query` (with `ActorUserID`), `Classification`/`Status` consts; no `AccessPolicy`/`SubjectType`/`Effect` types (removed)
> - `internal/modules/search/domain/port.go` — `Reader` interface: `ListDocuments(ctx, query, limit, offset)`; no `ListAccessPolicies` (removed)
> - `internal/modules/search/infrastructure/v2documents/reader.go` — SQL reader: visibility enforced via `$13` (ActorUserID) against unified model (AD-3); phantom columns absent

## Approach

The v2 reader joins source-of-truth tables on read rather than maintaining a denormalized search index. This trades a bit of query cost for zero-staleness — search results always reflect the live state.

Per-document visibility is enforced entirely at the **data layer** inside the SQL query in `internal/modules/search/infrastructure/v2documents/reader.go`. The service extracts the authenticated caller's user ID from context, populates `Query.ActorUserID`, and issues a single `ListDocuments` call. Unauthenticated callers receive an empty result set without touching the reader. There is no post-fetch authz filter loop and no SQL paging-until-limit.

### Visibility predicate (unified model, AD-3)

A row in `public.documents` is included in results only when at least one condition is met for the caller (`$13`):

| Case | Condition |
|---|---|
| Standalone document (no linked controlled document) | `d.created_by = $13` |
| Controlled document — company scope | `cd.visibility_scope = 'company'` |
| Controlled document — owner | `cd.owner_user_id = $13` |
| Controlled document — restricted + area grant | active `controlled_document_area_grants` row joined to active `user_process_areas` row for caller |
| Controlled document — restricted + user grant | `controlled_document_user_grants` row for caller |

This is a verbatim port of the predicate at `internal/modules/controlleddocuments/infrastructure/repository.go:133-164`. There is no `system_admin` bypass — matches the controlled-documents list semantics.

### Columns returned by the reader

Only columns present on `public.documents` and `public.controlled_documents` are selected. The legacy governance columns (`subject_code`, `business_unit`, `classification`, `tags`) live on the decommissioned `metaldocs.documents` schema and are **not** part of v2 search.

| Field | Source |
|---|---|
| Title | `d.name` |
| DocumentProfile / DocumentType | `d.profile_code_snapshot` |
| DocumentFamily | subquery against `metaldocs.document_profiles` |
| ProcessArea | `d.process_area_code_snapshot` |
| Department | `cd.department_code` |
| OwnerID | `d.created_by` |
| DocumentCode | `COALESCE(cd.code, d.code)` |
| DocumentSequence | `COALESCE(cd.sequence_num, d.revision_number)` |
| EffectiveAt / ExpiryAt | `d.effective_from` / `d.effective_to` |
| CreatedAt | `d.created_at` |

**Wave 1 (F-13a/b):** `Subject`, `BusinessUnit`, `Classification`, and `Tags` fields were removed from `SearchDocumentResponse`; the undocumented camelCase `businessUnit` query parameter was also removed. The legacy governance columns (`subject_code`, `business_unit`, `classification`, `tags`) live on the decommissioned `metaldocs.documents` schema and are not part of v2 search.

## Indexed entities

- Controlled documents (by code, title, profile, area)
- Document versions (by state, dates)
- Templates (by title — deferred, see T-002 in tech-debt register)

## Open questions

- When/whether to introduce a real search index (Elasticsearch, Postgres tsvector) — keep deferred until query patterns demand it.
- PDF body search — out of scope until requested.

## Failure modes

| Failure | Symptom | Detection | Response |
|---|---|---|---|
| Postgres unavailable | Search returns 5xx; UI shows generic error | `infrastructure/v2documents/reader.go` SQL error | Check `metaldocs-api` → Postgres connectivity; no fallback (no separate index) |
| Unauthenticated caller | Empty result set returned (no error) | Service returns before calling reader | Expected behavior — no auth = no visibility |
| Live join cost on hot list views | Search latency climbs as `controlled_documents` JOIN row count grows | DB query latency metrics; slow query log | Open Issue if SLA breached; see `Open questions` — index introduction tracked |
| Caller has no grants on restricted documents | Rows excluded by SQL visibility predicate | Empty result is correct; verify `user_process_areas` / `controlled_document_*_grants` rows | Verify IAM area membership and grant records for the caller |
| Stale data because no separate index | (Not applicable — design choice) | N/A | By design: live JOIN guarantees zero staleness |

## See also

- [architecture/data-model.md](../architecture/data-model.md)
- [modules/documents.md](documents.md)
- [search-tech-debt.md](search-tech-debt.md)
- [backlog/search-refactor.md](../backlog/search-refactor.md)
- [decisions/0022-authz-capability-coherence.md](../decisions/0022-authz-capability-coherence.md) — AD-3 unified visibility model ported into this reader

## Risks & Technical Debt

Open tech-debt items by severity (source: [search-tech-debt.md](search-tech-debt.md)):

- Critical: 0
- Major: 0
- Minor: 1 (T-001: ✅ CLOSED Wave 1 — 405 bare responses → `httpresponse.WriteMethodNotAllowed`; T-002: template search path deferred)

Refactor backlog: [../backlog/search-refactor.md](../backlog/search-refactor.md)
