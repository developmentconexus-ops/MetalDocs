package objectstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

const defaultMaxObjectBytes = 25 * 1024 * 1024

// SystemTenantID is the well-known sentinel for system-owned objects (e.g. built-in
// templates). Copy allows a system-tenant srcKey to be copied into any tenant.
const SystemTenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"

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

// KeyHasTenantPrefix reports whether key resides inside tenantID's namespace
// (tenants/{tenantID}/…). It is the single source of truth for the tenant-prefix
// rule: assertTenant, Copy's srcKey check, and out-of-package callers (e.g. the
// PDF-completion webhook) all delegate here so the rule cannot silently diverge.
// An empty tenantID never matches — without this guard a "" tenant would reduce
// the check to the prefix "tenants//", which a crafted "tenants//…" key would
// satisfy, bypassing isolation at the kernel boundary.
func KeyHasTenantPrefix(tenantID, key string) bool {
	if tenantID == "" {
		return false
	}
	return strings.HasPrefix(key, "tenants/"+tenantID+"/")
}

func (s *VerifiedStore) assertTenant(tenantID, key string) error {
	if !KeyHasTenantPrefix(tenantID, key) {
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

// Copy duplicates an existing object to a new tenant-scoped key, server-side
// (no bytes stream through the app, so the producer invariant is preserved — no
// docx is authored by the Go server). Both srcKey and dstKey are tenant-prefix
// guarded; the only exception is a srcKey owned by the system tenant
// (SystemTenantID), which may be copied into any tenant's namespace (e.g. built-in
// template copy-on-spawn). Used by copy-on-spawn so each template version owns a
// distinct object instead of sharing a key with its source.
func (s *VerifiedStore) Copy(ctx context.Context, tenantID, srcKey, dstKey string) error {
	if err := s.assertTenant(tenantID, dstKey); err != nil {
		return err
	}
	// Assert srcKey belongs to the same tenant OR to the system tenant — explicit
	// exception, not an omission of the check.
	if !KeyHasTenantPrefix(tenantID, srcKey) && !KeyHasTenantPrefix(SystemTenantID, srcKey) {
		return ErrKeyOutsideTenant
	}
	_, err := s.client.CopyObject(ctx,
		minio.CopyDestOptions{Bucket: s.bucket, Object: dstKey},
		minio.CopySrcOptions{Bucket: s.bucket, Object: srcKey},
	)
	if err != nil {
		if isNoSuchKeyErr(err) {
			return ErrObjectMissing
		}
		return fmt.Errorf("objectstore: copy: %w", err)
	}
	return nil
}

// --- read path ---

// AssertedPresignGet is a tenant-guarded variant of PresignGet for callers that
// derive the key from user-controlled input rather than from a DB-sourced column.
// It mirrors PresignPut (:55-63) in taking a tenantID and asserting the prefix
// before issuing the presigned URL. DB-sourced callers should use PresignGet.
func (s *VerifiedStore) AssertedPresignGet(ctx context.Context, tenantID, key string, ttl time.Duration) (string, error) {
	if err := s.assertTenant(tenantID, key); err != nil {
		return "", err
	}
	return s.PresignGet(ctx, key, ttl)
}

// PresignGet presigns a GET URL for a DB-sourced / server-trusted key (NOT guarded).
func (s *VerifiedStore) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	u, err := s.signingClient.PresignedGetObject(ctx, s.bucket, key, ttl, nil)
	if err != nil {
		return "", fmt.Errorf("objectstore: presign get: %w", err)
	}
	return u.String(), nil
}

// AssertedReadObject is the tenant-guarded read primitive: it asserts key resides
// in tenantID's namespace OR the system tenant's (mirroring Copy :114-123 — the
// schema files it reads may be system-owned built-in templates) before delegating
// to ReadObject. Callers that hold a tenantID should use this rather than the
// unguarded ReadObject so a poisoned/mis-scoped DB key cannot fetch a cross-tenant
// object — defense-in-depth above the SQL tenant scoping.
func (s *VerifiedStore) AssertedReadObject(ctx context.Context, tenantID, key string, maxBytes int64) ([]byte, error) {
	if !KeyHasTenantPrefix(tenantID, key) && !KeyHasTenantPrefix(SystemTenantID, key) {
		return nil, ErrKeyOutsideTenant
	}
	return s.ReadObject(ctx, key, maxBytes)
}

// ReadObject fetches an object and returns its bytes, enforcing a caller-supplied
// maxBytes size limit (mirrors Confirm :68-100 — read + LimitReader + size check —
// but without a hash comparison). Intended for trusted internal reads (e.g. schema
// files) where the caller knows the key from the DB but does not have an expected
// hash. Returns ErrObjectMissing when the object does not exist and
// ErrObjectTooLarge when the size limit is exceeded (the object is NOT deleted).
// Prefer AssertedReadObject when a tenantID is in scope.
func (s *VerifiedStore) ReadObject(ctx context.Context, key string, maxBytes int64) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		if isNoSuchKeyErr(err) {
			return nil, ErrObjectMissing
		}
		return nil, fmt.Errorf("objectstore: read get: %w", err)
	}
	defer obj.Close()

	payload, err := io.ReadAll(io.LimitReader(obj, maxBytes+1))
	if err != nil {
		if isNoSuchKeyErr(err) {
			return nil, ErrObjectMissing
		}
		return nil, fmt.Errorf("objectstore: read body: %w", err)
	}
	if int64(len(payload)) > maxBytes {
		return nil, ErrObjectTooLarge
	}
	return payload, nil
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

// --- enumeration (tenant-guarded) ---

// ObjectInfo is the minimal metadata an enumeration returns for one stored
// object. It deliberately exposes only the fields a reconciliation janitor needs
// (key + last-modified age) and holds no domain knowledge.
type ObjectInfo struct {
	Key          string
	LastModified time.Time
	SizeBytes    int64
}

// ListTenantObjects enumerates every object whose key starts with prefix, which
// MUST reside inside tenantID's namespace (tenants/{tenantID}/…). The prefix is
// asserted through the same KeyHasTenantPrefix rule that guards the write/read
// paths, so an enumeration can never escape the tenant boundary — a reconciliation
// sweeper listing one tenant's prefix can never observe (and therefore never
// delete) another tenant's objects. Mirrors minio ListObjects with recursive
// enumeration under the prefix.
func (s *VerifiedStore) ListTenantObjects(ctx context.Context, tenantID, prefix string) ([]ObjectInfo, error) {
	if !KeyHasTenantPrefix(tenantID, prefix) {
		return nil, ErrKeyOutsideTenant
	}
	var out []ObjectInfo
	for obj := range s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}) {
		if obj.Err != nil {
			return nil, fmt.Errorf("objectstore: list: %w", obj.Err)
		}
		out = append(out, ObjectInfo{
			Key:          obj.Key,
			LastModified: obj.LastModified,
			SizeBytes:    obj.Size,
		})
	}
	return out, nil
}
