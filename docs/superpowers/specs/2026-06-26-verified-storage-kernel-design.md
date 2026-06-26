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
- **Q3 = B (write-path only)** — kernel enforces a tenant-scoped key prefix assertion
  on the **write-entry** methods (`PresignPut`, `Confirm`) where the key originates
  server-side and creates new state. The **read/lifecycle** methods
  (`PresignGet`, `Exists`, `Size`, `Delete`) are NOT guarded: their keys are sourced
  from tenant-scoped DB rows or are legitimate non-tenant `system/` keys (see §5).
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

// --- write path: tenant-guarded (server-built key, creates state) ---
func (s *VerifiedStore) PresignPut(ctx context.Context, tenantID, key string, ttl time.Duration) (url string, err error) // signingClient
func (s *VerifiedStore) Confirm(ctx context.Context, tenantID, key, expectedHash string) (VerifiedPointer, error)         // client

// --- read path: NOT guarded (key is DB-sourced / server-trusted; may be system/ scoped) ---
func (s *VerifiedStore) PresignGet(ctx context.Context, key string, ttl time.Duration) (url string, err error) // signingClient
func (s *VerifiedStore) Exists(ctx context.Context, key string) (bool, error)                                  // client
func (s *VerifiedStore) Size(ctx context.Context, key string) (int64, error)                                   // client

// --- lifecycle: NOT guarded (DB-sourced or internal tenant key) ---
func (s *VerifiedStore) Delete(ctx context.Context, key string) error // client
```

**Client selection:** presign methods (`PresignPut`, `PresignGet`) sign with
`signingClient` (the public-facing endpoint baked into the browser-bound URL);
server-to-server IO (`Confirm`, `Exists`, `Size`, `Delete`) uses `client` (the internal
endpoint). This split is load-bearing — see `document_presigner_test.go:13-56`, which
proves `client` = internal `minio:9000` and `signingClient` = public `127.0.0.1:9000`.

**Read-path trust contract:** callers of `PresignGet`/`Exists`/`Size`/`Delete` MUST pass
a key that was either built server-side or retrieved from a tenant-scoped DB query. They
MUST NOT pass raw client input. See §5 for the one current violation and its fix.

`Confirm` semantics (the invariant):
1. Assert tenant key prefix (see §5). On violation → `ErrKeyOutsideTenant`, no IO.
2. GET object. NoSuchKey → `ErrObjectMissing`.
3. Hash via `io.Copy(sha256, io.LimitReader(obj, maxSizeBytes+1))`.
4. Over limit → `Delete` object → `ErrObjectTooLarge`.
5. `hash != expectedHash` → `Delete` object → `ErrHashMismatch`.
6. Else return `VerifiedPointer{key, hash, size}` — `size` is the byte count already
   measured while hashing (no extra round-trip). Object is now the only verified state;
   caller persists the pointer in its own table/transaction. Callers that today make a
   separate `SizeObject` call right after hashing (`documents/.../service.go:587`) drop
   it and use `vp.SizeBytes`.

Removed vs today:
- `AdoptTempObject` — dead (interface + test fakes only; `service.go:253` comment states
  "no AdoptTempObject"). Removed from kernel and from the documents port.
- Raw `HashObject` / `HeadContentHash` — folded into `Confirm`.
- `PresignRevisionPUT`'s key construction — moves into documents' own `keys.go`.

## 5. Tenant guard (write-path only)

The **write-entry** methods (`PresignPut`, `Confirm`) take an explicit `tenantID` and
assert:

```go
prefix := "tenants/" + tenantID + "/"
if !strings.HasPrefix(key, prefix) {
    return ..., ErrKeyOutsideTenant
}
```

This is where new state is created from a server-built key, so it is the correct
chokepoint: a key-builder bug or injected key cannot write/confirm outside its tenant.
Defense-in-depth *under* the existing HTTP capability checks (not a replacement).

**Why the read/lifecycle methods are NOT guarded.** `PresignGet`/`Exists`/`Size`/`Delete`
operate on keys that already exist and were retrieved from a tenant-scoped DB row (the
SQL query is the access control), OR on legitimate non-tenant `system/` keys. Concrete
evidence that a tenant prefix on reads would break production:
- `controlleddocuments/.../service.go:475` calls `Exists` with a published-version docx
  key that is `"system/templates/blank.docx"` for the blank/system template
  (`service_test.go:517`, `:650`) — no `tenants/` prefix, and correctly so.
- `documents/.../view_service.go:80` and `export_service.go:85,103` presign/stat keys
  read raw from DB columns (`final_pdf_s3_key`, export keys). These are already
  tenant-scoped by the row they came from; re-asserting the prefix adds no security and
  forces `tenantID` plumbing into read-only tx callbacks for no benefit.

Guarding reads would convert these into `ErrKeyOutsideTenant` failures. The guard belongs
only where untrusted/server-built keys enter the *write* path.

### 5.1 Templates key migration

`create.go` must emit `tenants/{tenant}/templates/{templateID}/versions/{n}.docx` (and
`.schema.json`) so confirmed template uploads satisfy the write-path guard. No backfill
is required: v1 is not released and there is no production data (clean re-baseline at
release), so existing dev/test objects under the old bare `templates/...` keys are
disposable. Tests that assert the old key format are updated.

### 5.2 Security finding — `RedirectSignedUrl` presigns an unvalidated client key

Surfaced during design review (independent of, but resolved by, this work). The handler
`RedirectSignedUrl` (`templates/.../routes_generated.go:150`) passes a **raw client
`key` query param** to `PresignStoredObject` (`queries.go:60`), which calls `PresignGET`
with zero tenant/ownership validation. Any authenticated user can presign a GET for any
object in the bucket — a cross-tenant read. The safe sibling `GetDocxURL`
(`queries.go:46`) does a proper `GetTemplate` + `GetVersion` tenant lookup and presigns
the *stored* key.

Because the kernel read path is deliberately unguarded (§5), this caller must be fixed at
its boundary. **Decision (to confirm in the plan):** delete `PresignStoredObject` +
`RedirectSignedUrl` as a redundant unsafe shim and route any remaining consumer through
`GetDocxURL`. If a non-docx stored object genuinely needs signing, replace it with a
tenant+ownership-validated lookup that resolves the key from a DB row, never from client
input. This finding is in scope: a cross-tenant presign is a storage-integrity defect.

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
  internal/modules/templates/application/queries.go        (drop PresignStoredObject — §5.2)
  internal/modules/templates/delivery/http/routes_generated.go (drop RedirectSignedUrl — §5.2)
  internal/modules/documents/application/service.go        (narrow Presigner port; Confirm; drop Adopt; use vp.SizeBytes)
  internal/modules/documents/application/view_service.go   (ViewPresigner port → PresignGet)
  internal/modules/documents/application/export_service.go (ExportPresigner port → PresignGet/Exists/Size)
  internal/modules/controlleddocuments/...                 (Exists via kernel port)
  internal/modules/*/module.go                             (wire *VerifiedStore in place of two presigners)
  + test fakes updated to the narrowed ports
```

