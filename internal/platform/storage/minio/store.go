package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"

	miniogo "github.com/minio/minio-go/v7"

	"metaldocs/internal/platform/config"
)

// Store is an object-storage adapter backed by a MinIO (S3-compatible) client.
type Store struct {
	client           *miniogo.Client
	bucket           string
	autoCreateBucket bool
}

// NewStore creates a Store using an already-initialised MinIO client.
// The caller is responsible for constructing the client (credentials, endpoint, TLS).
func NewStore(client *miniogo.Client, cfg config.AttachmentsConfig) *Store {
	return &Store{
		client:           client,
		bucket:           cfg.MinIOBucket,
		autoCreateBucket: cfg.MinIOAutoCreateBucket,
	}
}

// EnsureBucket verifies the configured bucket exists, creating it when
// autoCreateBucket is enabled and it does not.
func (s *Store) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check minio bucket: %w", err)
	}
	if exists {
		return nil
	}
	if !s.autoCreateBucket {
		return fmt.Errorf("minio bucket %q does not exist and auto create is disabled", s.bucket)
	}
	if err := s.client.MakeBucket(ctx, s.bucket, miniogo.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create minio bucket: %w", err)
	}
	return nil
}

// Save writes content to storageKey in the configured bucket.
func (s *Store) Save(ctx context.Context, storageKey string, content []byte) error {
	reader := bytes.NewReader(content)
	_, err := s.client.PutObject(ctx, s.bucket, storageKey, reader, int64(len(content)), miniogo.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return fmt.Errorf("save attachment content in minio: %w", err)
	}
	return nil
}

// Open returns a reader for the object at storageKey in the configured bucket.
func (s *Store) Open(ctx context.Context, storageKey string) (io.ReadCloser, error) {
	object, err := s.client.GetObject(ctx, s.bucket, storageKey, miniogo.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("open attachment content from minio: %w", err)
	}
	if _, err := object.Stat(); err != nil {
		_ = object.Close()
		return nil, fmt.Errorf("stat attachment content from minio: %w", err)
	}
	return object, nil
}

// Delete removes the object at storageKey from the configured bucket.
func (s *Store) Delete(ctx context.Context, storageKey string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, storageKey, miniogo.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete attachment content from minio: %w", err)
	}
	return nil
}
