# Verified-Storage Kernel — Design

**Date:** 2026-06-26
**Status:** Approved (design); implementation plan pending
**Author:** MetalDocs maintainer (Claude)
**Supersedes scope of:** the storage-duplication finding in `2026-06-26-storage-integrity-design.md` (Spec B)

---

## 1. Problem

Object-storage handling is duplicated across modules. Two near-identical presigner
types live side-by-side in `internal/platform/objectstore/`:

- `DocumentPresigner` (`document_presigner.go`, `document_presigner_export.go`) — 7+ methods.
- `TemplatesPresigner` (`templates_presigner.go`) — 4 methods.

The confirm-by-hash logic is copy-paste (`HashObject` vs `HeadContentHash`: GetObject →
Stat → `LimitReader` SHA256 → size check, both returning a per-module domain sentinel).
The verify-and-cleanup dance (hash → compare to expected → delete on mismatch) is
*also* duplicated, in each module's `CommitAutosave` application method.

Consequences:
- The security-critical invariant — *no unverified bytes ever become a readable
  pointer* — is reimplemented per module, so a future module can reintroduce an
  unverified read path.
- Each presigner imports its module's `domain` package (`documents/domain`,
  `templates/domain`), coupling infrastructure to domain.
- Templates storage keys are **not tenant-namespaced** (`templates/{id}/versions/N.docx`,
  built at `create.go:54` and `:145`), unlike documents (`tenants/{t}/documents/...`).
  Cross-tenant isolation at the storage layer rests on convention, not structure.

## 2. Goal

One shared **verified-storage kernel** that both templates and documents (and
controlleddocuments, which consumes `Exists`) call, replacing the duplicated
presigners. The kernel owns the `presign → confirm → pointer` invariant in one place.

Non-goals (explicitly out of scope, YAGNI):
- No shared state-machine engine (audit found two of three FSMs are dead code — a
  separate cleanup).
- No central `verified_objects` DB ledger (no current driver; per-module pending-upload
  expiry already sweeps orphans).
- No object-ACL/token system.
- No Document God-aggregate split (real, tracked as a separate follow-on workstream).
- No shared versioning kernel (audit verdict: templates vs documents versioning are
  genuinely different concepts — keep separate).

## 3. Architecture & boundary

One concrete type `VerifiedStore` in package `internal/platform/objectstore/`, replacing
`DocumentPresigner` and `TemplatesPresigner`.

- **Pure blob mechanism.** Operates on opaque `string` storage keys. Zero domain
  imports. Zero database access.
- **Owns the invariant.** `presign → confirm → pointer`: no bytes become a readable
  pointer until hashed and verified against the caller's expected hash.
- **Consumer-defined ports.** Each module keeps its *own narrow* port interface (Go
  idiom — interfaces defined at the consumer). `*VerifiedStore` satisfies all of them.
  Modules never import each other; each depends only on the kernel via its own port.

Decision record (from brainstorming):
- **Q1 = A** — pure mechanism; key construction stays per-module (not domain-aware
  kernel, not injected KeyPolicy).
- **Q2 = B** — thick `Confirm`: kernel hashes, compares, deletes on mismatch, returns a
  `VerifiedPointer`. Verification invariant lives in one place.
- **Q3 = B** — kernel enforces a tenant-scoped key contract (prefix assertion).
- **Q4 = A** — `Confirm` returns `{key, hash, size}` only; module owns all DB writes.

## 4. Kernel contract

```go
package objectstore

type VerifiedPointer struct {
    StorageKey  string
    ContentHash string // sha256 hex, lowercase
    SizeBytes   int64
}

type VerifiedStore struct {
    client        *minio.Client
    signingClient *minio.Client
    bucket        string
    maxSizeBytes  int64
}

func NewVerifiedStore(client, signingClient *minio.Client, bucket string, maxSizeBytes int64) *VerifiedStore

// --- write path ---
func (s *VerifiedStore) PresignPut(ctx context.Context, tenantID, key string, ttl time.Duration) (url string, err error)
func (s *VerifiedStore) Confirm(ctx context.Context, tenantID, key, expectedHash string) (VerifiedPointer, error)

// --- read path ---
func (s *VerifiedStore) PresignGet(ctx context.Context, tenantID, key string, ttl time.Duration) (url string, err error)
func (s *VerifiedStore) Exists(ctx context.Context, tenantID, key string) (bool, error)
func (s *VerifiedStore) Size(ctx context.Context, tenantID, key string) (int64, error)

// --- lifecycle ---
func (s *VerifiedStore) Delete(ctx context.Context, tenantID, key string) error
```

`Confirm` semantics (the invariant):
1. Assert tenant key prefix (see §5). On violation → `ErrKeyOutsideTenant`, no IO.
2. GET object. NoSuchKey → `ErrObjectMissing`.
3. Hash via `io.Copy(sha256, io.LimitReader(obj, maxSizeBytes+1))`.
4. Over limit → `Delete` object → `ErrObjectTooLarge`.
5. `hash != expectedHash` → `Delete` object → `ErrHashMismatch`.
6. Else return `VerifiedPointer{key, hash, size}`. Object is now the only verified state;
   caller persists the pointer in its own table/transaction.

Removed vs today:
- `AdoptTempObject` — dead (interface + test fakes only; `service.go:253` comment states
  "no AdoptTempObject"). Removed from kernel and from the documents port.
- Raw `HashObject` / `HeadContentHash` — folded into `Confirm`.
- `PresignRevisionPUT`'s key construction — moves into documents' own `keys.go`.

## 5. Tenant guard

Every method takes an explicit `tenantID` and asserts:

```go
prefix := "tenants/" + tenantID + "/"
if !strings.HasPrefix(key, prefix) {
    return ..., ErrKeyOutsideTenant
}
```