### 9.1 Three separate consumer ports to migrate

There is not one port but three; the plan must migrate each explicitly. All three are
consumer-defined interfaces that `*VerifiedStore` satisfies:

| Port (file) | Method today | → kernel method | tenantID source |
|---|---|---|---|
| `documents .../service.go:68` `Presigner` | `PresignRevisionPUT(tenant,doc,hash)` | `PresignPut(tenant,key,ttl)` + key built in `keys.go` | cmd |
|  | `HashObject(key)` | `Confirm(tenant,key,expected)` | cmd |
|  | `PresignObjectGET(key)` | `PresignGet(key,ttl)` | — |
|  | `SizeObject(key)` | `Size(key)` (or `vp.SizeBytes` post-Confirm) | — |
|  | `Exists(key)` | `Exists(key)` | — |
|  | `DeleteObject(key)` | `Delete(key)` | — |
|  | `AdoptTempObject(...)` | **removed** (dead) | — |
| `documents .../export_service.go:20` `ExportPresigner` | `PresignObjectGET(key)` | `PresignGet(key,ttl)` | — |
|  | `HeadObject(key)` | `Exists(key)` | — |
|  | `SizeObject(key)` | `Size(key)` | — |
| `documents .../view_service.go:17` `ViewPresigner` | `PresignObjectGET(key)` | `PresignGet(key,ttl)` | — |
| `templates .../ports.go:40` `Presigner` | `PresignPUT(key,ttl)` | `PresignPut(tenant,key,ttl)` | cmd/handler |
|  | `PresignGET(key,ttl)` | `PresignGet(key,ttl)` | — |
|  | `HeadContentHash(key)` | `Confirm(tenant,key,expected)` | cmd |
|  | `Delete(key)` | `Delete(key)` | — |
| `controlleddocuments .../service.go` (docInit) | `Exists(key)` | `Exists(key)` | — |

### 9.2 Test-fake cleanup

Removing `AdoptTempObject` from the documents `Presigner` port makes the fake methods
dead: delete `AdoptTempObject` from `service_test.go:325` and
`service_review_roundtrip_integration_test.go:345`. All fakes are narrowed to only the
methods their module's port still declares.

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

- `verified_store_test.go` (kernel, table-driven): `Confirm` success returns correct
  pointer (incl. `SizeBytes` measured during hashing); mismatch deletes + returns
  `ErrHashMismatch`; missing returns `ErrObjectMissing`; over-limit deletes + returns
  `ErrObjectTooLarge`; **write-path guard** — `PresignPut`/`Confirm` reject an
  out-of-tenant key with `ErrKeyOutsideTenant` and make no IO call; **read-path** —
  `PresignGet`/`Exists`/`Size` succeed for a `system/`-scoped key (no guard).
- Regression test for §5.2: the unvalidated client-key presign path is gone (handler +
  `PresignStoredObject` removed; safe `GetDocxURL` retained).
- Module application tests updated to the narrowed ports and the `Confirm` call shape;
  fakes implement only the methods their module's port declares.
- `go build ./...` and `go test ./...` green before commit.

## 13. Open follow-ons (tracked, NOT in this spec)

- Dead-FSM cleanup: `documents` `CanTransitionDocument` (called 0×) and approval
  `IsLegalTransition` (test-only) — wire or delete.
- Document God-aggregate split: extract editor-sessions and checkpoints out of the
  1,794-line documents repository.
