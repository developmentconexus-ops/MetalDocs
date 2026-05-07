# Controlled Document Atomic Create + Per-Area Numbering + Idempotency

> **Date:** 2026-05-07
> **Status:** Approved (brainstorm). Awaiting implementation plan.
> **Goal:** One transactional endpoint creates Controlled Document + first document version. Numbering becomes `{PROFILE}-{AREA}-{NNN}` per existing wiki spec. `Idempotency-Key` header replays prior result on retry. Two-screen legacy creation flow deleted.

---

## Context and motivation

The `wiki/concepts/controlled-documents.md:24-37` spec already defines:

- Format `{profile}-{area}-{NNN}` (3-segment, 3-digit zero-padded).
- One sequence counter per `(profile, area)` pair.
- Monotonic, never reused.

Backend reality deviates:

- `internal/modules/registry/domain/controlled_document.go:48` — `AutoCode(profile, seq) → "{PROFILE}-{NN}"` (2-segment, 2-digit pad).
- `migrations/0124_registry_controlled_documents.sql:31-38` — `profile_sequence_counters` PK is `(tenant_id, profile_code)`. Area absent.

The wizard (`NewDocumentWizardPage.tsx:151`) already issues two HTTP calls: `POST /api/v2/controlled-documents` then `POST /api/v2/documents`. If the second call fails after the first commits, the slot is orphaned and its sequence number is permanently consumed (`wiki/backlog/novo-documento.md` `slot-rollback` item). For an ISO/QMS product this is a numbering-integrity defect: auditors flag missing sequence numbers.

This work brings the backend in line with the documented spec and closes the integrity gap by collapsing creation into a single transaction.

---

## Decisions made during brainstorm

| # | Decision | Rationale |
|---|---|---|
| 1 | Single PR delivers numbering + atomic create together (option B). | Numbering looks better only matters if integrity is also fixed; two PRs always slip. |
| 2 | Endpoint owned by registry module at `POST /api/v2/controlled-documents`. | CD is the primary resource. Registry depends on documents module via thin `DocumentInitializer` port. |
| 3 | Cleaner REST: revisions move to `POST /api/v2/controlled-documents/{id}/revisions`. `POST /api/v2/documents` deleted. | CD-rooted resource model. Single creation path. |
| 4 | `Idempotency-Key` header required on both writes. | Existing infra (`metaldocs.idempotency_keys` table + janitor job) already supports route-template-keyed replay. Senior move = consume it, not duplicate. |
| 5 | Refactor `PostgresSignoffIdempStore` into generic `internal/platform/idempotency/postgres_store.go`. | Avoid two parallel stores. Future opt-in routes get the primitive for free. |
| 6 | Three-digit pad per wiki spec. | Smoke fixtures used 2-digit; superseded. |
| 7 | Delete legacy two-screen flow: `RegistryCreateDialog`, `DocumentCreatePage` and their routes. | Wizard is sole creation path. Manual-code admin override stays as a body field on the atomic endpoint. |
| 8 | Drop `controlled_documents` smoke fixtures + `profile_sequence_counters` in migration. | Dev-only DB. No prod tenants with allocated codes (confirm before merge). |

---

## Architecture

```
Wizard (frontend)
  └─ POST /api/v2/controlled-documents       [atomic, idempotent]
       │
  Registry HTTP handler
       │   idempotency middleware
       │   (CheckReplay → hit: replay stored response)
       │
  RegistryService.Create(cmd)                [single sql.Tx]
       │
       ├─ profile + area validation
       ├─ SequenceAllocator.NextAndIncrement(tx, tenant, profile, area) → next
       ├─ AutoCode(profile, area, next) → "DC-RH-001"
       ├─ ControlledDocumentRepo.CreateTx(tx, cd)
       ├─ DocumentInitializer.CloneTemplate(tx, cd, templateVersionID, name)
       │     (interface satisfied by documents module adapter)
       └─ commit
```

The transaction guarantees that either both rows exist or neither. The sequence number is not consumed on failure because the counter `UPDATE` lives in the same transaction as the CD `INSERT`.

---

## Endpoints

