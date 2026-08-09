# Pass 03 — Module Architecture Maps: documents · controlleddocuments · templates · approval

- Date: 2026-08-09
- Baseline: `main@418070bf`, branch `docs/architecture-audit-current-state@9e48a6a1`
- Status: reproduced-current
- Scope: the 4 hotspot modules under `internal/modules/`. Evidence is file:line grounded; package-edge counts are taken from `docs/superpowers/analysis/audit-2026-08-09/module-edge-evidence.txt` (already-computed Go import graph) rather than re-derived.
- Positive exemplar used as the yardstick for every seam below: `internal/modules/documents/application/service.go:144` (`DictionaryValueReader`) + `apps/api/cmd/metaldocs-api/dictionary_reader_adapter.go`. See "The exemplar, in detail" at the end.

---

## 1. `internal/modules/documents`

### 1.1 Current responsibility
Owns the `Document` aggregate — the working/draft/frozen revision lifecycle for a controlled-document instance: creation, autosave/fill-in, freeze/snapshot into an immutable revision, export/view, comments, and the review-due/expiry surfacing feed. It is the largest of the four modules by file count (52 non-test `.go` files) and sits at the center of nearly every other module's dependency graph.

### 1.2 Owned domain concepts (`internal/modules/documents/domain`)
- `Document`, `DocumentStatus`, `ReleaseState`, `ReleaseProjection` — `model.go:11,41,98,115`. `Document` carries `ControlledDocumentID *string` (`model.go:59`) and a `Release` projection sourced from `public.release_generations` (`model.go:78-81`) — i.e. the aggregate's own struct embeds foreign-owned identifiers/projections as bridge fields ("Spec 1" migration 0129 note, `model.go:58`).
- `Session`, `Revision`, `RevisionHistoryItem`, `PendingUpload`, `Checkpoint` — `model.go:132,145,159,170,184`.
- `Comment`, `CommentCreateInput`, `CommentUpdateInput` — `comment.go:12,28,37`.
- `Export`, `ExportResult` — `export.go:6,20`.
- `TemplateSnapshot`, `RevisionRef`, `SnapshotHashes` — `snapshot.go:13,34,41`.
- Ports it **publishes** for others: `ActiveInstanceReader`/`NoopActiveInstanceReader` (`active_instance_port.go:35,45`), `ReviewDueReader`/`NoopReviewDueReader` (`review_due_port.go:43,71`), `ReviewSurfaceWriter`/`NoopReviewSurfaceWriter` (`review_surface_port.go:29,56`), `LifecycleEventEnqueuer` (`notification_events.go:39`), `ApproverEligibilityReader` (`approver_eligibility_port.go:8`).
- 8 sentinel errors (`domain/errors.go`, `domain/snapshot.go:10`): `ErrValidationFailed`, `ErrPlaceholderNotAuthorEditable`, `ErrDictionaryTokenMissing`, `ErrEffectiveDateMissing`, `ErrDocumentNotDraft`, `ErrProfileNotConfigured`, `ErrApprovalRouteMissing`, `ErrSnapshotTemplateNotFound`.

### 1.3 Owned DB tables (repositories INSERT/UPDATE/DELETE)
`documents`, `document_revisions`, `document_comments`, `document_checkpoints`, `document_exports`, `document_placeholder_values`, `autosave_pending_uploads`, `editor_sessions` — all in `internal/modules/documents/infrastructure/repository.go` (14 non-test `BeginTx` call sites: lines 450,796,832,921,958,1059,1100,1207,1423,1549,1584,1731,1939,1977).

### 1.4 Foreign tables read/written
| Table | Owner (inferred) | file:line | R/W | Why |
|---|---|---|---|---|
| `approval_instances` | approval | `application/context_builder.go:49`; `infrastructure/active_instance_reader.go:159`; `infrastructure/repository.go:2054`; `infrastructure/resolver_readers.go:144,179` | R | Active-instance lookups, resolver context for placeholders |
| `approval_stage_instances` | approval | `infrastructure/repository.go:2053`; `infrastructure/resolver_readers.go:142` | R | Stage-count for resolver context |
| `approval_signoffs` | approval | `infrastructure/resolver_readers.go:102,180` | R | Signoff resolver context (e.g. `{{signoff.date}}` placeholders) |
| `release_generations` | approval | `infrastructure/repository.go:380` | R | Release/readiness-hold projection joined onto document rows |
| `documents.controlled_document_id` (own column, FK into `controlled_documents`) | — | `infrastructure/repository.go:221-223,1664`; `application/document_area.go:54`; `application/document_cdid.go:17` | R (own table only) | Never joins `controlled_documents` directly — see 1.9 healthy counter-example |

All foreign reads are confined to `infrastructure/` (repository/resolver adapters) — **not** `application/`. Zero writes to any foreign-owned table were found. Zero foreign reads of `controlled_documents` or `templates_template*` tables — those go through ports (1.9/1.10).

### 1.5 Public application surface (`internal/modules/documents/application`)
Exported symbols (top-level `type`/`func`/`var`/`const` + exported methods, `grep -c`): **311**. Top-level exported types: 22, largest being `Service` (`service.go:152`), `FillInService` (`fillin_service.go:43`), `ExportService` (`export_service.go:40`), `FreezeService` (`freeze_service.go:96`), `ViewService` (`view_service.go:33`), `ReconstructionService` (`reconstruct_service.go:24`), `SnapshotService` (`snapshot_service.go:17`), `CDDocumentInitializer` (`cd_initializer.go:16`), `DocumentContextBuilder` (`context_builder.go:14`).

