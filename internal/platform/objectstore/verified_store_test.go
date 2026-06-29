package objectstore

import (
	"context"
	"errors"
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

func TestVerifiedStore_Copy_DestOutsideTenant_Rejected(t *testing.T) {
	s := newTestStore(t)
	err := s.Copy(context.Background(), "tenant-a",
		"tenants/tenant-a/templates/tpl-1/versions/1.docx",
		"tenants/OTHER/templates/tpl-1/versions/2.docx")
	if err != ErrKeyOutsideTenant {
		t.Fatalf("expected ErrKeyOutsideTenant, got %v", err)
	}
}

func TestVerifiedStore_Copy_ValidDestNotRejected(t *testing.T) {
	s := newTestStore(t)
	// Short deadline so the guard still runs synchronously (passes) but the doomed
	// minio network call fails fast instead of retrying against an unreachable host.
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	err := s.Copy(ctx, "tenant-a",
		"tenants/tenant-a/templates/tpl-1/versions/1.docx",
		"tenants/tenant-a/templates/tpl-1/versions/2.docx")
	// Guard passes; minio call fails (no live server) — must NOT be ErrKeyOutsideTenant.
	if errors.Is(err, ErrKeyOutsideTenant) {
		t.Fatalf("valid dst should not be rejected by guard")
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
