# Phase 0 — Context load: registry

## Existing wiki position

- `wiki/modules/controlled-documents.md` exists as 128-line stub (Last verified 2026-05-07). Covers `controlled_documents` table, atomic create handler, `getActiveDocument` FULL OUTER JOIN, 8 HTTP routes, frontend RegistryDetailPage / PublishedDownloadCell, `DocumentInitializer` port. Refresh to Arc42 + C4 living doc.
- `wiki/concepts/controlled-documents.md` (Last verified 2026-05-07) — CD format `{profile}-{area}-{NNN}`, atomic create endpoint, preview, revision endpoint, `cd_sequence_counters` per (tenant_id, profile_code, process_area_code).
- `wiki/decisions/0011-cd-atomic-create.md` (Accepted 2026-05-07; verified 2026-05-10) — atomic create + per-area numbering + `Idempotency-Key` adoption in one PR; deleted legacy two-call flow + `RegistryCreateDialog` + `DocumentCreatePage`.
- `wiki/modules/documents.md` — downstream consumer. Documents implements `DocumentInitializer` via `CDDocumentInitializer`; exposes `CreateDocumentTx(*sql.Tx, ...)` port to registry.
- `wiki/modules/taxonomy.md` — upstream. Profiles + Areas (tenant-scoped) drive CD code prefix. FK targets: `document_profiles.code`, `process_areas.code`.

## Boundary resolution (registry vs taxonomy)

- **Taxonomy** owns the hierarchical classification: `document_families → document_profiles → process_areas`. These are the abstract types/buckets.
- **Registry** owns the concrete catalog of issued documents: each `controlled_documents` row is a numbered instance bound to (`profile_code`, `process_area_code`), with a per-(tenant, profile, area) monotonic sequence and a chain of `documents` revisions.
- Registry depends on taxonomy via FK (profile_code, process_area_code resolution at create-time via `TaxonomyProfileReader` / `TaxonomyAreaReader` — `module.go:29-30`).

## Cross-deps preview (validated in Phase 3)

- **IN-edges:** templates (template version clone), documents (CreateDocumentTx port), approval (CD-scoped sign-off routes), search, frontend `features/controlled-documents/`, novo-documento wizard.
- **OUT-edges:** documents (`CreateDocumentTx`), taxonomy (profile/area reader), `internal/platform/idempotency`, audit (governance logger; currently via `taxonomyapp.NewDBGovernanceLogger`).
- **Capability namespace:** `registry.create` seeded in migration `0165_role_capabilities_reseed.sql` for roles `editor`, `author`, `system_admin`. T-001 dual-namespace applies — check if both `registry.*` and legacy capability codes exist; document gap.

## Migrations touching registry

- `0103` (W1 scaffold — dropped by 0168)
- Atomic create / sequence overhaul (per ADR 0011): introduces `cd_sequence_counters` (tenant_id, profile_code, process_area_code), retires `profile_sequence_counters`.
- `0165` — capability reseed (registry.create).
- `0168` — drop `public.documents_v2`.
- Full list enumerated in Phase 4 artifact.

## Phase 2 op picks

- **Read op:** `GET /api/v1/controlled-documents/{id}/active-document` — FULL OUTER JOIN published-vs-active resolution (interesting query, E10 fix history).
- **Write op:** `POST /api/v1/controlled-documents` — atomic CD + first revision (multi-module tx, idempotency middleware, sequence allocation, cross-module port call).
- **State-transition:** `PUT /api/v1/controlled-documents/{id}/obsolete` and `/supersede` — lifecycle transitions. The CD itself has a slim status (`active | obsolete | superseded`), distinct from document revision approval state. Document one transition trace (obsolete). If lifecycle is trivial (status-only flip with no guards), record "minimal state machine — see §6" rather than skipping §6 entirely.

## Open questions deferred to tech-debt

- Does `controlled_documents` have a `tenant_id` column + tripwire? Confirm in Phase 4. If absent, multi-tenant data leak → Critical T-row.
- Is the legacy `profile_sequence_counters` table dropped, or coexists with `cd_sequence_counters`? If coexists with no migration cleanup, log as maint:migration-cleanup.
- Is `CreateRevision` route protected by `registry.create` or a separate capability (`registry.revise`)? If a single capability covers both, log as authz-granularity Minor.
- `governanceLogger` is sourced from `taxonomyapp.NewDBGovernanceLogger` — cross-module audit-logger reuse. Is registry meant to own its own logger or is this intentional sharing? Record in tech-debt.
- The 8th constructor arg in `application.NewRegistryService(..., nil)` is `nil` — what is unwired? Phase 1/2 to identify.

## Phase-specific notes

- §6 (Runtime View) requirement: read + write + obsolete state-transition. NOT n/a.
- §5 (Industry comparison): expected anchors — Stripe idempotency-key replay semantics (`platform/idempotency`), atomic resource creation pattern (multi-row write in single tx).
- §8 (Cross-cutting): two-tier authz (cap `registry.create` + `authz.Require` area scope), Postgres tripwire (TBD — Phase 4), RFC 9457 envelope status.