Consumer-owned ports declared here (interfaces this package defines that foreign infra implements): `Repository` (`service.go:43`), `Presigner` (`service.go:89`), `TemplateReader` (`service.go:100`), `FormValidator` (`service.go:106`), `Audit` (`service.go:112`), `ControlledDocumentDuplicator` (`service.go:119`), `ProfileDefaultTemplateReader` (`service.go:136`), **`DictionaryValueReader` (`service.go:144`, the exemplar)**, `CapabilityChecker` (`ports.go:12`), `ExportRepo`/`ExportPresigner`/`DocgenPDFClient` (`export_service.go:16,26,34`), `SchemaReader`/`FillInWriter`/`FillInReader` (`fillin_service.go:22,34,299`), `FreezeFinalizer`/`SnapshotReader`/`RevisionBodyReader`/`FanoutClient`/`ResolverContextBuilder` (`freeze_service.go:30,39,57,74,121`), `IAMUserOptionsReader` (`iam_user_options.go:13`), `ReconstructionRunner` (`reconstruct_service.go:17`), `SnapshotTemplateReader` (`snapshot_service.go:12`), `ViewPresigner`/`PDFOutboxStateReader` (`view_service.go:21,27`).

### 1.6 Public domain surface
`DocumentStatus`/status constants, `ActiveInstanceView`/`ActiveInstanceReader`, `ReviewDueView`/`ReviewDueReader`, `SurfacedDoc`/`ReviewSurfaceWriter`, `LifecycleEventArgs`/`LifecycleEventEnqueuer`/5 `EventType*` constants (`notification_events.go:29-35`, consumed by approval — see 4.7), `CanTransitionDocumentStatus` (state-legality predicate, consumed directly by approval — see 4.4/4.7), `ErrProfileNotConfigured`, `ErrApprovalRouteMissing`, `ErrDocumentNotDraft`.

### 1.7 Inbound dependencies (who imports `documents`)
From edge evidence: **approval** (`approval/application` → `documents/application`+`documents/domain`; `approval/http`, `approval/jobs` → `documents/domain`), **controlleddocuments** (`controlleddocuments`, `controlleddocuments/infrastructure` → `documents/domain`), **notifications** (`notifications/infrastructure` → `documents/domain`), **jobs** (`jobs/document_review_surfacer` → `documents/domain`), platform `docgenv2` → `documents/application`+`domain`, plus `apps/api`, `apps/jobs`, `apps/worker`, `scripts/release-backfill/backfill`, `cmd/problem-codes-dump`, `composition/tenantdata/registry`.

### 1.8 Outbound dependencies
`controlleddocuments/domain` (heavy — 1.9), `iam/authz`+`iam/domain` (9 edges), `render/fanout`+`render/resolvers` (5 edges), `taxonomy/domain` (2 edges), `templates/domain` (heavy — 1.10), plus platform: `db`, `httpresponse`, `httprouter`, `objectstore`, `pagination`, `ratelimit`, `servicebus`, `sqlescape`, `tenant`, `tenantdata`, `apibase`, `problem`.

