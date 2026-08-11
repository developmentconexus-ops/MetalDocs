# Module: render-fanout

> **Last verified:** 2026-08-11 (ZIP-timestamp normalization seam — `normalizeZipTimestamps` in `packages/eigenpal-adapter/src/index.ts:160` makes `content_hash` reproducible forward-only; new pipeline step 1b + its own section; bounded against ROADMAP unit 4.11) | **Prior:** 2026-07-02 (StagingOutboxWorker consolidation: per-table outbox worker/repo files replaced by generic `staging_outbox_worker.go`/`staging_outbox.go`; APP-01 `pdf_dispatcher.go`/`pdf_dispatch_adapter.go` deleted — key files + failure matrix updated) | **Prior:** 2026-06-29 (ADR 0050 — new `render/domain` package with `ComputedCatalog()` single source of truth; bidirectional parity guard; `approval_date` resolver now returns sentinel pre-approval; prior: 2026-06-12 Wave F — `Enqueue` now fails loud on nil tx; prior: 2026-06-01 P2 consolidation)
> **Status:** active (pipeline module)
> **Maturity:** L2
> **Scope:** DOCX → PDF rendering pipeline, token substitution engine, outbox-driven dispatch.
> **Out of scope:** Approval-triggered freeze invocation (see `modules/approval.md`).
> **Key files:**
> - `internal/modules/render/domain/computed_catalog.go` — **NEW** published domain package: `ComputedToken{Key,Label,Description,AuthorVisible}` + `ComputedCatalog()` returning the 8 canonical computed tokens. Single source of truth per ADR 0050.
> - `internal/modules/render/resolvers/catalog_parity_test.go` — bidirectional parity guard: every registered resolver has a `render/domain` descriptor and vice-versa (drift is impossible at the source)
> - `internal/modules/render/resolvers/builtins.go` — resolver implementations (8 resolvers, all bound to keys declared in `render/domain.ComputedCatalog()`)
> - `internal/modules/render/resolvers/approval_date.go` — `ApprovalDateResolver`: returns `"[aguardando aprovação]"` when the final approval date is zero (draft / pre-approval); authors may safely declare `{approval_date}` in templates
> - `internal/platform/messaging/events.go` - typed event envelope and `PDFConvertPayload`
> - `internal/modules/render/fanout/client.go` — HTTP client to docx-renderer
> - `apps/docx-renderer/src/routes/fanout.ts` — docx-renderer fanout route
> - `internal/modules/render/fanout/staging_outbox.go:33` — generic `StagingOutboxRepository` (allowlisted table binding; `ReadState` at `:179` returns latest outbox status, used by `view_service.go` to report `pdf_status=failed`)
> - `internal/modules/render/fanout/staging_outbox_worker.go:23` — generic `StagingOutboxWorker`: background worker polls + dispatches (PDF + materialize instances)
> - `internal/platform/worker/pdf_job_runner.go` — outbox consumer
> - `apps/docx-renderer/src/render/fanout.ts` — eigenpal headless substitution

## Computed-token catalog (render/domain) — ADR 0050

`render` is the authoritative owner of the computed-token catalog: it defines, resolves, and now
**publishes** the catalog from its `domain` layer.

### `render/domain.ComputedCatalog()`

`internal/modules/render/domain/computed_catalog.go` is the **single source of truth** for all
computed (system-filled) tokens:

| Key | Label (PT-BR) | AuthorVisible |
|---|---|---|
| `doc_code` | Código do documento | true |
| `doc_title` | Título do documento | true |
| `revision_number` | Número da revisão | true |
| `author` | Autor | true |
| `effective_date` | Data efetiva | true |
| `approvers` | Aprovadores | true |
| `controlled_by_area` | Área controladora | true |
| `approval_date` | Data de aprovação | true |

`AuthorVisible: true` on all 8 tokens means the full set is surfaced in the authoring palette.

### Cross-module consumers (legal boundary)

`templates` is the only cross-module consumer.  It imports `render/domain` (the domain layer
only), which is the legal form per `scripts/check-module-boundaries.ps1:52`.  No reverse import
(`render → templates`) exists.  This mirrors the `tokens/domain.DictionaryReader` symmetry.

### Bidirectional parity guard

`internal/modules/render/resolvers/catalog_parity_test.go` (`TestCatalogResolverParity`) asserts:
- every resolver key in the registry has a matching `ComputedCatalog()` descriptor, AND
- every `ComputedCatalog()` key has a matching registered resolver.

The test lives in the same module as both the resolvers and the catalog, so drift is caught
immediately without depending on any external test or consumer.

### `approval_date` — sentinel pre-approval

`internal/modules/render/resolvers/approval_date.go` returns `"[aguardando aprovação]"` when the
workflow reader returns a zero approval date (document not yet approved or still in draft).  This
matches the behaviour of the `approvers` resolver.  Authors can safely place `{approval_date}` in
a template without the resolver erroring on unpublished documents.

## Pipeline (high-level)

1. Freeze service substitutes the 8 fixed tokens in the DOCX (eigenpal-native format).
1b. **The produced ZIP's entry timestamps are normalized before it is hashed** — see the section below. `content_hash` is a SHA-256 over the post-normalization bytes.
2. Frozen DOCX uploaded to MinIO.
3. `DecisionService` enqueues a `pdf_dispatch_outbox` row inside the approval transaction; `StagingOutboxWorker` polls it and publishes the `docgen_v2_pdf` event with `messaging.PDFConvertPayload` (APP-01 2026-07-01: the old post-commit `PDFDispatcher`/`PDFDispatchAdapter` path was deleted — outbox is the only dispatch path).
4. PDFJobRunner picks up the typed payload, calls Gotenberg via docx-renderer.
5. Resulting PDF stored alongside the DOCX in MinIO.

