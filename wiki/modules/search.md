# Module: search

> **Last verified:** 2026-05-26
> **Status:** active (limited surface)
> **Maturity:** L2
> **Scope:** Cross-module search across templates, controlled documents, document versions.
> **Out of scope:** Full-text search of frozen PDFs (TBD, separate concern).
> **Key files:**
> - `internal/modules/search/` â€” backend module
> - `internal/modules/search/infrastructure/v2documents/reader.go` â€” v2 documents reader (JOINs `controlled_documents` to populate `DocumentCode`, fixed 2026-04-27)

## Approach

V2 reader joins source-of-truth tables on read rather than maintaining a denormalized search index. This trades a bit of query cost for zero-staleness â€” search results always reflect the live state.

As of 2026-05-26, the `public.documents` reader is intentionally bounded to fields that exist in the live runtime schema and are explicitly selected/scanned by `internal/modules/search/infrastructure/v2documents/reader.go`: title, profile/type, family, process area, subject, owner, business unit, department, classification, tags, code/sequence, status, effective/expiry dates, and created-at ordering. The service now pages through SQL batches before trimming to the caller's limit so authorization filtering cannot silently drop later authorized matches just because earlier rows were denied, and area-policy evaluation continues to use the live `businessUnit:department` resource key instead of a partially hydrated surrogate.

## Indexed entities

- Controlled documents (by code, title, profile, area)
- Document versions (by state, dates)
- Templates (by title)

## Open questions

- When/whether to introduce a real search index (Elasticsearch, Postgres tsvector) â€” keep deferred until query patterns demand it.
- PDF body search â€” out of scope until requested.

## See also

- [architecture/data-model.md](../architecture/data-model.md)
- [modules/documents.md](documents.md)
- [search-tech-debt.md](search-tech-debt.md)
- [backlog/search-refactor.md](../backlog/search-refactor.md)


## 11. Risks & Technical Debt

- Critical: 0
- Major: 1
- Minor: 2

Refactor backlog: [../backlog/search-refactor.md](../backlog/search-refactor.md)