### 1.9 Seam: documents → controlleddocuments (T + one clean C)
This is the module pair ADR 0093 (issue #94/A9) says are **not** peer contexts, and the code backs that up with two contradictory patterns living side by side:

- **Healthy (C):** `application/document_area.go:43` `LoadDocumentAreaCode` takes `controlleddocumentsdomain.CDFieldReader` (a narrow, producer-published, 2-method port — `controlleddocuments/domain/cd_field_reader_port.go:23`, explicitly governed by ADR-0039 D1/D3(b)) instead of joining `controlled_documents`. The doc comment at `document_area.go:35-42` names the anti-pattern it is avoiding.
- **Unhealthy (T):** `application/service.go:458` calls `controlleddocumentsdomain.Resolve(...)` — a **pure domain function owned by controlleddocuments** (`controlleddocuments/domain/resolution.go:56`) that decides template-version eligibility for a controlled document — directly from documents' `Service`, passing/receiving controlleddocuments' concrete `TemplateResolutionInput`/`TemplateVersionCandidate` structs (`resolution.go:20,29`). documents also imports the `ControlledDocument` struct itself, `CDStatusActive`, `ErrCDNotFound`, `ErrCDNotActive`, `DocumentRef`, `CloneTemplateRequest`/`NewCloneTemplateRequest`, `DocumentInitializer` (full symbol list: `internal/modules/documents/module.go:9`, `application/service.go:12`, `application/cd_initializer.go:8`, `application/reconstruct_service.go:8`, `application/fillin_service.go:13`, `application/fillin_authz.go:8`, `delivery/http/handler.go:16`, `infrastructure/repository.go:15`).

Seam classification: **T (foreign type coupling) + E** (errors.Is on `ErrCDNotFound`/`ErrCDNotActive`/`ErrProfileHasNoDefaultTemplate`, `application/service.go:1315,1345,1350`). Consumer? documents. Producer? controlleddocuments. Who should own the contract: ambiguous by design — `Resolve` is a pure decision that logically belongs to whichever module owns "which template version applies," and the fact that **documents calls a foreign pure function with foreign DTOs instead of a port** is itself evidence the two are one aggregate split across a package boundary (this is the strongest single piece of code evidence for #94/A9: template-resolution business logic is invoked cross-module by value, not through an interface). Does the consumer define only what it needs? No — it takes the full `ControlledDocument` struct and a pure function, not a narrow interface. Adapter in composition root? No — direct package import, not injected. Reciprocal? Yes — controlleddocuments also imports `documents/domain` (1.9 reverse in §2.9). Should be: at minimum a consumer-owned `TemplateResolver` port on documents' side (mirroring `CDFieldReader`'s discipline); at the ADR-0093 level, the recommended global-maximum is collapsing the `Resolve` logic and the shared `ControlledDocument`/`Document` bridge fields into a single bounded context rather than patching another port.

### 1.10 Seam: documents → templates (T, pervasive)
`templatesdomain.Placeholder` and its kind constants (`PHText`, `PHSelect`, `PHNumber`, `PHDate`, `PHUser`, `PHComputed`, `PHDictionary`, `PHPicture`) are used ~50 times across `application/fillin_service.go`, `application/context_builder.go`, and related files (representative: `fillin_service.go:105-147,180-236`) — documents' fill-in engine is built directly around templates' placeholder schema domain type, not a documents-local projection of it. Alongside this, `templatesdomain.TemplateVersionPort` (`service.go:100`-wrapped `TemplateReader`, backed by `templates/domain/template_version_port.go:7` — `IsPublished`, `GetTemplateVersionState`, `PlaceholderSchema`) **is** a clean producer-published port. Classification: **mixed C (for state/schema reads via `TemplateVersionPort`) + T (for the `Placeholder` schema type itself, which is consumed as templates' native domain type rather than translated at the boundary)**. No `module:templates → module:documents` edge exists in the evidence file — this direction is one-way, not reciprocal.

### 1.11 Seam: documents → approval (thin outbound, heavy inbound — asymmetric)
Outbound from documents is thin: `delivery/http/handler.go:15` imports `approvalapp` only for `errors.Is(err, approvalapp.ErrRevisionTitleRequired)`; `infrastructure` imports `approvaldomain.InstanceInProgress` (a status constant) only. Classification: **E** (sentinel/constant coupling), minor. Compare to §4.4/4.7 where approval's inbound dependency on documents is far heavier (calls documents' own state-transition-legality function and writes the `documents` table directly). The reciprocal pair is real but **not symmetric in severity** — this asymmetry is itself a finding: the fix belongs on approval's side, not documents'.

### 1.12 Producer-owned ports it exposes for others
`ActiveInstanceReader` (consumed by controlleddocuments, `controlleddocuments/module.go:54`), `LifecycleEventEnqueuer`+`LifecycleEventArgs` (implemented by **approval**'s `jobs/lifecycle_event_enqueuer.go:33,42` — approval emits documents-owned lifecycle events at its own mutation sites), `ReviewDueReader`, `ReviewSurfaceWriter`, `ApproverEligibilityReader`, `CanTransitionDocumentStatus`.

### 1.13 Transaction participation
Repository (`infrastructure/repository.go`) manages its own `*sql.DB.BeginTx` directly (14 call sites, §1.3) rather than composing through `platform/db.TxRunner`. Only 6 `TxRunner`-pattern references found module-wide vs. approval's 20/0 split (§4.13) — documents is the module most reliant on ad hoc `BeginTx`, which is relevant to #92/A5 (persistence pattern consistency).

### 1.14 Events emitted / async jobs
Emits via `platform/servicebus`/outbox from `application/export_service.go`, `freeze_service.go`, `service.go`, `view_service.go`, wired in `module.go`. Owns 2 River jobs: `jobs/orphan_pending_sweeper.go`, `jobs/session_sweeper.go`.

### 1.15 HTTP routes owned
`delivery/http/routes_generated.go` (oapi-codegen, contract-first — `api/` dir has `api.gen.go`/`cfg.yaml`/`gen.go`) + `export_handler.go`, `fillin_handler.go`, `handler.go`, `placeholder_options_handler.go`, `reconstruct_handler.go`, `view_handler.go`. Handler method count: **59**.

### 1.16 Test coverage shape
75 test files: 21 integration-tagged (`//go:build integration`), 54 unit.

---

## 2. `internal/modules/controlleddocuments`

### 2.1 Current responsibility
Owns the numbered `ControlledDocument` slot: `Module` docstring (`module.go:1-10`) states it directly — "owns the numbered ControlledDocument slot (`metaldocs.controlled_documents`) that binds a (profile, area) pair to a chain of documents-module revisions," depending on taxonomy, documents, and templates "purely through published ports — never their repositories or SQL" (an intent the code only partially satisfies — see 2.9/2.10 vs. §1.9).

### 2.2 Owned domain concepts
- `ControlledDocument` — `controlled_document.go` (with `CDStatusActive` etc.).
- `DocumentInitializer` port, `CloneTemplateRequest`/`NewCloneTemplateRequest` — `document_initializer.go:14-19`.
- `TemplateResolutionInput`/`TemplateVersionCandidate`/`TemplateResolutionResult`/`Resolve` — `resolution.go:20,29,37,56` (pure function, see §1.9).
- `CDFieldReader`/`NoopCDFieldReader` — `cd_field_reader_port.go:23,41` (the module's own well-governed narrow port, ADR-0039 D1/D3(b)).
- `VisibilityScope`, sequence types — `visibility.go`, `sequence.go`.
- 6 error sentinels via `var (...)` blocks (no top-level `var Err` singles): `ErrManualCodeReasonRequired`, `ErrApprovalRouteMissing` (`controlled_document.go:62,77`), `ErrCloneTemplateNameRequired`, `ErrDictionaryTokenMissing` (`document_initializer.go:14,19`), `ErrProfileHasNoDefaultTemplate` (`resolution.go:44`), `ErrVisibilityScopeInvalid` (`visibility.go:32`); plus 4 more in `application`: `ErrCreationContextUnconfigured`, `ErrTemplateArtifactMissing`, `ErrTemplateArtifactInvariantUnconfigured`, `ErrActorMissing` (`application/creation_context.go:31`, `application/service.go:130,135,141`).

### 2.3 Owned DB tables
`controlled_documents`, `cd_sequence_counters`, `controlled_document_area_grants`, `controlled_document_user_grants` — `infrastructure/repository.go` (1 non-test `BeginTx` at line 385).

### 2.4 Foreign tables read/written
**None found.** No SQL referencing `documents`, `approval_*`, or `templates_*` tables inside `internal/modules/controlleddocuments` (grep for those table names in SQL-verb context returned zero non-comment hits). Every cross-module read goes through a Go-level port (2.9/2.10). This is the cleanest of the four modules on the SQL-coupling axis (#92/A5 — zero S-classified seams).

### 2.5 Public application surface (`internal/modules/controlleddocuments/application`)
Exported symbols: **139**. Top-level types: `ControlledDocumentService` (`service.go:65`), plus consumer-owned ports `TemplateVersionChecker`, `ProfileReader`, `AreaReader` (`service.go:35,41,47`), `ProfileLister`, `AreaLister` (`creation_context.go:17,23`), commands `CreateControlledDocumentCmd`/`CreateResult`/`CreateRevisionCmd` (`service.go:100,122,952`).

### 2.6 Public domain surface
`ControlledDocument`, `CDStatusActive`/status constants, `DocumentRef`, `CloneTemplateRequest`, `ErrCDNotFound`, `ErrCDNotActive`, `ErrProfileHasNoDefaultTemplate`, `Resolve`/`TemplateResolutionInput`/`TemplateVersionCandidate` (consumed directly by documents, §1.9), `CDFieldReader`/`NoopCDFieldReader` (the module's healthy published port), `RouteReadinessReader` is **not** this module's — it's approval's, consumed here (2.9).

### 2.7 Inbound dependencies
**documents** (`documents`, `documents/application`, `documents/delivery/http`, `documents/infrastructure` → `controlleddocuments/domain`, heavy — §1.9), plus `apps/api`, `composition/tenantdata/registry`, `cmd/problem-codes-dump`.

### 2.8 Outbound dependencies
**approval** (`approvaldomain.RouteReadinessReader` — 2.9), **documents** (`documentsdomain.ActiveInstanceReader` — module.go:54), **iam** (`authz`, `domain`, `AreaCapabilityReader` — module.go:89), **taxonomy** (`domain` — `ProfileReader`/`AreaReader`/`GovernanceLogger` backing), plus platform `db`, `authn`, `httprouter`, `httpresponse`, `idempotency`, `pagination`, `problem`, `sqlescape`, `tenant`, `tenantdata`, `apibase`.

### 2.9 Seam: controlleddocuments ← documents, controlleddocuments ← approval (the healthy pattern, producer-published-port variant)
`module.go:47-90` (`Dependencies` struct) is the composition-root injection point: `ActiveInstanceReader documentsdomain.ActiveInstanceReader` (documents-published port), `RouteReadinessReader approvaldomain.RouteReadinessReader` (approval-published port, backing "does this profile have an active approval route" — the hard creation gate D2, `module.go:72-77`), `AreaCapabilityReader iamdomain.AreaCapabilityReader`. All three are **producer-owned published ports** (interface lives in the producer's `domain` package, not consumer-defined) — a valid, explicitly-documented sibling variant of the exemplar (`module.go:36-46` names each one's origin). `New()` panics if any required dependency is nil (`module.go:100-124`) — fail-loud composition, no silent no-op except the explicitly-chosen `NoopActiveInstanceReader` default. This is the strongest evidence in the whole audit that the codebase *knows* how to do seams correctly when it wants to.

Contrast: `application/service.go:23` and `creation_context.go:8` import `approvaldomain` directly for `RouteReadinessReader` — consistent with the port pattern, not raw domain-internals reach-through.

### 2.10 Seam: controlleddocuments → documents (T, one call site)
`documents/module.go:9` / `infrastructure/repository.go:15` show documents exposing `controlleddocumentsdomain` back (§1.9's mirror); on the controlleddocuments side, `infrastructure/repository_test.go` and `repository_like_escape_test.go` import `documentsdomain` for parity, and `module.go:22,54` consume `documentsdomain.ActiveInstanceReader`/`NoopActiveInstanceReader` as a straightforward published-port consumption (**C**, not T — no documents domain *struct* is imported by controlleddocuments, only the port interface + its Noop).

### 2.11 Consumer-owned vs producer-owned ports
Consumer-owned (defined in controlleddocuments/application, implemented elsewhere): `TemplateVersionChecker`, `ProfileReader`, `AreaReader`, `ProfileLister`, `AreaLister` (§2.5). Producer-owned it exposes: `CDFieldReader` (§2.2, consumed by documents §1.9), `DocumentInitializer` (consumed by documents, wired post-construction per `module.go:96-99` to break an init cycle).

### 2.12 Transaction participation
1 non-test `BeginTx` (`infrastructure/repository.go:385`); `application.NewControlledDocumentService` takes `platformdb.NewTxRunner(deps.DB)` (`module.go:136`) — TxRunner is the primary pattern here, unlike documents.

### 2.13 Events / async jobs
No `platform/messaging`/`servicebus`/`outbox` references found. No `jobs/` subpackage. controlleddocuments does not emit outbox events directly.

### 2.14 HTTP routes owned
`delivery/http/handler.go`, `routes.go`. Handler method count: **13**.

### 2.15 Test coverage shape
22 test files: 7 integration-tagged, 15 unit.

---

## 3. `internal/modules/templates`

### 3.1 Current responsibility
Owns the `Template`/`TemplateVersion` aggregate: docx upload, placeholder-schema authoring/validation, versioning, and publish lifecycle — including submitting a template version into the approval kernel and reacting to its decisions.

### 3.2 Owned domain concepts
- `Template`, `TemplateVersion`, `VersionStatusPublished`/`VersionStatusObsolete` etc. — `template.go`, `version.go`.
- `Placeholder` + kind constants `PHText`/`PHSelect`/`PHNumber`/`PHDate`/`PHUser`/`PHComputed`/`PHDictionary`/`PHPicture` — `schemas.go` (consumed pervasively by documents, §1.10).
- `TemplateVersionPort` — `template_version_port.go:7` (the module's own clean producer-published port).
- `ReadModel` types — `read_model.go`.
- 20 sentinel errors — `domain/errors.go` (full list in the earlier grep pass): `ErrISOSegregationViolation`, `ErrForbidden`, `ErrUploadMissing`, `ErrUploadTooLarge`, `ErrPlaceholderIDEmpty`, `ErrDuplicatePlaceholderID`, `ErrPlaceholderNameInvalid`, `ErrDuplicatePlaceholderName`, `ErrInvalidConstraint`, `ErrPlaceholderCycle`, `ErrUnknownResolver`, `ErrPlaceholderNotInCatalog`, `ErrPlaceholderNotComputed`, `ErrPlaceholderReservedName`, `ErrPlaceholderDictionaryInvalid`, `ErrDocTypeCodeRequired`, `ErrApprovalRouteMissing`, `ErrContentMaterializationFailed`, `ErrTransactionRequired`, `ErrConcurrentTransition`.

### 3.3 Owned DB tables
`templates_template`, `templates_template_version` — `infrastructure/postgres.go` (INSERT/UPDATE at lines 64,227,271,309,439,472,534,598,651,678), `infrastructure/approval_completion_writer.go:78,159,189` (CAS UPDATE from the approval-decision completion path — see 3.9), `infrastructure/tenant_data_port.go:75`.

### 3.4 Foreign tables read/written
**None found.** No SQL against `documents`, `controlled_documents`, or `approval_*` tables inside `internal/modules/templates`. All cross-module collaboration is Go-level (3.9).

### 3.5 Public application surface (`internal/modules/templates/application`)
Exported symbols: **154**. Top-level types: `Service` (`service.go:20`), commands `CreateTemplateCmd`/`CreateTemplateResult`/`CreateVersionCmd` (`create.go:19,30,172`), `PresignAutosaveCmd`/`PresignAutosaveResult`/`PresignTemplateUploadCmd`/`CommitAutosaveCmd` (`autosave.go`), `ArchiveCmd`/`PublishTemplateVersionCmd`/`PublishTemplateVersionResult` (`lifecycle.go`), `UpdateSchemasCmd` (`schema.go:19`), ports `Repository`/`Presigner`/`Clock`/`UUIDGen`/`ResolverRegistryReader` (`ports.go:18,56,71,75,80`).

### 3.6 Public domain surface
`Placeholder` + kind constants (heavily consumed by documents), `VersionStatusPublished`/`VersionStatusObsolete`, `TemplateVersionPort`, `ErrApprovalRouteMissing`.

### 3.7 Inbound dependencies
**documents** (`documents`, `documents/application`, `documents/delivery/http`, `documents/infrastructure` → `templates/domain`), `docgenv2` platform package does not touch templates directly (it touches documents), `apps/api`, `composition/tenantdata/registry`, `cmd/problem-codes-dump`. No inbound edge from approval or controlleddocuments was found in the evidence file — templates is not consumed by either sibling module.

### 3.8 Outbound dependencies
**approval** — the heaviest outbound edge templates has (3.9): `templates/application → approval/domain` (4 edges incl. `application`, `delivery/http`, `infrastructure`), `iam` (`authz`, `domain`, 5 edges), `render/domain` (2 edges), plus platform `db`, `objectstore`, `apibase`, `httpresponse`, `httprouter`, `idempotency`, `problem`, `tenant`, `tenantdata`.

### 3.9 Seam: templates → approval (C-with-DTO-leak, one-directional)
`delivery/http/handler.go:31-51` declares three narrow, **consumer-owned** interfaces explicitly modeled on approval's own internal seam pattern (comment cites `internal/modules/approval/http/handler.go`): `approvalSubmitService` (`SubmitTemplateVersionForReview`, `PreviewRoute`), `approvalDecisionService` (`RecordSignoff`), `approvalReadService` (`LoadActiveInstanceBySubjectForMutation`). This is genuinely **C**-shaped at the interface level — but every method signature is typed in terms of approval's own DTOs (`approvalapp.TemplateSubmitRequest`, `approvalapp.TemplateSubmitResult`, `approvalapp.RoutePreview`, `approvalapp.SignoffRequest`, `approvalapp.SignoffResult`, `approvaldomain.Instance`), so the package import (`templates/delivery/http → approval/application` + `approval/domain`) exists regardless of the interface indirection — a "leaky C": the seam shape is right, the payload types are not translated at the boundary the way the exemplar translates `tokensdomain.ErrNotFound`.

`infrastructure/approval_completion_writer.go` is the reverse-direction half of the same seam: it is templates' own CAS-UPDATE adapter that **approval calls into** to mark a template version's approval outcome (`UPDATE templates_template_version ... approver_id`, `approval_completion_writer.go:78,159,189`) — i.e. approval writes into a templates-owned table through a templates-provided writer rather than raw SQL. This is the one place in the whole audit where the `documents`-style raw-SQL-into-a-foreign-table pattern (§4.4/4.7) was **not** taken; worth naming as the "template completion writer" precedent when designing the fix for approval→documents.

Errors.Is on ~10 foreign sentinels (`ErrContentHashMismatch`, `ErrDuplicateSubmission`, `ErrIdempotencyKeyRequired`, `ErrInstanceCompleted`, `ErrNoActiveApprovalRoute`, `ErrNoActiveInstance`, `ErrStageNotActive`, `ErrTemplateVersionNoContent`, `ErrTemplateVersionNotDraft`, `ErrTemplateVersionNotFound`) — **E** classification, heavy but at least confined to the delivery/http translation layer. Consumer: templates. Producer: approval. Reciprocal? **No** — approval does not import `templates/*` (only a comment reference at `approval/http/router.go:19`); the completion-writer direction is templates exposing a *port implementation for approval to call*, not templates importing approval further.

### 3.10 Consumer-owned vs producer-owned ports
Consumer-owned: `approvalSubmitService`, `approvalDecisionService`, `approvalReadService` (§3.9, delivery/http layer — unusual placement; most modules put consumer ports in `application`, not `delivery/http`). Producer-owned exposed: `TemplateVersionPort` (§3.2, consumed by documents), the template-version completion writer (§3.9, consumed by approval).

### 3.11 Transaction participation
2 `TxRunner` references, 0 direct `BeginTx` in production code — cleanest tx discipline of the four modules alongside controlleddocuments.

### 3.12 Events / async jobs
`platform/messaging`/`servicebus` reference in `application/create.go`. 1 River job: `jobs/orphan_object_sweeper.go`.

### 3.13 HTTP routes owned
`delivery/http/handler.go` + generated `api/`. Handler method count: **40**.

### 3.14 Test coverage shape
37 test files: 4 integration-tagged, 33 unit.

---

## 4. `internal/modules/approval`

### 4.1 Current responsibility
The governance kernel: subject-generic (`subject_kind`, `subject_key`) approval routes, stages, quorum, delegation, sign-off/e-signature, review verdicts, release evaluation, and the SLA/obsolete/cancel lifecycle. Promoted from nested `documents/approval` to the 15th top-level module by ADR 0082 (supersedes ADR 0072). By file count (102 non-test `.go` files) and application-package exported-symbol count (740 incl. methods) it is now the largest and most heavily-coupled module in the system — larger than documents.

### 4.2 Owned domain concepts
`Instance`, `StageInstance`, `StageActive`, `InstanceInProgress` — `instance.go`. `Route`, route-stage types — `route.go`. `Signoff` — `signoff.go`. `Delegation` — `delegation.go`. `Quorum` — `quorum.go`. `Release` — `release.go`. `ReviewVerdict` — `review_verdict.go`. `Selector`/`ActorSelector` — `selector.go`. `Subject`/`SubjectKindTemplate` (subject-generic per ADR 0082/0083) — `subject.go`. `SoD` (segregation of duties) — `sod.go`. `Viewer` — `viewer.go`. Published ports: `RouteReadinessReader` (`route_readiness_reader_port.go`, consumed by controlleddocuments — §2.9), `ReleaseHoldReader`(`release_hold_port.go`), `SLAPort`(`sla_port.go`). 22 sentinel errors (`domain/*.go`, listed in §Evidence pass above) — the largest error surface of the four modules.

### 4.3 Owned DB tables
`approval_instances`, `approval_stage_instances`, `approval_signoffs`, `approval_review_verdicts`, `approval_delegations`, `approval_routes`, `approval_route_stages`, `approval_route_stage_selectors` — `infrastructure/postgres_approval_repository.go` (writes at lines 71,182,218,916,1390,1428,1452,1491,1812,2243,2262) and `application/route_admin_service.go` (writes at lines 362,665,678,706,718,847,1045,1108 — **route-table SQL lives in `application/`, not `infrastructure/`**, an internal layering deviation from the module's own repository pattern). Also `auth_failure_counters` — `infrastructure/signature/postgres_auth_failure_rate_limiter.go:64,86` (e-signature rate limiting).

### 4.4 Foreign tables read/written — the headline finding
| Table | Owner | file:line (representative, not exhaustive) | R/W | Why |
|---|---|---|---|---|
| `documents` | documents | `application/cancel_service.go:171`; `application/decision_service.go:767`; `application/document_terminal_approval.go:129`; `application/mark_reviewed_service.go:191`; `application/obsolete_service.go:137`; `application/release_coordinator.go:350,484`; `application/review_verdict_service.go:470`; `application/submit_service.go:557`; `infrastructure/postgres_approval_repository.go:916` | **W** (`UPDATE documents SET status=..., revision_version=revision_version+1 WHERE id=$1 AND tenant_id=$2 AND status=$old`) | Approval decisions (submit/approve/reject/publish/obsolete/cancel) drive the document's own status column, OCC-guarded by an explicit `WHERE status = <expected>` and backstopped by a DB trigger (comment at `decision_service.go:758-765` explicitly names the trigger as the real enforcement) |
| `documents` | documents | `application/context_builder.go:42`; `application/document_area.go:37`; `application/document_cdid.go:17`; `application/read_service.go:717` | R | Area/CD-id context, plus `LEFT JOIN controlled_documents` in the same read-service query |
| `controlled_documents` | controlleddocuments | `application/read_service.go:717`; `application/release_coordinator.go:181`; `infrastructure/postgres_approval_repository.go:630,2024` | R only (no write found) | Profile/area context for inbox and release evaluation |
| `document_comments` | documents | `infrastructure/postgres_approval_repository.go:1775,1793` | R | Comment context surfaced alongside a decision |

**This is the module's central architecture problem.** Unlike documents' foreign reads (confined to `infrastructure/`, read-only, all via `SELECT`), approval's foreign access to `documents` is (a) a **write**, (b) issued from **8 different `application/` files**, not centralized in one repository method, and (c) paired with a direct call to `docsdomain.CanTransitionDocumentStatus` (documents' own state-legality predicate — `application/cancel_service.go:763`, imported as `docsdomain` in 8 files: `cancel_service.go`, `decision_service.go`, `document_terminal_approval.go`, `mark_reviewed_service.go`, `obsolete_service.go`, `release_coordinator.go`, `route_preview.go`, `services.go`, `submit_service.go`). approval both *reasons about* and *mutates* documents' aggregate state directly — two modules performing DML on the same row with no single owner beyond a DB trigger as the last line of defense. Contrast with templates' `approval_completion_writer.go` (§3.9), which is the module-provides-a-writer pattern approval should be consuming from documents instead of raw SQL.

### 4.5 Public application surface (`internal/modules/approval/application`)
Exported symbols: **740** (largest of the four). ~20 top-level service types, the largest concentration of any module: `MarkReviewedService`, `DelegationService`, `ReadService`, `ObsoleteService`, `ReleaseCoordinator`, `ReviewVerdictService`, `FastForwardService`, `ReleaseFactRecorder`/`ArtifactFactWriter`, `CancelService`, `DecisionService`, `RouteAdminService`, `Services` (top-level aggregate, `services.go:34`), `SLAExtensionService`, `SubmitService`, `TemplateSubmitService` (consumed by templates, §3.9). Consumer-owned ports: `ProfilePolicyReader` (`ports.go:20`), `EventEmitter`/`MemoryEmitter` (`events.go:73,107`), `ReleaseEvaluationEnqueuer` (`release_facts.go:42`), `ProfileReviewIntervalReader` (`release_coordinator.go:44`), `PinInvoker`/`TemplateCompletionWriter` (`decision_service.go:47,72` — the latter is the port that calls templates' completion writer, §3.9), `RouteAdminReplayCommitter`/`RouteAdminIdempStore` (`route_admin_idemp.go:24,32`), `TerminalApprovalReleaseRecorder` (`release_terminal_approval.go:83`), `SignoffReplayCommitter` (`signoff_idemp.go:25`), `SubmitDefaultsResolver` (`submit_defaults.go:18`), `TemplateVersionReader`/`TemplateVersionSubmitWriter` (`template_submit_service.go:28,77`).

### 4.6 Public domain surface
`Instance`, `StageInstance`, `Decision`/`DecisionApprove`/`DecisionReject`, `ActorSelector`, `Subject`/`SubjectKindTemplate`, `RouteReadinessReader` (§2.9), plus ~10 error sentinels consumed by templates (§3.9).

### 4.7 Inbound dependencies
`apps/api` (6 package edges), `apps/jobs` (3), `apps/worker` (2), `composition/tenantdata/registry`, **jobs** module (`approval_sla_surfacer`, `release_hold_reconciler`, `stuck_instance_watchdog` → `approval/application`/`domain`), **notifications** (`infrastructure → approval/domain`), **taxonomy** (`taxonomy`, `taxonomy/application → approval/domain`), **templates** (§3.9), **controlleddocuments** (§2.9, `RouteReadinessReader`), `scripts/release-backfill/backfill`, `cmd/problem-codes-dump`.

### 4.8 Outbound dependencies
**documents** (§4.4, heaviest and most invasive), **controlleddocuments** (`CDFieldReader`/`NoopCDFieldReader` only — §2.9, clean), **taxonomy** (`taxonomy/domain`, 3 edges: `application`, `domain` itself, `infrastructure`), **iam** (`authz`, `domain`, 7 edges — heaviest iam consumer of the four modules), plus platform `apibase`, `db` (5 edges), `httprouter`, `idempotency`, `legacystatus`, `pagination`, `passwordhash` (e-signature), `problem`, `strictjson`, `tenant`, `tenantdata`.

### 4.9 Seam: approval → taxonomy (reciprocal, not deep-dived — noted per known facts)
`application → taxonomy/domain`, `domain → taxonomy/domain`, `infrastructure → taxonomy/domain` (3 edges); reverse `taxonomy → taxonomy/application → approval/domain` (2 edges) confirms the reciprocal pair from the known-facts list. Out of scope for this pass's deep-dive (taxonomy is not one of the 4 hotspot modules) but flagged for #93/A4 as another reciprocal edge worth auditing in a follow-up pass.

### 4.10 Consumer-owned vs producer-owned ports
Consumer-owned: the ~15 ports in §4.5 (largest set of any module — approval pulls in more foreign capabilities as narrow interfaces than any other module, which is architecturally correct *when* it stays at the interface). Producer-owned exposed: `RouteReadinessReader` (§2.9, clean), `CDFieldReader`-consuming pattern is one-way (approval only consumes controlleddocuments' port, doesn't expose one back).

### 4.11 Transaction participation
20 `TxRunner`-pattern references, **0** non-test direct `BeginTx` — the best tx discipline of the four modules by this metric, which makes the raw-SQL-into-`documents` finding (§4.4) more notable: the module consistently uses the shared transaction abstraction for its own tables but still reaches around the module boundary for `documents`.

### 4.12 Events / async jobs
Heaviest outbox/messaging user: `decision_service.go`, `events.go`, `mark_reviewed_service.go`, `obsolete_service.go`, `release_facts.go`, `services.go`, `submit_service.go`, `template_submit_service.go`, `domain/notification_events.go`. 4 jobs files: `approval_notification_enqueuer.go`, `lifecycle_event_enqueuer.go` (implements documents' `LifecycleEventEnqueuer` port, §1.12), `release_evaluate_args.go`, `release_evaluate_job.go`.

### 4.13 HTTP routes owned
`http/` (not `delivery/http` like the other three — naming inconsistency) + `http/contracts/`. Handler method count: **50**.

### 4.14 Test coverage shape
100 test files: 27 integration-tagged, 73 unit — largest test surface of the four, proportionate to its size.

### 4.15 AuthZ surface (#89/A8)
32 `authz.Require(` call sites — most of any module (documents 21, controlleddocuments 8, templates 6). 17 distinct `Cap*` capability constants referenced — again the most, consistent with approval owning the widest action surface (submit/sign/delegate/route-admin/cancel/obsolete/SLA-extend). No role-based conditionals found in a quick scan; capability-only reasoning appears intact per ADR 0022.

---

## 5. Cross-cutting: the exemplar, in detail

`documents/application/service.go:144`:
```go
type DictionaryValueReader interface {
    Lookup(ctx context.Context, tenantID, name string) (string, bool, error)
}
```
and `apps/api/cmd/metaldocs-api/dictionary_reader_adapter.go:14-27`:
```go
type dictionaryValueReaderAdapter struct {
    reader tokensdomain.DictionaryReader
}
func (a dictionaryValueReaderAdapter) Lookup(ctx context.Context, tenantID, name string) (string, bool, error) {
    e, err := a.reader.GetByName(ctx, tenantID, name)
    if err != nil {
        if errors.Is(err, tokensdomain.ErrNotFound) {
            return "", false, nil
        }
        return "", false, err
    }
    return e.Value, true, nil
}
```
Why this is the healthy pattern, checked against every axis this audit uses:
1. **Interface lives in the consumer** (`documents/application`), typed only in primitives (`string, bool, error`) — no `tokensdomain` type crosses into documents at all.
2. **Zero import of `tokens` anywhere in `internal/modules/documents`** (verified: `grep -rln "modules/tokens" internal/modules/documents` returns nothing) — invariant #6 (`dictionary_reader_adapter.go:12`) is mechanically true, not just claimed.
3. **Adapter lives at the composition root** (`apps/api/cmd/metaldocs-api`), not in either module.
4. **Foreign error is translated at the boundary**: `tokensdomain.ErrNotFound` becomes `found=false` — the consumer never does `errors.Is` on a foreign sentinel (contrast with §1.9/§1.11/§3.9 where `errors.Is` on foreign sentinels is the norm, not the exception).
5. **No reciprocal dependency**: tokens does not import documents.

Every other seam audited in §1–§4 deviates from at least one of these five properties. The two closest matches are controlleddocuments' `Dependencies` struct (§2.9 — properties 1,3,5 hold; property 2 doesn't apply the same way since it's a producer-published port, not consumer-defined, but the discipline is equally rigorous) and templates' `approvalSubmitService`/`approvalDecisionService`/`approvalReadService` (§3.9 — property 1 holds, property 2 fails because payload DTOs are approval's own types).

---

## 6. Top-10 findings mapped to owning issue

1. **[#94/A9, severe]** approval directly mutates the `documents` table via raw `UPDATE documents SET status=...` from **8 separate `application/` files** (§4.4), while also calling documents' own `CanTransitionDocumentStatus` state-legality function (§4.4). Two modules perform DML on one aggregate root with only a DB trigger as the tiebreaker. This is the single strongest piece of evidence that documents/approval (and by extension documents/controlleddocuments/templates) are not cleanly separable bounded contexts under the current split — directly substantiates ADR 0093.
2. **[#94/A9, severe]** documents calls controlleddocuments' pure domain function `Resolve()` directly with controlleddocuments' own DTOs (`service.go:458`, §1.9) instead of a port — business logic invoked cross-module by value.
3. **[#94/A9, high]** documents' fill-in engine consumes templates' `Placeholder` domain type and all 8 kind constants natively (~50 call sites, §1.10) rather than through a documents-local projection — the placeholder schema is effectively shared, not owned by either side.
4. **[#93/A4, high]** The reciprocal edge documents↔approval is **asymmetric in severity**: documents' outbound coupling to approval is thin (2 sentinel/constant references, §1.11); approval's outbound coupling to documents is deep (state function + 8 raw-SQL write sites, §4.4/4.8). A seam audit that only counts "is it reciprocal" without weighing direction would under-report this.
5. **[#93/A4, medium]** templates' `approvalSubmitService`/`approvalDecisionService`/`approvalReadService` (§3.9) get the *interface* shape right (consumer-owned, narrow) but leak approval's own DTOs through every method signature — a "C-shaped but T-payloaded" hybrid worth naming as its own seam sub-class in future passes.
6. **[#92/A5, medium]** Transaction-pattern inconsistency: documents does 14 non-test `r.db.BeginTx` calls directly in its repository (§1.13) vs. approval's 20 `TxRunner`/0 `BeginTx` split (§4.11) and controlleddocuments'/templates' near-total `TxRunner` adoption (§2.12/§3.11). documents is the outlier.
7. **[#92/A5, low-medium]** Internal layering deviation inside approval itself: `application/route_admin_service.go` issues raw SQL against approval's *own* tables (`approval_routes`, `approval_route_stages`, `approval_route_stage_selectors`, §4.3) directly from the application layer rather than through `infrastructure/`. This is a smaller version of the same discipline gap that makes the cross-module `documents` writes (finding 1) easy to add without friction — the module doesn't consistently confine SQL to `infrastructure/` even for tables it owns.
8. **[#93/A4, positive counter-evidence — worth preserving as precedent]** Three genuinely healthy patterns exist and should be the templates for remediation, not replaced: (a) documents' `CDFieldReader`-based `LoadDocumentAreaCode` (§1.9, avoids the exact anti-pattern finding 1 exhibits); (b) controlleddocuments' `Dependencies` struct at `module.go:47-90` (§2.9, producer-published ports, fail-loud composition); (c) templates' `approval_completion_writer.go` (§3.9) — the one place approval's decision outcome is written into a *foreign* table (`templates_template_version`) through a writer the foreign module provides, not raw SQL. Finding 1's fix should look like (c), applied to `documents`.
9. **[#89/A8, informational, no defect found]** Capability surface is proportionate and consistent across all four modules (documents 21 / controlleddocuments 8 / templates 6 / approval 32 `authz.Require` sites, §4.15); no role-based reasoning found in a targeted scan. This pass did not find an A8 defect in these four modules — a negative result, not a clean bill of health for the whole system (only these 4 modules were in scope).
10. **[#90/A3, informational]** All four modules are structurally contract-first (each has a generated `api/api.gen.go`+`cfg.yaml`+`gen.go`, §1.15/§2.14/§3.13/§4.13); no hand-rolled routes bypassing oapi-codegen were found. A deeper DTO-drift check (ADR 0035 flat/envelope) was out of scope for this pass.

### Surprises vs. the known facts brief
- The known facts stated "approval reads documents/document_comments/controlled_documents tables directly" — confirmed, but the far more significant fact is that approval also **writes** `documents` (status transitions), from **application-layer** code in 8 files, not just infrastructure-layer reads. The brief undersold the severity.
- "documents reads approval_instances/approval_signoffs/release_generations" — confirmed, and **additionally** `approval_stage_instances`; all of documents' foreign reads are cleanly confined to `infrastructure/`, unlike approval's foreign writes into `application/`. The two reciprocal-pair members are not equally undisciplined.
- Not previously called out: the templates↔approval and documents↔controlleddocuments pairs each contain **one exemplary seam and one poor seam simultaneously** — this is not "module X is bad, module Y is good," it's inconsistent seam discipline *within the same collaborator pair*, which is arguably a harder problem to fix by simple convention than a uniformly bad pair would be.
- taxonomy↔approval reciprocal pair (mentioned in known facts) is real per the edge evidence but was not deep-dived here — templates and controlleddocuments were the two module pairs the task scoped for depth; taxonomy's own document (if the audit does one) should re-derive it directly.
