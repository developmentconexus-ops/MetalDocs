# Phase 0 — Context for `documents` module doc

**Date:** 2026-05-10
**Composer:** main agent (Opus 4.7)
**Module path:** `internal/modules/documents/`
**Table:** `public.documents`

## Inputs read

- `wiki/README.md` — index, last verified 2026-05-10.
- `wiki/modules/documents.md` — existing module doc (Last verified 2026-05-09). Already substantial; this run upgrades to Arc42 + C4 + tech-debt + backlog deliverables.
- `wiki/modules/documents-v2.md` — DEPRECATED stub (`renamed to documents` post-migration 0167/0168). **Decision:** retire after this run; wiki-curator removes/redirects in Phase 7. The current `documents.md` already supersedes it.
- `wiki/decisions/0001-eigenpal-adoption.md` — DOCX editor choice.
- `wiki/decisions/0007-two-tier-authz.md` — tier-1 / tier-2 split; documents wired via `NewCapabilityChecker` adapter (J2 amendment).
- `wiki/decisions/0011-cd-atomic-create.md` — atomic CD+document create; `CreateDocumentTx` port crosses registry→documents boundary; `POST /api/v2/documents` deleted.
- `wiki/decisions/0012-contract-first-api.md` — oapi-codegen bootstrap on documents; handler migration deferred due to spec drift.
- `wiki/concepts/placeholders.md` — fixed 7-token catalog substituted at freeze; `placeholder_schema_snapshot` populated by `application.SnapshotService`; `enforce_snapshot_on_submit_trg` enforces non-NULL pre-`under_review`.
- `wiki/concepts/token-syntax.md` — `{name}` single-brace tokens; relevant to fillin / freeze paths.
- `wiki/modules/iam-tech-debt.md` — IN-edge surface for IAM dependencies, used to source cross-module fact framing.

## Iam dependency surface to capture as IN-edges in Phase 3

Per user pre-flight note + iam-tech-debt:

- **T-001 (dual capability namespaces).** `internal/modules/documents/application/fillin_authz.go:9` imports the typed `iamdomain.Capability` consts (`CapDocumentView/Create/Edit`, `CapWorkflowReview/Approve`). `apps/api/internal/wiring/documents.go:7` imports the same. Documents straddles both namespaces (string `doc.*` via Postgres tripwire, typed `Capability` via in-process check). Record verbatim as IN-edge fact.
- **T-009 (`ErrCapabilityDenied` collision).** `internal/modules/documents/delivery/http/handler.go:17` imports `iamapp.ErrCapabilityDenied` (sentinel variant). Record as IN-edge fact + cross-link in §8.
- **ADR 0007 J2 amendment.** `internal/modules/documents/application/ports.go` declares the consumer port `CapabilityChecker`; `apps/api/internal/wiring/documents.go:24` provides the `NewCapabilityChecker` adapter. Record in §5/§8.

## Module shape (preliminary, before Codex Phase 1)

- `application/` — large surface (~25 files): service.go, snapshot_service.go, freeze_service.go, fillin_service.go, fillin_authz.go, view_service.go, export_service.go, draft_resolver_service.go, reconstruct_service.go, context_builder.go, cd_initializer.go, snapshot_resolver, iam_user_options, ports, list_options.
- `delivery/http/` — handler.go (large), export_handler.go.
- `repository/` — repository.go (CRUD + paginated list + stats), fillin_repository.go, export_repository.go, snapshot_repository.go, resolver_readers.go.
- `domain/` — model.go, state.go (state machine), errors.go, snapshot.go, composite_hash.go, values_hash.go, comment.go, export.go.
- `jobs/` — orphan_pending_sweeper.go, session_sweeper.go.
- `approval/` — **sub-package owned by documents** with own application/domain/http/infra/infrastructure/repository. Cross-module question: this is the "submit for review / approval instance" surface within documents (different from top-level `internal/modules/approval/`). Will need to map carefully in Phase 1/3.
- `api/api.gen.go` — codegen bootstrap (per ADR 0012; handler migration deferred).
- `module.go` — DI wiring.

## Migrations to enumerate in Phase 4

Search range (rough): 0001..0014 (legacy), 0103 (`docx_v2_documents`), 0105 (`docx_v2_document_revisions`), 0107 (`docx_v2_document_checkpoints`), 0126/0129/0131 (`documents_v2` patches), 0167/0168 (rename `documents_v2` → `documents` + drop old), 0181 (drop `documents.locked_at`), 0183 (NOT NULL name + CHECK). Codex Phase 4 must run full enumeration; the list above is a hint, not authoritative.

## Phase 2 operation picks

Per skill rubric (one read, one write, one state-transition):

1. **Read:** `GET /api/v2/documents` (`listDocuments`, paginated list with stats sibling). Hits tier-1 cap check + RBAC scoping via `effectiveUserID`, two-query pagination, no tx.
2. **Write:** `PUT /api/v2/documents/:id/revisions` (autosave). Acquires writer session implicitly, commits new revision, updates `storage_key`. Touches `document_revisions`.
3. **State transition:** `POST /api/v2/documents/:id/finalize`. Atomic `draft → under_review` + approval-instance create. Owned by `internal/modules/documents/approval/` sub-package OR delegated to `internal/modules/approval/` — Codex must determine. Cross-module call surface is the most important thing to map.

`POST /api/v2/controlled-documents` (CD + draft create) is in **registry**, not documents — but it consumes the `documents.CreateDocumentTx` port. Note as OUT-edge in §3, not a documents-owned op.

## RFC 9457 envelope status

ADR 0012 + iam T-006 evidence: documents is **mid-migration**. The codegen bootstrap is in place (`api.gen.go`) but handlers still use legacy `{error:{code,message,details,trace_id}}` shape. Record this as factual current state in the doc; do not aspirationally claim RFC 9457 compliance.

## Open questions deferred to tech-debt / backlog

- `internal/modules/documents/approval/` vs `internal/modules/approval/` — naming collision: which is canonical for approval instances on documents? Will be answered by Phase 2 + Phase 3 artifacts.
- Comments CRUD has handlers (per `wiki/backlog/contract-first-followups.md` "spec gaps") but no spec ops. Surface in tech-debt.
- `renameDocument`, `duplicateDocument` likewise have no spec ops. Surface in tech-debt.
- Idempotency middleware coverage: per ADR 0011, only `/api/v2/controlled-documents` + revisions route are idempotent. `POST /finalize` is NOT idempotent — flag in tech-debt if confirmed by Phase 2.

## Phase skips

None. All 8 phases will run.
