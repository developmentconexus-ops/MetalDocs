# Modules

> **Last verified:** 2026-05-06
> **Scope:** Per-module deep dives. One file per backend module / frontend feature area.

- [editor-ui-eigenpal.md](editor-ui-eigenpal.md) — eigenpal integration layer (Last verified: 2026-05-06)
- [editor-chrome.md](editor-chrome.md) — shared toolbar overlay + eigenpal CSS overrides for eigenpal-based pages; slot API, `VersionBadge`, `AutosaveStatus`; consumed by templates + documents (Last verified: 2026-05-06)
- [templates-v2.md](templates-v2.md) — template authoring, versioning, approval; List screen + creation wizard Steps 1–4 (Step 5 stub); mocked DOCX flow + mocked permissions roles/areas/counts (Last verified: 2026-05-09)
- [documents.md](documents.md) — document instances, Library screen, editing flow, session model, API; `libraryStatus.ts` status-meta, `AuthorCell`, `asApiError`, debounced search, `keepPreviousData`/`staleTime`; `internal/modules/documents/`, table `public.documents` (Last verified: 2026-05-06)
- [registry.md](registry.md) — controlled-document catalog, code generation, active-document FULL OUTER JOIN (E10), registry detail page, PublishedDownloadCell (Last verified: 2026-05-04)
- [taxonomy.md](taxonomy.md) — families (global), profiles, areas; CRUD routes, scoping distinction, deactivation guards (Last verified: 2026-05-02)
- [approval.md](approval.md) — routes, signoffs, ISO segregation; known gaps D4/E4/outbox (Last verified: 2026-05-03)
- render-fanout.md — TBD (Last verified: 2026-05-04)
- [iam-rbac.md](iam-rbac.md) — capabilities, roles, DB-backed CanDo, system_admin bypass, group grants (Last verified: 2026-05-03)
