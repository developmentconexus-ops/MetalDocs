package objectstore

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func TestDocumentPresigner_UsesPublicSigningEndpointForBrowserURLs(t *testing.T) {
	internalClient, err := minio.New("minio:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("minioadmin", "minioadmin", ""),
		Secure: false,
		Region: "us-east-1",
	})
	if err != nil {
		t.Fatalf("internal client: %v", err)
	}
	publicClient, err := minio.New("127.0.0.1:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("minioadmin", "minioadmin", ""),
		Secure: false,
		Region: "us-east-1",
	})
	if err != nil {
		t.Fatalf("public client: %v", err)
	}

	p := NewDocumentPresigner(internalClient, publicClient, "metaldocs-attachments", 15*time.Minute, 25*1024*1024)

	rawURL, _, err := p.PresignRevisionPUT(context.Background(), "tenant-a", "doc-a", "hash-a")
	if err != nil {
		t.Fatalf("PresignRevisionPUT: %v", err)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse put url: %v", err)
	}
	if parsed.Host != "127.0.0.1:9000" {
		t.Fatalf("put host = %q, want %q", parsed.Host, "127.0.0.1:9000")
	}

	getURL, err := p.PresignObjectGET(context.Background(), "system/templates/blank.docx")
	if err != nil {
		t.Fatalf("PresignObjectGET: %v", err)
	}
	parsedGet, err := url.Parse(getURL)
	if err != nil {
		t.Fatalf("parse get url: %v", err)
	}
	if parsedGet.Host != "127.0.0.1:9000" {
		t.Fatalf("get host = %q, want %q", parsedGet.Host, "127.0.0.1:9000")
	}
}
