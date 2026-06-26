# Verified-Storage Kernel Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the two duplicated object-storage presigners with one shared `VerifiedStore` kernel that owns the presign→confirm→pointer invariant, and close the cross-tenant presign hole.

**Architecture:** One concrete `*VerifiedStore` in `internal/platform/objectstore/` operating on opaque tenant-scoped string keys, satisfying each module's own narrow consumer-defined port. Write-entry methods (`PresignPut`, `Confirm`) enforce a `tenants/{tenantID}/` key prefix; read/lifecycle methods trust DB-sourced keys. `Confirm` hashes, compares to the caller's expected hash, deletes on mismatch, and returns `VerifiedPointer{key,hash,size}`.

**Tech Stack:** Go, minio-go/v7, database/sql, sqlmock (tests), oapi-codegen + openapi-typescript (contract regen).

**Spec:** `docs/superpowers/specs/2026-06-26-verified-storage-kernel-design.md`

**Global rules for every task:**
- Run from repo root `C:\Users\leandro.theodoro\Documents\MetalDocs` unless stated.
- Use PowerShell (`.\scripts\...`) — never bash for startup. Go commands run directly.
- After each task: `go build ./...` must be green before commit.
- Commit after each task (standing authorization; never push).
- Do NOT touch `.claude/worktrees`.

---

## File Structure

```
internal/platform/objectstore/
  errors.go               (NEW — sentinels; later also isNoSuchKeyErr)
  verified_store.go       (NEW — the kernel)
  verified_store_test.go  (NEW — presign-host + tenant-guard tests)
  document_presigner.go        (DELETE in Task 9)
  document_presigner_export.go (DELETE in Task 9)
  templates_presigner.go       (DELETE in Task 9)
  document_presigner_test.go   (DELETE in Task 9)
  templates_presigner_test.go  (DELETE in Task 9)

internal/modules/templates/application/keys.go   (NEW)
internal/modules/documents/application/keys.go   (NEW)

Modified:
  templates/application/{ports.go, autosave.go, create.go, queries.go, fakes_test.go, *_test.go}
  templates/delivery/http/routes_generated.go        (drop RedirectSignedUrl)
  documents/application/{service.go, view_service.go, export_service.go, service_test.go, service_review_roundtrip_integration_test.go, export_service_test.go}
  apps/api/cmd/metaldocs-api/main.go                  (wire *VerifiedStore)
  api/openapi/v1/openapi.yaml + regenerated artifacts  (drop redirectSignedUrl op)
```

---

## Task 1: Kernel error sentinels

**Files:**
- Create: `internal/platform/objectstore/errors.go`
- Create: `internal/platform/objectstore/errors_test.go`

- [ ] **Step 1: Write the failing test**

`internal/platform/objectstore/errors_test.go`:
```go
package objectstore

import (
	"errors"
	"testing"
)

func TestKernelSentinelsAreDistinct(t *testing.T) {
	all := []error{ErrObjectMissing, ErrHashMismatch, ErrObjectTooLarge, ErrKeyOutsideTenant}
	for i := range all {
		for j := range all {
			if i != j && errors.Is(all[i], all[j]) {
				t.Fatalf("sentinel %d and %d are not distinct", i, j)
			}
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/platform/objectstore/ -run TestKernelSentinelsAreDistinct`
Expected: FAIL — `undefined: ErrObjectMissing`.

- [ ] **Step 3: Write minimal implementation**

`internal/platform/objectstore/errors.go`:
```go
package objectstore

import "errors"

// Kernel-owned sentinels. No module-domain imports; each module maps these to its
// own domain error at the application boundary.
var (
	ErrObjectMissing    = errors.New("objectstore: object not found")
	ErrHashMismatch     = errors.New("objectstore: content hash mismatch")
	ErrObjectTooLarge   = errors.New("objectstore: object exceeds max size")
	ErrKeyOutsideTenant = errors.New("objectstore: key outside tenant scope")
)
```

> NOTE: `isNoSuchKeyErr` is NOT moved here yet — it still lives in `document_presigner.go` and is package-visible. It relocates here in Task 9 when that file is deleted.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/platform/objectstore/ -run TestKernelSentinelsAreDistinct`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/platform/objectstore/errors.go internal/platform/objectstore/errors_test.go
git commit -m "feat(objectstore): verified-store error sentinels"
```

---

## Task 2: VerifiedStore kernel + tests

**Files:**
- Create: `internal/platform/objectstore/verified_store.go`
- Create: `internal/platform/objectstore/verified_store_test.go`

- [ ] **Step 1: Write the failing test**