This makes cross-tenant object access structurally impossible at the storage boundary,
as defense-in-depth *under* the existing HTTP capability checks (not a replacement).

Consequence — **templates key migration**: `create.go` must emit
`tenants/{tenant}/templates/{templateID}/versions/{n}.docx` (and `.schema.json`).
No backfill is required: v1 is not released and there is no production data
(clean re-baseline at release), so existing dev/test objects under the old bare
`templates/...` keys are disposable. Tests that assert the old key format are updated.

## 6. Error taxonomy

Kernel-owned sentinels (no module-domain imports):

```go
var ErrObjectMissing    = errors.New("objectstore: object not found")
var ErrHashMismatch     = errors.New("objectstore: content hash mismatch")
var ErrObjectTooLarge   = errors.New("objectstore: object exceeds max size")
var ErrKeyOutsideTenant = errors.New("objectstore: key outside tenant scope")
```

Each module maps these to its own domain error at the application boundary (e.g.
templates `ErrObjectMissing → domain.ErrUploadMissing`). The `isNoSuchKeyErr` helper
moves to `errors.go` (shared, already package-local today).

## 7. Key construction (per-module, stays in module)

- **templates** — new `internal/modules/templates/application/keys.go`:
  `tenants/{tenant}/templates/{templateID}/versions/{n}.docx` and the matching
  `.schema.json`.
- **documents** — new `internal/modules/documents/application/keys.go`:
  `tenants/{tenant}/documents/{docID}/revisions/{contentHash}.docx` (moved out of the
  presigner).
- **controlleddocuments** — consumes only `Exists`; no builder.

## 8. Auth / capabilities fit (unchanged)

HTTP-layer capability gates (`CapTemplateEdit`, document edit capabilities) stay exactly
where they are. The kernel is infrastructure and never sees identity or capabilities.
The tenant guard (§5) is a structural backstop beneath those checks.

## 9. File structure

```
internal/platform/objectstore/
  verified_store.go            (NEW — the kernel type + methods)
  verified_store_test.go       (NEW — table tests: success, mismatch→delete,
                                missing, too-large, tenant-guard reject)
  errors.go                    (NEW — sentinels + isNoSuchKeyErr)
  document_presigner.go        (DELETE)
  document_presigner_export.go (DELETE)
  templates_presigner.go       (DELETE)

internal/modules/templates/application/keys.go     (NEW — key builders)
internal/modules/documents/application/keys.go     (NEW — key builders)

Edits:
  internal/modules/templates/application/ports.go        (narrow Presigner port to kernel methods)
  internal/modules/templates/application/create.go        (emit tenant-prefixed keys)
  internal/modules/templates/application/autosave.go      (Confirm replaces HeadContentHash+compare+delete)
  internal/modules/documents/application/service.go        (narrow Presigner port; Confirm; drop Adopt)
  internal/modules/documents/application/view_service.go   (PresignGet via kernel)
  internal/modules/documents/application/export_service.go (PresignGet/Size via kernel)
  internal/modules/controlleddocuments/...                 (Exists via kernel port)
  internal/modules/*/module.go                             (wire *VerifiedStore in place of two presigners)
  + test fakes updated to the narrowed ports
```

## 10. Walkthrough — success (template version upload)

1. HTTP `POST /templates/{id}/versions/{n}/presign` → capability check
   `CapTemplateEdit` passes.
2. Application builds the key via `keys.go`, calls
   `store.PresignPut(tenant, key, 15m)`. Tenant guard OK → presigned PUT URL returned.
3. Client PUTs the docx to MinIO; computes sha256 locally; sends it as
   `expected_content_hash` on the commit request.
4. Application calls `store.Confirm(tenant, key, expected)`. Kernel GETs, hashes,
   `hash == expected` → returns `{key, hash, size}`.
5. Application persists the pointer via `UpdateVersionSchemaCAS` (its own transaction
   and `lock_version` CAS). Done.

Documents follow the same shape via their own `CommitAutosave` and
`document_revisions` write.

## 11. Walkthrough — error paths

- **Hash mismatch:** `Confirm` hashes, `hash != expected` → kernel deletes the object →
  `ErrHashMismatch`. Application maps to HTTP 422, never persists a pointer. No
  unverified bytes survive.
- **Never uploaded:** `Confirm` GET → NoSuchKey → `ErrObjectMissing` → 409/410. Nothing
  persisted.
- **Too large:** `LimitReader` trips at `maxSizeBytes+1` → object deleted →
  `ErrObjectTooLarge` → 413.
- **Wrong-tenant key (builder bug or attack):** guard rejects before any IO →
  `ErrKeyOutsideTenant` → 500 + alert; no MinIO call made.
- **Confirm crashes mid-way:** object is orphaned under its key; the existing per-module
  pending-upload expiry sweep removes it (unchanged). The DB never received a pointer, so
  there is no dangling reference.

## 12. Testing

- `verified_store_test.go` (kernel, table-driven): success returns correct pointer;
  mismatch deletes + returns `ErrHashMismatch`; missing returns `ErrObjectMissing`;
  over-limit deletes + returns `ErrObjectTooLarge`; tenant-guard rejects with no IO;
  `PresignPut`/`PresignGet` produce URLs only for in-tenant keys.
- Module application tests updated to the narrowed ports and the `Confirm` call shape;
  fakes implement only the methods their module's port declares.
- `go build ./...` and `go test ./...` green before commit.

## 13. Open follow-ons (tracked, NOT in this spec)

- Dead-FSM cleanup: `documents` `CanTransitionDocument` (called 0×) and approval
  `IsLegalTransition` (test-only) — wire or delete.
- Document God-aggregate split: extract editor-sessions and checkpoints out of the
  1,794-line documents repository.