## ZIP-timestamp normalization — why `content_hash` is reproducible at all

A DOCX is a ZIP, and every ZIP entry carries a DOS timestamp. The template engine
wrote each entry with the wall clock at write time, so two renders of byte-identical
content produced different archives whenever they straddled a DOS-timestamp tick
(DOS time has 2-second granularity). `content_hash` is a SHA-256 over the produced
buffer (`apps/docx-renderer/src/render/fanout.ts:65`), so it inherited that
non-determinism directly: the same inputs could hash differently for no reason
attributable to the document.

**The seam.** `normalizeZipTimestamps` in `packages/eigenpal-adapter/src/index.ts:160`
re-opens the archive the engine produced, re-adds every entry with
`date: NORMALIZED_ZIP_DATE` (`1980-01-01T00:00:00.000Z`, the DOS floor) and
`createFolders: false`, and re-serializes DEFLATE at level 6 — the same settings the
engine used. `processTemplate` became `async` to accommodate it. Entry order, entry
paths, directory records and permissions are copied through unchanged, and archive-level
and per-entry ZIP comments are preserved explicitly (dropping them would be silent
metadata loss in the frozen artifact, even though determinism would hold without it).

The claim this buys is narrow and worth stating precisely: decompressed content alone
does not determine ZIP bytes — entry paths, entry order, directory entries, permissions
and serialization settings all do. What the seam removes is the **clock**, the one input
that varied between two renders of the same document. Given the same normalized archive
structure and content, the bytes and therefore the hash are identical.

**The guard.** `apps/docx-renderer/src/render/__tests__/fanout.test.ts` —
`renders byte-identical ZIPs across clock instants with normalized entry dates` renders
twice at different clock instants and asserts full buffer equality, hash equality, and
that every entry date equals the normalized date. The RED failure lands on the
byte-equality assertion, so the proof is attributable to the defect and not to the
date assertion incidentally passing.

**Bound this claim.** Reproducibility holds **forward only**, for renditions produced
after this seam landed. It says nothing about revisions pinned before it, and it does
not make the forensic reconstruction endpoint work — that endpoint has never run
end-to-end and its defects are ordered such that the obvious partial fix is dangerous
(ROADMAP unit 4.11). Historical policy for already-pinned revisions is an open operator
decision. Nothing here backfills or mutates `matches_original`.

## Failure modes

| Failure | Symptom | Detection | Response |
|---|---|---|---|
| Gotenberg / LibreOffice down | PDF outbox rows accumulate in `pdf_dispatch_outbox` with `pending`; published doc shows `pdf_status=pending` for >SLA | `StagingOutboxWorker` logs HTTP error; `StagingOutboxRepository.ReadState` returns `pending`/`failed` | Restart Gotenberg container; failed rows retried by worker until `max_attempts`, then marked `failed` (see [`render-fanout-tech-debt.md`](render-fanout-tech-debt.md)) |
| docx-renderer fanout substitution error | Freeze emits `docgen.substitution_failed`; outbox row may be `failed` | `apps/docx-renderer/src/routes/fanout.ts` returns 5xx; `internal/modules/render/fanout/client.go` surfaces error | Inspect docx-renderer logs for token/resolver mismatch; check `concepts/placeholders.md` 8-token catalog drift |
| Resolver returns empty value for required token | Frozen DOCX renders blank placeholder | `internal/modules/render/resolvers/builtins.go` returns `""`; tracked by `concepts/placeholders.md` | Confirm upstream data populated (e.g. controlled-document fields); add resolver coverage |
| Nil tx passed to `Enqueue` | `Enqueue` returns an error immediately (`"metaldocs.pdf_dispatch_outbox enqueue: tx must not be nil"` / `"metaldocs.materialize_dispatch_outbox enqueue: tx must not be nil"`) — `staging_outbox.go:52-55` | Error propagates to caller; no outbox row written | Fix caller to pass the business transaction; a nil tx would break the transactional-outbox guarantee (Wave F, commit `f698d1fd2`) |
| Outbox replay (worker restart mid-dispatch) | Same `(tenant_id, revision_id)` processed twice | `ON CONFLICT (tenant_id, revision_id) DO NOTHING` on outbox INSERT dedupes; consumer is idempotent | Expected; no operator action |
| MinIO upload fails for frozen PDF | Worker logs S3 PUT error; outbox row marked `failed` | Worker error log + outbox status | MinIO healthcheck; retry by clearing `failed` status or reissuing freeze |

## See also

- [workflows/freeze-and-fanout.md](../workflows/freeze-and-fanout.md) — full pipeline (this is the canonical doc)
- [concepts/placeholders.md](../concepts/placeholders.md) — what the resolvers substitute
- [concepts/freeze-and-hashing.md](../concepts/freeze-and-hashing.md) — content/values/schema hashes

- [render-fanout-tech-debt.md](render-fanout-tech-debt.md)
- [backlog/render-fanout-refactor.md](../backlog/render-fanout-refactor.md)


## 11. Risks & Technical Debt

- Critical: 0
- Major: 2
- Minor: 1

Refactor backlog: [../backlog/render-fanout-refactor.md](../backlog/render-fanout-refactor.md)