| Method | Path | Purpose |
|---|---|---|
| **POST** | `/api/v2/controlled-documents` | **Atomic create.** Required header `Idempotency-Key` (UUID v4). Body: `{profileCode, processAreaCode, title, ownerUserId, templateVersionId?, manualCode?, manualCodeReason?, documentName}`. Response 201 `{controlledDocument, document}`. |
| **POST** | `/api/v2/controlled-documents/{id}/revisions` | Create new revision on existing CD. Required header `Idempotency-Key`. Body: `{name, formData?, templateVersionId?}`. Response 201 `{document}`. |
| **GET** | `/api/v2/controlled-documents/preview-code?profileCode=&areaCode=` | Race-prone informational preview. Response 200 `{profileCode, areaCode, nextSeq, code}`. Read-only — does not reserve. |
| ~~POST~~ | ~~`/api/v2/documents`~~ | **Deleted.** |

### Error codes

- `400 IDEMPOTENCY_KEY_REQUIRED` — missing header on atomic / revision routes.
- `400 IDEMPOTENCY_KEY_INVALID` — header is not a UUID.
- `422 IDEMPOTENCY_KEY_CONFLICT` — same key + different payload hash.
- `409 CD_CODE_TAKEN` — manual-code collision.
- `409 SEQUENCE_COUNTER_NOT_FOUND` — internal allocator failure (existing).
- `400 VALIDATION_ERROR` — body field validation.

---

## Schema changes

### Migration `0182_cd_sequence_per_area.sql`

```sql
-- WARNING: dev-only DB. Confirms no prod tenants have allocated codes before merge.
DROP TABLE IF EXISTS profile_sequence_counters;
DELETE FROM controlled_documents;

CREATE TABLE cd_sequence_counters (
  tenant_id          UUID NOT NULL,
  profile_code       TEXT NOT NULL,
  process_area_code  TEXT NOT NULL,
  next_seq           INT  NOT NULL DEFAULT 1,
  PRIMARY KEY (tenant_id, profile_code, process_area_code),
  FOREIGN KEY (tenant_id, profile_code)
    REFERENCES metaldocs.document_profiles (tenant_id, code),
  FOREIGN KEY (tenant_id, process_area_code)
    REFERENCES metaldocs.document_process_areas (tenant_id, code)
);

GRANT SELECT, INSERT, UPDATE ON cd_sequence_counters TO metaldocs_app;
```

`metaldocs.idempotency_keys` already exists (created by the migration that introduced the signoff idempotency store) and requires no schema change.

---

## Domain changes

### `internal/modules/registry/domain/controlled_document.go:48`

```go
func AutoCode(profileCode, areaCode string, seq int) string {
    return fmt.Sprintf("%s-%s-%03d",
        strings.ToUpper(profileCode),
        strings.ToUpper(areaCode),
        seq)
}
```

### `internal/modules/registry/domain/sequence.go`

```go
type SequenceAllocator interface {
    NextAndIncrement(ctx context.Context, tx DBExecutor, tenantID, profileCode, areaCode string) (int, error)
    Peek(ctx context.Context, tenantID, profileCode, areaCode string) (int, error)
    EnsureCounter(ctx context.Context, tenantID, profileCode, areaCode string) error
}
```

`Peek` is read-only; defaults to `1` when no counter row exists. Used by the preview endpoint.

### New port `internal/modules/registry/domain/document_initializer.go`

```go
type DocumentInitializer interface {
    CloneTemplate(ctx context.Context, tx *sql.Tx, cd *ControlledDocument, req CloneTemplateRequest) (*DocumentRef, error)
}

type CloneTemplateRequest struct {
    TemplateVersionID *string  // nil → use profile default
    Name              string
    FormData          map[string]any
}

type DocumentRef struct {
    ID          string
    ContentHash string
}
```

### `internal/modules/registry/application/service.go`

`CreateControlledDocumentCmd` adds:

```go
TemplateVersionID *string
DocumentName      string
FormData          map[string]any
```

`Create` body changes:

1. Allocates sequence inside the existing tx, passing `cmd.ProcessAreaCode`.
2. After CD insert, calls `s.docInit.CloneTemplate(ctx, tx, doc, ...)` and includes the resulting `DocumentRef` in the response.
3. Commits once. No partial state on failure.

New service method:

```go
func (s *RegistryService) PreviewCode(ctx context.Context, tenantID, profileCode, areaCode string) (string, error)
```

Calls `Peek` and `AutoCode`. Returns the formatted code without mutating state.

---