`internal/platform/objectstore/verified_store_test.go`:
```go
package objectstore

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func newTestStore(t *testing.T) *VerifiedStore {
	t.Helper()
	internalClient, err := minio.New("minio:9000", &minio.Options{
		Creds: credentials.NewStaticV4("minioadmin", "minioadmin", ""), Region: "us-east-1",
	})
	if err != nil {
		t.Fatalf("internal client: %v", err)
	}
	publicClient, err := minio.New("127.0.0.1:9000", &minio.Options{
		Creds: credentials.NewStaticV4("minioadmin", "minioadmin", ""), Region: "us-east-1",
	})
	if err != nil {
		t.Fatalf("public client: %v", err)
	}
	return NewVerifiedStore(internalClient, publicClient, "metaldocs-attachments", 25*1024*1024)
}

func TestVerifiedStore_PresignUsesPublicHost(t *testing.T) {
	s := newTestStore(t)
	key := "tenants/t1/templates/x/versions/1.docx"

	putURL, err := s.PresignPut(context.Background(), "t1", key, 15*time.Minute)
	if err != nil {
		t.Fatalf("PresignPut: %v", err)
	}
	if h := mustHost(t, putURL); h != "127.0.0.1:9000" {
		t.Fatalf("put host = %q", h)
	}

	getURL, err := s.PresignGet(context.Background(), key, 15*time.Minute)
	if err != nil {
		t.Fatalf("PresignGet: %v", err)
	}
	if h := mustHost(t, getURL); h != "127.0.0.1:9000" {
		t.Fatalf("get host = %q", h)
	}
}

func TestVerifiedStore_WritePathGuardRejectsForeignTenant(t *testing.T) {
	s := newTestStore(t)
	foreign := "tenants/other/documents/d/revisions/h.docx"

	if _, err := s.PresignPut(context.Background(), "t1", foreign, time.Minute); err != ErrKeyOutsideTenant {
		t.Fatalf("PresignPut err = %v, want ErrKeyOutsideTenant", err)
	}
	if _, err := s.Confirm(context.Background(), "t1", foreign, "deadbeef"); err != ErrKeyOutsideTenant {
		t.Fatalf("Confirm err = %v, want ErrKeyOutsideTenant", err)
	}
}

func TestVerifiedStore_ReadPathAllowsSystemKey(t *testing.T) {
	s := newTestStore(t)
	// system/ keys have no tenant prefix and must still presign (no guard on reads).
	if _, err := s.PresignGet(context.Background(), "system/templates/blank.docx", time.Minute); err != nil {
		t.Fatalf("PresignGet system key: %v", err)
	}
}

func mustHost(t *testing.T, raw string) string {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	return u.Host
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/platform/objectstore/ -run TestVerifiedStore`
Expected: FAIL — `undefined: VerifiedStore` / `undefined: NewVerifiedStore`.

- [ ] **Step 3: Write minimal implementation**

`internal/platform/objectstore/verified_store.go`:
```go
package objectstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

const defaultMaxObjectBytes = 25 * 1024 * 1024

// VerifiedPointer is the verified result of a confirmed upload.
type VerifiedPointer struct {
	StorageKey  string
	ContentHash string // sha256 hex, lowercase
	SizeBytes   int64
}

// VerifiedStore is the shared object-storage kernel. It operates on opaque string
// keys, holds no domain knowledge, and never touches the database. It owns the
// presign -> confirm -> pointer invariant: no bytes become a readable pointer until
// hashed and verified against the caller's expected hash.
type VerifiedStore struct {
	client        *minio.Client // internal endpoint: server-to-server IO
	signingClient *minio.Client // public endpoint: browser-bound presigned URLs
	bucket        string
	maxSizeBytes  int64
}

func NewVerifiedStore(client, signingClient *minio.Client, bucket string, maxSizeBytes int64) *VerifiedStore {
	if signingClient == nil {
		signingClient = client
	}
	if maxSizeBytes <= 0 {
		maxSizeBytes = defaultMaxObjectBytes
	}
	return &VerifiedStore{client: client, signingClient: signingClient, bucket: bucket, maxSizeBytes: maxSizeBytes}
}

func (s *VerifiedStore) assertTenant(tenantID, key string) error {
	if !strings.HasPrefix(key, "tenants/"+tenantID+"/") {
		return ErrKeyOutsideTenant
	}
	return nil
}

// --- write path (tenant-guarded) ---

func (s *VerifiedStore) PresignPut(ctx context.Context, tenantID, key string, ttl time.Duration) (string, error) {
	if err := s.assertTenant(tenantID, key); err != nil {
		return "", err
	}
	u, err := s.signingClient.PresignedPutObject(ctx, s.bucket, key, ttl)
	if err != nil {
		return "", fmt.Errorf("objectstore: presign put: %w", err)
	}
	return u.String(), nil
}

// Confirm GETs the object, hashes it, compares to expectedHash, deletes on mismatch
// or over-size, and returns the verified pointer on success.
func (s *VerifiedStore) Confirm(ctx context.Context, tenantID, key, expectedHash string) (VerifiedPointer, error) {
	if err := s.assertTenant(tenantID, key); err != nil {
		return VerifiedPointer{}, err
	}

	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		if isNoSuchKeyErr(err) {
			return VerifiedPointer{}, ErrObjectMissing
		}
		return VerifiedPointer{}, fmt.Errorf("objectstore: confirm get: %w", err)
	}
	defer obj.Close()

	if _, err := obj.Stat(); err != nil {
		if isNoSuchKeyErr(err) {
			return VerifiedPointer{}, ErrObjectMissing
		}
		return VerifiedPointer{}, fmt.Errorf("objectstore: confirm stat: %w", err)
	}

	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(obj, s.maxSizeBytes+1))
	if err != nil {
		if isNoSuchKeyErr(err) {
			return VerifiedPointer{}, ErrObjectMissing
		}
		return VerifiedPointer{}, fmt.Errorf("objectstore: confirm hash: %w", err)
	}
	if n > s.maxSizeBytes {
		s.deleteQuiet(ctx, key)
		return VerifiedPointer{}, ErrObjectTooLarge
	}

	actual := hex.EncodeToString(h.Sum(nil))
	if actual != expectedHash {
		s.deleteQuiet(ctx, key)
		return VerifiedPointer{}, ErrHashMismatch
	}
	return VerifiedPointer{StorageKey: key, ContentHash: actual, SizeBytes: n}, nil
}

// --- read path (NOT guarded: keys are DB-sourced / server-trusted) ---

func (s *VerifiedStore) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	u, err := s.signingClient.PresignedGetObject(ctx, s.bucket, key, ttl, nil)
	if err != nil {
		return "", fmt.Errorf("objectstore: presign get: %w", err)
	}
	return u.String(), nil
}

func (s *VerifiedStore) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	if err == nil {
		return true, nil
	}
	if isNoSuchKeyErr(err) {
		return false, nil
	}
	return false, fmt.Errorf("objectstore: exists: %w", err)
}

func (s *VerifiedStore) Size(ctx context.Context, key string) (int64, error) {
	info, err := s.client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		if isNoSuchKeyErr(err) {
			return 0, ErrObjectMissing
		}
		return 0, fmt.Errorf("objectstore: size: %w", err)
	}
	return info.Size, nil
}

// --- lifecycle (NOT guarded) ---

func (s *VerifiedStore) Delete(ctx context.Context, key string) error {
	err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
	if err == nil || isNoSuchKeyErr(err) {
		return nil
	}
	return fmt.Errorf("objectstore: delete: %w", err)
}

func (s *VerifiedStore) deleteQuiet(ctx context.Context, key string) {
	if err := s.Delete(ctx, key); err != nil {
		slog.WarnContext(ctx, "objectstore: cleanup delete failed", "key", key, "err", err)
	}
}

var _ = errors.Is // keep errors import if future edits drop direct use
```

