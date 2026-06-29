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
// docx is authored by the Go server). The DESTINATION is tenant-prefix guarded,
// same as the write path; the SOURCE is a DB-sourced / server-trusted key (read
// path, not guarded). Used by copy-on-spawn so each template version owns a
// distinct object instead of sharing a key with its source.
func (s *VerifiedStore) Copy(ctx context.Context, tenantID, srcKey, dstKey string) error {
	if err := s.assertTenant(tenantID, dstKey); err != nil {
		return err
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