## Cross-module wiring

`internal/modules/documents/application/cd_initializer.go` (new):

```go
type CDDocumentInitializer struct{ svc *DocumentService }

func (i *CDDocumentInitializer) CloneTemplate(ctx context.Context, tx *sql.Tx, cd *registrydomain.ControlledDocument, req registrydomain.CloneTemplateRequest) (*registrydomain.DocumentRef, error) { ... }
```

`internal/modules/registry/module.go`:

```go
type Dependencies struct {
    DB                  *sql.DB
    Profiles            ProfileReader
    Areas               AreaReader
    TemplateChecker     TemplateVersionChecker
    DocumentInitializer registrydomain.DocumentInitializer
    GovernanceLogger    taxonomydomain.GovernanceLogger
}
```

Composition root (`cmd/api/main.go` or equivalent) wires `documents.NewCDDocumentInitializer(documentsSvc)` into `registry.Module`.

---

## Idempotency

### Generic store `internal/platform/idempotency/postgres_store.go`

```go
type Store struct {
    db            *sql.DB
    routeTemplate string
}

type Replay struct {
    Status int
    Body   []byte
}

var ErrIdempotencyConflict = errors.New("idempotency key reused with different payload")

func New(db *sql.DB, routeTemplate string) *Store

func (s *Store) CheckReplay(ctx context.Context, tenantID, actorID, key, payloadHash string) (*Replay, error)
// returns nil, nil   on miss
// returns *Replay, nil  on hit with matching hash
// returns nil, ErrIdempotencyConflict on hit with different hash

func (s *Store) RecordReplay(ctx context.Context, tenantID, actorID, key, payloadHash string, status int, body []byte) error
```

`PostgresSignoffIdempStore` is rewritten as a thin wrapper around `Store` constructed with `routeTemplate = "POST /api/v2/documents/{id}/signoff"`. No behavior change for the signoff route.

### Middleware `internal/platform/idempotency/middleware.go`

```go
func Require(store *Store, actorFromCtx func(context.Context) string) func(http.Handler) http.Handler
```

Middleware contract:

1. Read `Idempotency-Key` header. Empty → `400 IDEMPOTENCY_KEY_REQUIRED`. Non-UUID → `400 IDEMPOTENCY_KEY_INVALID`.
2. Read full request body, compute `sha256` hex hash, restore body for downstream handler.
3. `CheckReplay` → on hit: write stored status + body, return.
4. On `ErrIdempotencyConflict`: `422 IDEMPOTENCY_KEY_CONFLICT`.
5. On miss: invoke next handler with a `httptest.ResponseRecorder`-like recorder. After completion, if status is 2xx, call `RecordReplay`. Non-2xx is not recorded — next attempt may allocate fresh.

Applied to:
- `POST /api/v2/controlled-documents` (atomic create).
- `POST /api/v2/controlled-documents/{id}/revisions`.

Not applied retroactively to other routes in this PR. System-wide adoption is a separate ADR.

---

## Frontend changes

| File | Change |
|---|---|
| `lib/api/client.ts` | `apiFetch` accepts `{ idempotencyKey?: string }`; adds `Idempotency-Key` header when present. |
| `lib/queryKeys.ts` | New entry `QK.controlledDocuments.preview(profile, area)`. |
| `features/registry/api/controlledDocuments.ts` | New `createControlledDocumentAtomic(req, key)` → atomic POST. New `createRevision(cdID, req, key)` → revision POST. New `previewCode(profile, area)` → preview GET. **Delete** old `createControlledDocument`. |
| `features/documents/api/documentsV2.ts` | **Delete** `createDocument`. |
| `features/registry/queries/usePreviewCodeQuery.ts` | New TanStack query, debounced 250ms, `enabled: profile && area`. |
| `features/documents/components/wizard/CodePreviewBanner.tsx` | Replace `???` with query data. Skeleton while loading. Disabled empty state preserved when profile/area unset. |
| `features/documents/pages/NewDocumentWizardPage.tsx` | `handleCreate` collapses to single call. Generates `crypto.randomUUID()` once at submit start, stable across in-flight retries. New UUID on user-initiated retry after error. Delete slot-rollback TODO. |
| `features/registry/RegistryDetailPage.tsx` | `handleCreateNewRevision` uses `createRevision` with new UUID. |
| `features/registry/RegistryCreateDialog.tsx` + `.test.tsx` | **Delete**. Route removed. |
| `features/documents/pages/DocumentCreatePage.tsx` | **Delete**. Route removed. |

