# Module: search

> **Last verified:** 2026-05-01
> **Status:** Stub. Expand when v2 reader covers all entities and indexing strategy is locked.
> **Scope:** Cross-module search across templates, controlled documents, document versions.
> **Out of scope:** Full-text search of frozen PDFs (TBD, separate concern).
> **Key files:**
> - `internal/modules/search/` — backend module
> - `internal/modules/search/infrastructure/v2documents/reader.go` — v2 documents reader (JOINs `controlled_documents` to populate `DocumentCode`, fixed 2026-04-27)

## Approach

V2 reader joins source-of-truth tables on read rather than maintaining a denormalized search index. This trades a bit of query cost for zero-staleness — search results always reflect the live state.

## Indexed entities

- Controlled documents (by code, title, profile, area)
- Document versions (by state, dates)
- Templates (by title)

## Open questions

- When/whether to introduce a real search index (Elasticsearch, Postgres tsvector) — keep deferred until query patterns demand it.
- PDF body search — out of scope until requested.

## See also

- [architecture/data-model.md](../architecture/data-model.md)
- [modules/documents-v2.md](documents-v2.md)