> NOTE: `isNoSuchKeyErr` is referenced from `document_presigner.go` (same package) and resolves at compile time. Remove the `var _ = errors.Is` line if `go vet` flags it; it is only a guard against an unused-import churn during iteration.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/platform/objectstore/ -run TestVerifiedStore`
Expected: PASS (presign + guard tests are offline; no MinIO needed).

- [ ] **Step 5: Commit**

```bash
git add internal/platform/objectstore/verified_store.go internal/platform/objectstore/verified_store_test.go
git commit -m "feat(objectstore): VerifiedStore kernel with write-path tenant guard"
```

---

## Task 3: Templates key builder + tenant-prefixed creation

**Files:**
- Create: `internal/modules/templates/application/keys.go`
- Create: `internal/modules/templates/application/keys_test.go`
- Modify: `internal/modules/templates/application/create.go:54`, `:145`

- [ ] **Step 1: Write the failing test**

`internal/modules/templates/application/keys_test.go`:
```go
package application

import "testing"

func TestTemplateDocxKey_IsTenantScoped(t *testing.T) {
	got := templateDocxKey("tenant-1", "tpl-9", 3)
	want := "tenants/tenant-1/templates/tpl-9/versions/3.docx"
	if got != want {
		t.Fatalf("templateDocxKey = %q, want %q", got, want)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/modules/templates/application/ -run TestTemplateDocxKey_IsTenantScoped`
Expected: FAIL — `undefined: templateDocxKey`.

- [ ] **Step 3: Write minimal implementation**

`internal/modules/templates/application/keys.go`:
```go
package application

import "fmt"

// templateDocxKey builds the canonical tenant-scoped object key for a template
// version's docx. The tenants/{tenantID}/ prefix satisfies the VerifiedStore
// write-path guard.
func templateDocxKey(tenantID, templateID string, versionNumber int) string {
	return fmt.Sprintf("tenants/%s/templates/%s/versions/%d.docx", tenantID, templateID, versionNumber)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/modules/templates/application/ -run TestTemplateDocxKey_IsTenantScoped`
Expected: PASS.

- [ ] **Step 5: Switch create.go to the builder**

In `internal/modules/templates/application/create.go`, replace line 54:
```go
		fmt.Sprintf("templates/%s/versions/1.docx", template.ID),
```
with:
```go
		templateDocxKey(cmd.TenantID, template.ID, 1),
```

And replace line 145:
```go
		fmt.Sprintf("templates/%s/versions/%d.docx", cmd.TemplateID, newNum),
```
with:
```go
		templateDocxKey(cmd.TenantID, cmd.TemplateID, newNum),
```

If `fmt` becomes unused in `create.go` after this, remove it from the import block.

- [ ] **Step 6: Run build + template app tests**

Run: `go build ./... && go test ./internal/modules/templates/...`
Expected: build PASS. Some `create`/`lifecycle` tests may now assert the new key format — fix any failing assertion by changing expected `templates/<id>/versions/N.docx` to `tenants/<tenant>/templates/<id>/versions/N.docx` in those tests only.

- [ ] **Step 7: Commit**

```bash
git add internal/modules/templates/application/keys.go internal/modules/templates/application/keys_test.go internal/modules/templates/application/create.go internal/modules/templates/application/*_test.go
git commit -m "feat(templates): tenant-scoped docx storage keys"
```

---

## Task 4: Documents key builder

**Files:**
- Create: `internal/modules/documents/application/keys.go`
- Create: `internal/modules/documents/application/keys_test.go`

- [ ] **Step 1: Write the failing test**

`internal/modules/documents/application/keys_test.go`:
```go
package application

import "testing"

func TestDocumentRevisionKey_IsTenantScoped(t *testing.T) {
	got := documentRevisionKey("tenant-1", "doc-7", "abc123")
	want := "tenants/tenant-1/documents/doc-7/revisions/abc123.docx"
	if got != want {
		t.Fatalf("documentRevisionKey = %q, want %q", got, want)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/modules/documents/application/ -run TestDocumentRevisionKey_IsTenantScoped`
Expected: FAIL — `undefined: documentRevisionKey`.

- [ ] **Step 3: Write minimal implementation**

`internal/modules/documents/application/keys.go`:
```go
package application

import "fmt"

// documentRevisionKey builds the canonical tenant-scoped object key for a document
// revision's docx. Mirrors the layout the old DocumentPresigner.PresignRevisionPUT
// constructed internally.
func documentRevisionKey(tenantID, documentID, contentHash string) string {
	return fmt.Sprintf("tenants/%s/documents/%s/revisions/%s.docx", tenantID, documentID, contentHash)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/modules/documents/application/ -run TestDocumentRevisionKey_IsTenantScoped`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/modules/documents/application/keys.go internal/modules/documents/application/keys_test.go
git commit -m "feat(documents): tenant-scoped revision storage key builder"
```

---

## Task 5: Migrate templates Presigner port to the kernel

**Files:**
- Modify: `internal/modules/templates/application/ports.go:40-45`
- Modify: `internal/modules/templates/application/autosave.go`
- Modify: `internal/modules/templates/application/fakes_test.go`
- Modify: `apps/api/cmd/metaldocs-api/main.go:749`

- [ ] **Step 1: Narrow the port to kernel shape**

Replace `ports.go:40-45` (the `Presigner` interface) with:
```go
type Presigner interface {
	PresignPut(ctx context.Context, tenantID, key string, ttl time.Duration) (url string, err error)
	PresignGet(ctx context.Context, key string, ttl time.Duration) (url string, err error)
	Confirm(ctx context.Context, tenantID, key, expectedHash string) (objectstore.VerifiedPointer, error)
	Delete(ctx context.Context, key string) error
}
```
Add the import `"metaldocs/internal/platform/objectstore"` to `ports.go`.

- [ ] **Step 2: Update autosave.go presign + commit call sites**

In `autosave.go`:

Replace the `PresignTemplateUpload` presign call (line 46):
```go
	url, err := s.presign.PresignPUT(ctx, key, autosaveUploadTTL)
```
with:
```go
	url, err := s.presign.PresignPut(ctx, cmd.TenantID, key, autosaveUploadTTL)
```

Replace the `PresignAutosave` presign call (line 70):
```go
	url, err := s.presign.PresignPUT(ctx, version.DocxStorageKey, autosaveUploadTTL)
```
with:
```go
	url, err := s.presign.PresignPut(ctx, cmd.TenantID, version.DocxStorageKey, autosaveUploadTTL)
```

Replace the `CommitAutosave` hash+compare+delete block (lines 101-113):
```go
	actualHash, err := s.presign.HeadContentHash(ctx, version.DocxStorageKey)
	if err != nil {
		if errors.Is(err, domain.ErrUploadMissing) {
			return nil, domain.ErrUploadMissing
		}
		return nil, fmt.Errorf("templates commit autosave: head content hash: %w", err)
	}
	if actualHash != cmd.ExpectedContentHash {
		if err := s.presign.Delete(ctx, version.DocxStorageKey); err != nil {
			return nil, errors.Join(domain.ErrContentHashMismatch, fmt.Errorf("delete mismatched upload: %w", err))
		}
		return nil, domain.ErrContentHashMismatch
	}

	version.ContentHash = actualHash
```
with:
```go
	vp, err := s.presign.Confirm(ctx, cmd.TenantID, version.DocxStorageKey, cmd.ExpectedContentHash)
	if err != nil {
		switch {
		case errors.Is(err, objectstore.ErrObjectMissing):
			return nil, domain.ErrUploadMissing
		case errors.Is(err, objectstore.ErrHashMismatch):
			return nil, domain.ErrContentHashMismatch
		default:
			return nil, fmt.Errorf("templates commit autosave: confirm: %w", err)
		}
	}

	version.ContentHash = vp.ContentHash
```
Add `"metaldocs/internal/platform/objectstore"` to `autosave.go` imports.

- [ ] **Step 3: Update the templates fake presigner**

In `internal/modules/templates/application/fakes_test.go`, find the fake implementing `PresignPUT/PresignGET/HeadContentHash/Delete` and replace those methods with the kernel shape. Use this (adjust the receiver/struct name to the existing fake, commonly `fakePresigner` with a `PutKeys` field referenced in `autosave_test.go:114`):
```go
func (f *fakePresigner) PresignPut(_ context.Context, _ string, key string, _ time.Duration) (string, error) {
	f.PutKeys = append(f.PutKeys, key)
	return "https://example/put", nil
}

func (f *fakePresigner) PresignGet(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://example/get/" + key, nil
}

func (f *fakePresigner) Confirm(_ context.Context, _ , key, expected string) (objectstore.VerifiedPointer, error) {
	if f.confirmErr != nil {
		return objectstore.VerifiedPointer{}, f.confirmErr
	}
	return objectstore.VerifiedPointer{StorageKey: key, ContentHash: expected, SizeBytes: 1}, nil
}

func (f *fakePresigner) Delete(_ context.Context, _ string) error { return nil }
```
Add a `confirmErr error` field to the fake struct, and an `objectstore` import. Update `autosave_test.go` cases that previously drove `HeadContentHash` returning a mismatching hash to instead set `confirmErr: objectstore.ErrHashMismatch` (mismatch case) or `objectstore.ErrObjectMissing` (missing case). The success cases need no error. Update any test that asserted the old `Delete`-on-mismatch behavior to assert the mapped domain error only.

- [ ] **Step 4: Wire main.go**

In `apps/api/cmd/metaldocs-api/main.go`, replace line 749:
```go
	templatesPresigner := objectstore.NewTemplatesPresigner(deps.MinioClient, deps.MinioPublicClient, deps.MinioBucket, 25*1024*1024)
```
with:
```go
	templatesPresigner := objectstore.NewVerifiedStore(deps.MinioClient, deps.MinioPublicClient, deps.MinioBucket, 25*1024*1024)
```

- [ ] **Step 5: Build + test**

Run: `go build ./... && go test ./internal/modules/templates/...`
Expected: PASS. Fix compile errors in templates tests by updating any other fake/mocks to the new port shape.

- [ ] **Step 6: Commit**

```bash
git add internal/modules/templates/application/ports.go internal/modules/templates/application/autosave.go internal/modules/templates/application/fakes_test.go internal/modules/templates/application/autosave_test.go apps/api/cmd/metaldocs-api/main.go
git commit -m "refactor(templates): use VerifiedStore kernel via narrowed port"
```

---

## Task 6: Migrate documents main Presigner port to the kernel

**Files:**
- Modify: `internal/modules/documents/application/service.go` (port at :68-76; call sites :528, :572-590, :644, :375, :767)
- Modify: `internal/modules/documents/application/service_test.go` (fake :307-334)
- Modify: `internal/modules/documents/application/service_review_roundtrip_integration_test.go` (fake :333-349)
- Modify: `apps/api/cmd/metaldocs-api/main.go:403`

- [ ] **Step 1: Narrow the port**

Replace `service.go:68-76` (`Presigner` interface) with:
```go
type Presigner interface {
	PresignPut(ctx context.Context, tenantID, key string, ttl time.Duration) (url string, err error)
	PresignGet(ctx context.Context, key string, ttl time.Duration) (url string, err error)
	Confirm(ctx context.Context, tenantID, key, expectedHash string) (objectstore.VerifiedPointer, error)
	Exists(ctx context.Context, key string) (bool, error)
	Size(ctx context.Context, key string) (int64, error)
	Delete(ctx context.Context, key string) error
}
```
Add `"metaldocs/internal/platform/objectstore"` and (if not present) `"time"` to `service.go` imports. Add near the other consts:
```go
const (
	revisionUploadTTL  = 15 * time.Minute
	objectDownloadTTL  = 15 * time.Minute
)
```

- [ ] **Step 2: Update PresignAutosave (build key locally)**

Replace `service.go:528`:
```go
	url, storageKey, err := s.presigner.PresignRevisionPUT(ctx, cmd.TenantID, cmd.DocumentID, cmd.ContentHash)
	if err != nil {
		return nil, err
	}
```
with:
```go
	storageKey := documentRevisionKey(cmd.TenantID, cmd.DocumentID, cmd.ContentHash)
	url, err := s.presigner.PresignPut(ctx, cmd.TenantID, storageKey, revisionUploadTTL)
	if err != nil {
		return nil, err
	}
```

- [ ] **Step 3: Update CommitAutosave (Confirm + vp.SizeBytes)**

Replace `service.go:572-590` (the `HashObject` → compare → `DeleteObject` → `SizeObject` block):
```go
	serverHash, err := s.presigner.HashObject(ctx, meta.StorageKey)
	if err != nil {
		if errors.Is(err, domain.ErrUploadMissing) {
			return nil, domain.ErrUploadMissing
		}
		return nil, fmt.Errorf("hash s3 object: %w", err)
	}
	if serverHash != meta.ExpectedContentHash {
		if err := s.presigner.DeleteObject(ctx, meta.StorageKey); err != nil {
			slog.WarnContext(ctx, "commit autosave: orphaned object cleanup failed after content-hash mismatch",
				"storage_key", meta.StorageKey, "document_id", cmd.DocumentID, "err", err)
		}
		return nil, domain.ErrContentHashMismatch
	}

	fileSizeBytes, err := s.presigner.SizeObject(ctx, meta.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("size s3 object: %w", err)
	}
```
with:
```go
	vp, err := s.presigner.Confirm(ctx, cmd.TenantID, meta.StorageKey, meta.ExpectedContentHash)
	if err != nil {
		switch {
		case errors.Is(err, objectstore.ErrObjectMissing):
			return nil, domain.ErrUploadMissing
		case errors.Is(err, objectstore.ErrHashMismatch):
			return nil, domain.ErrContentHashMismatch
		default:
			return nil, fmt.Errorf("confirm s3 object: %w", err)
		}
	}
	serverHash := vp.ContentHash
	fileSizeBytes := vp.SizeBytes
```
If `slog` becomes unused in `service.go` after this, remove its import.

- [ ] **Step 4: Update remaining call sites**

`service.go:644` (SyncArtifactMetadata):
```go
	fileSizeBytes, err := s.presigner.SizeObject(ctx, revision.StorageKey)
```
→
```go
	fileSizeBytes, err := s.presigner.Size(ctx, revision.StorageKey)
```

`service.go:375` (Exists) — change method name only:
```go
	return s.presigner.Exists(ctx, storageKey)
```
stays `Exists` (signature unchanged) — no edit needed.

`service.go:767` (signed docx URL):
```go
	return s.presigner.PresignObjectGET(ctx, rev.StorageKey)
```
→
```go
	return s.presigner.PresignGet(ctx, rev.StorageKey, objectDownloadTTL)
```

- [ ] **Step 5: Update fakes**

In `service_test.go`, replace the fake `Presigner` methods (`PresignRevisionPUT`, `SizeObject`, `AdoptTempObject`, `PresignObjectGET`, `Exists`, `DeleteObject`, `HashObject`) with the kernel shape:
```go
func (f *fakePresigner) PresignPut(_ context.Context, _, key string, _ time.Duration) (string, error) {
	return "https://example/put/" + key, nil
}
func (f *fakePresigner) PresignGet(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://example/get/" + key, nil
}
func (f *fakePresigner) Confirm(_ context.Context, _, key, expected string) (objectstore.VerifiedPointer, error) {
	if f.confirmErr != nil {
		return objectstore.VerifiedPointer{}, f.confirmErr
	}
	return objectstore.VerifiedPointer{StorageKey: key, ContentHash: expected, SizeBytes: f.size}, nil
}
func (f *fakePresigner) Exists(_ context.Context, _ string) (bool, error) { return f.exists, nil }
func (f *fakePresigner) Size(_ context.Context, _ string) (int64, error)  { return f.size, nil }
func (f *fakePresigner) Delete(_ context.Context, _ string) error         { return nil }
```
Add `confirmErr error`, `size int64`, `exists bool` fields as needed to the fake struct (reuse existing fields where present). Add `objectstore` import. **Delete** the now-orphaned `AdoptTempObject` fake method. Update any commit test that previously set a mismatching `HashObject` return to instead set `confirmErr: objectstore.ErrHashMismatch`.

In `service_review_roundtrip_integration_test.go`, apply the same replacement to `rtFakePresigner` (lines 333-349) and **delete** its `AdoptTempObject` method (line 345).

- [ ] **Step 6: Wire main.go**

Replace `main.go:403`:
```go
	docPresigner := objectstore.NewDocumentPresigner(deps.MinioClient, deps.MinioPublicClient, deps.MinioBucket, 15*time.Minute, 25*1024*1024)
```
with:
```go
	docPresigner := objectstore.NewVerifiedStore(deps.MinioClient, deps.MinioPublicClient, deps.MinioBucket, 25*1024*1024)
```

- [ ] **Step 7: Build + test**

Run: `go build ./... && go test ./internal/modules/documents/...`
Expected: build PASS. The `export_service` and `view_service` ports are still on the OLD signatures and now fail to compile against `*VerifiedStore` — that is fixed in Task 7. If `go build ./...` fails only inside export/view wiring, proceed to Task 7 before committing; otherwise commit now.

- [ ] **Step 8: Commit**

```bash
git add internal/modules/documents/application/service.go internal/modules/documents/application/service_test.go internal/modules/documents/application/service_review_roundtrip_integration_test.go apps/api/cmd/metaldocs-api/main.go
git commit -m "refactor(documents): main Presigner port on VerifiedStore kernel"
```

---

## Task 7: Migrate documents ExportPresigner + ViewPresigner ports

**Files:**
- Modify: `internal/modules/documents/application/export_service.go:20-24`, `:85`, `:103`, `:123`, `:144`
- Modify: `internal/modules/documents/application/view_service.go:16-18`, `:80`
- Modify: `internal/modules/documents/application/export_service_test.go:66-74`

- [ ] **Step 1: Narrow ExportPresigner**

Replace `export_service.go:20-24`:
```go
type ExportPresigner interface {
	PresignObjectGET(ctx context.Context, storageKey string) (url string, err error)
	HeadObject(ctx context.Context, key string) (bool, error)
	SizeObject(ctx context.Context, key string) (int64, error)
}
```
with:
```go
type ExportPresigner interface {
	PresignGet(ctx context.Context, key string, ttl time.Duration) (url string, err error)
	Exists(ctx context.Context, key string) (bool, error)
	Size(ctx context.Context, key string) (int64, error)
}
```
Add `"time"` to `export_service.go` imports and a const:
```go
const exportDownloadTTL = 15 * time.Minute
```

- [ ] **Step 2: Update ExportService call sites**

`export_service.go:85`:
```go
	headFound, err := s.presigner.HeadObject(ctx, storageKey)
```
→
```go
	headFound, err := s.presigner.Exists(ctx, storageKey)
```

`export_service.go:103`:
```go
	sizeBytes, err := s.presigner.SizeObject(ctx, storageKey)
```
→
```go
	sizeBytes, err := s.presigner.Size(ctx, storageKey)
```

`export_service.go:123`:
```go
	return s.presigner.PresignObjectGET(ctx, storageKey)
```
→
```go
	return s.presigner.PresignGet(ctx, storageKey, exportDownloadTTL)
```

`export_service.go:144`:
```go
	url, err := s.presigner.PresignObjectGET(ctx, rev.StorageKey)
```
→
```go
	url, err := s.presigner.PresignGet(ctx, rev.StorageKey, exportDownloadTTL)
```

- [ ] **Step 3: Narrow ViewPresigner**

Replace `view_service.go:16-18`:
```go
type ViewPresigner interface {
	PresignObjectGET(ctx context.Context, storageKey string) (string, error)
}
```
with:
```go
type ViewPresigner interface {
	PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error)
}
```
Add `"time"` to `view_service.go` imports and a const:
```go
const viewDownloadTTL = 15 * time.Minute
```

`view_service.go:80`:
```go
			url, err := s.presigner.PresignObjectGET(ctx, pdfKey.String)
```
→
```go
			url, err := s.presigner.PresignGet(ctx, pdfKey.String, viewDownloadTTL)
```

- [ ] **Step 4: Update export fake**

In `export_service_test.go`, replace `fakeExportPresigner` methods (lines 66-74):
```go
func (f *fakeExportPresigner) PresignObjectGET(_ context.Context, storageKey string) (string, error) { ... }
func (f *fakeExportPresigner) SizeObject(_ context.Context, _ string) (int64, error) { ... }
```
(and any `HeadObject`) with:
```go
func (f *fakeExportPresigner) PresignGet(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://example/get/" + key, nil
}
func (f *fakeExportPresigner) Exists(_ context.Context, _ string) (bool, error) { return f.headFound, nil }
func (f *fakeExportPresigner) Size(_ context.Context, _ string) (int64, error)  { return f.size, nil }
```
Add `"time"` import and any `headFound bool` / `size int64` fields the existing tests drive.

- [ ] **Step 5: Build + full documents test**

Run: `go build ./... && go test ./internal/modules/documents/... ./internal/modules/controlleddocuments/...`
Expected: PASS. controlleddocuments compiles unchanged (`CDDocumentInitializer.Exists` → `Service.Exists` → kernel `Exists`, signature identical).

- [ ] **Step 6: Commit**

```bash
git add internal/modules/documents/application/export_service.go internal/modules/documents/application/view_service.go internal/modules/documents/application/export_service_test.go
git commit -m "refactor(documents): export+view presigner ports on VerifiedStore kernel"
```

---

## Task 8: Remove the unvalidated-key presign shim (spec §5.2)

**Files:**
- Modify: `internal/modules/templates/application/queries.go:60-62` (delete `PresignStoredObject`)
- Modify: `internal/modules/templates/delivery/http/routes_generated.go:150-162` (delete `RedirectSignedUrl`)
- Modify: `internal/modules/templates/delivery/http/handler.go` (remove the route registration)
- Modify: `api/openapi/v1/openapi.yaml` (remove the `redirectSignedUrl` operation)
- Regenerate: `internal/modules/templates/api/api.gen.go`, `frontend/apps/web/src/lib/api-types/index.d.ts`

- [ ] **Step 1: Locate the route registration and openapi op**

Run: `grep -rn "RedirectSignedUrl\|redirectSignedUrl\|PresignStoredObject" internal/modules/templates api/openapi/v1/openapi.yaml`
Note each hit: the handler method, the `mux.HandleFunc`/registration line in `handler.go`, and the openapi path operation block.

- [ ] **Step 2: Delete the application method**

In `queries.go`, delete lines 60-62:
```go
func (s *Service) PresignStoredObject(ctx context.Context, key string) (string, error) {
	return s.presign.PresignGET(ctx, key, docxDownloadTTL)
}
```

- [ ] **Step 3: Delete the handler + its route**

In `routes_generated.go`, delete the entire `RedirectSignedUrl` handler (lines 150-162). In `handler.go`, delete the line registering it (the `mux.HandleFunc(... RedirectSignedUrl)` / generated registration). If the registration lives in a generated registrar inside `routes_generated.go`, delete that entry too.

- [ ] **Step 4: Remove the openapi operation**

In `api/openapi/v1/openapi.yaml`, delete the path/operation that generates `RedirectSignedUrl` (operationId `redirectSignedUrl`, the signed-url redirect GET with a `key` query param). Remove the whole path item if that operation is its only method.

- [ ] **Step 5: Regenerate contract artifacts**

Run:
```bash
go generate ./...
```
then from `frontend/apps/web`:
```bash
npm run gen:api
```
Verify zero references remain:
```bash
grep -rn "RedirectSignedUrl\|redirectSignedUrl" internal frontend/apps/web/src || echo "clean"
```
Expected: `clean`.

- [ ] **Step 6: Build + test + contract drift guard**

Run:
```bash
go build ./... && go test ./internal/modules/templates/...
```
Expected: PASS. (CI drift guard `.github/workflows/api-contract.yml` will recheck generated alignment; locally the regen above keeps it aligned.)

- [ ] **Step 7: Commit**

```bash
git add internal/modules/templates api/openapi/v1/openapi.yaml frontend/apps/web/src/lib/api-types/index.d.ts
git commit -m "fix(templates): remove unvalidated-key presign shim (cross-tenant read)"
```

---

## Task 9: Delete the old presigners and relocate isNoSuchKeyErr

**Files:**
- Delete: `internal/platform/objectstore/document_presigner.go`
- Delete: `internal/platform/objectstore/document_presigner_export.go`
- Delete: `internal/platform/objectstore/templates_presigner.go`
- Delete: `internal/platform/objectstore/document_presigner_test.go`
- Delete: `internal/platform/objectstore/templates_presigner_test.go`
- Modify: `internal/platform/objectstore/errors.go` (add `isNoSuchKeyErr`)

- [ ] **Step 1: Move isNoSuchKeyErr into errors.go**

Add to `internal/platform/objectstore/errors.go` (it currently lives in `document_presigner.go:152-168`, which is about to be deleted):
```go
import (
	"errors"
	"net/url"
	"strings"

	"github.com/minio/minio-go/v7"
)

func isNoSuchKeyErr(err error) bool {
	if err == nil {
		return false
	}
	var resp minio.ErrorResponse
	if errors.As(err, &resp) && strings.EqualFold(resp.Code, "NoSuchKey") {
		return true
	}
	if strings.Contains(err.Error(), "NoSuchKey") {
		return true
	}
	var ue *url.Error
	if errors.As(err, &ue) && strings.Contains(ue.Error(), "NoSuchKey") {
		return true
	}
	return false
}
```
Merge the `errors` import with the existing one in `errors.go` (don't duplicate the import block). Remove the temporary `var _ = errors.Is` line from `verified_store.go` if present.

- [ ] **Step 2: Delete the old files**

```bash
git rm internal/platform/objectstore/document_presigner.go \
       internal/platform/objectstore/document_presigner_export.go \
       internal/platform/objectstore/templates_presigner.go \
       internal/platform/objectstore/document_presigner_test.go \
       internal/platform/objectstore/templates_presigner_test.go
```

- [ ] **Step 3: Verify no dangling references**

Run: `grep -rn "DocumentPresigner\|TemplatesPresigner\|PresignRevisionPUT\|HeadContentHash\|HashObject\|AdoptTempObject\|PresignObjectGET\|PresignPUT\|PresignGET\|HeadObject\|SizeObject\|DeleteObject" internal apps frontend/apps/web/src || echo "clean"`
Expected: `clean` (only the new kernel method names remain). Any hit is a missed call site — fix it before continuing.

- [ ] **Step 4: Build the whole tree**

Run: `go build ./...`
Expected: PASS — `objectstore` package compiles with only `verified_store.go` + `errors.go`.

- [ ] **Step 5: Commit**

```bash
git add internal/platform/objectstore/
git commit -m "refactor(objectstore): delete duplicated presigners; relocate isNoSuchKeyErr"
```

---

## Task 10: Full verification gate

**Files:** none (verification only).

- [ ] **Step 1: Full build**

Run: `go build ./...`
Expected: PASS.

- [ ] **Step 2: Full test suite**

Run: `go test ./...`
Expected: PASS. Integration tests that require a live MinIO/Postgres are gated by the existing harness; run the same subset CI runs if the full suite needs services. Unit + sqlmock layers must be green.

- [ ] **Step 3: Confirm the invariant surface**

Run: `grep -rn "func (s \*VerifiedStore)" internal/platform/objectstore/verified_store.go`
Expected: exactly `PresignPut`, `Confirm`, `PresignGet`, `Exists`, `Size`, `Delete` (+ unexported `assertTenant`, `deleteQuiet`).

- [ ] **Step 4: Confirm tenant guard is write-path only**

Run: `grep -n "assertTenant" internal/platform/objectstore/verified_store.go`
Expected: called only inside `PresignPut` and `Confirm`.

- [ ] **Step 5: Final commit (if any test-only fixups were needed)**

```bash
git add -A
git commit -m "test(objectstore): finalize verified-store kernel migration"
```

---

## Notes for the executor

- **Test churn is expected** in templates `*_test.go` (key-format assertions) and documents fakes (port reshape). Fix assertions to the new tenant-scoped keys and kernel method shapes; do not weaken what a test verifies.
- **Error mapping is load-bearing:** `ErrObjectMissing → domain.ErrUploadMissing`, `ErrHashMismatch → domain.ErrContentHashMismatch` in both modules' commit paths. Do not collapse these — HTTP status codes depend on them.
- **No new behavior:** `Confirm` is the union of the old `HashObject/HeadContentHash` + compare + `Delete`-on-mismatch; the only functional change is that cleanup now happens inside the kernel and size comes from the same read.
- **Follow-ons (NOT in this plan):** dead-FSM cleanup; Document God-aggregate split (spec §13).