---

## Tests

### Backend

| File | Test |
|---|---|
| `internal/modules/registry/domain/sequence_test.go` | Concurrent `(profile, area)` independence: `DC-RH-001..N` and `DC-FIN-001..M` advance independently. |
| `internal/modules/registry/application/service_test.go` | Atomic create happy path with fake `DocumentInitializer`. Manual-code path. Initializer error rolls back CD insert and does not consume sequence. |
| `internal/modules/registry/delivery/http/routes_contract_test.go` | Atomic POST 201, revision POST 201, preview GET 200, idempotency 400/422. |
| `internal/platform/idempotency/postgres_store_test.go` | Replay hit, replay conflict, expiry, missing row. |
| `internal/modules/registry/integration_test.go` (new or existing) | End-to-end against real Postgres: atomic create persists both CD + document, sequence counter increments by 1, second create with same key returns prior response. |

### Frontend

| File | Test |
|---|---|
| `features/documents/components/wizard/CodePreviewBanner.test.tsx` | Loading, loaded, disabled empty state. |
| `features/documents/pages/NewDocumentWizardPage.test.tsx` | Single-call submit; idempotency UUID stable across in-flight retry; new UUID after error retry. |
| Delete | `features/registry/RegistryCreateDialog.test.tsx`. |

---

## Wiki updates (post-merge, dispatch wiki-curator)

- `wiki/concepts/controlled-documents.md` — bump verified, drop stub warnings, document atomic create, preview endpoint, idempotency contract.
- `wiki/backlog/novo-documento.md` — close `sequence-preview` and `slot-rollback` deferrals.
- `wiki/modules/documents.md` — describe atomic-create endpoint shape, idempotency requirement, deletion of legacy `POST /api/v2/documents`.
- `wiki/modules/taxonomy.md` — re-verify (no API surface change here).
- New `wiki/decisions/0009-cd-atomic-create.md` — ADR capturing spec drift, atomic decision, registry-owned ownership, idempotency-key adoption strategy.

---

## Out of scope

- Visibility column on `controlled_documents` (`wiki/backlog/novo-documento.md` `visibility`).
- Profile counts (`wiki/backlog/novo-documento.md` `profile-counts`).
- System-wide idempotency adoption beyond these two routes.
- Audit log entry for auto-allocated codes (manual-code path already governance-logged; auto path can be added later).
- Migration of any existing prod-allocated `DC-NN` codes — assumed dev-only DB.

---

## Open risks and mitigations

| Risk | Mitigation |
|---|---|
| Backend ships before frontend; wizard hits new endpoint without UUID. | Ship backend + frontend in same merge. Atomic endpoint rejects with `400 IDEMPOTENCY_KEY_REQUIRED` is acceptable until frontend updated. |
| `DocumentInitializer.CloneTemplate` requires the same `*sql.Tx` instance. Documents module currently exposes service methods that open their own tx. | Adapter accepts the tx and uses raw repo methods that take `tx DBExecutor`; bypass the service-level tx wrapper for this path. |
| Migration drops smoke fixtures — irreversible if prod data exists. | Confirmed dev-only at decision time; document the assumption in the migration header and ADR. |
| `idempotency_keys.payload_hash` differs across retries that legitimately edit the title. | Documented as expected: a body change is a new intent and gets a new code. Wizard regenerates the UUID after any user-visible error. |

---

## Acceptance criteria

1. Wizard creates a CD + document with a single HTTP call.
2. The created code matches `{PROFILE}-{AREA}-{NNN}` and the next CD in the same `(profile, area)` pair gets `NNN+1`. CDs in other areas are independent.
3. Submitting the same `Idempotency-Key` returns the original response without creating a second CD or consuming a second sequence number.
4. Submitting the same key with a different body returns `422 IDEMPOTENCY_KEY_CONFLICT`.
5. Preview banner shows the actual next code (`DC-RH-003`) instead of `???` once profile and area are selected.
6. `RegistryCreateDialog`, `DocumentCreatePage`, and `POST /api/v2/documents` no longer exist.
7. Signoff idempotency continues to work after the store refactor.
8. All listed tests pass.
