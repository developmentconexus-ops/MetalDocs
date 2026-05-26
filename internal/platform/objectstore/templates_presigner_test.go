package objectstore

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func TestTemplatesPresigner_UsesPublicSigningEndpointForBrowserURLs(t *testing.T) {
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

	p := NewTemplatesPresigner(internalClient, publicClient, "metaldocs-attachments", 25*1024*1024)

	putURL, err := p.PresignPUT(context.Background(), "templates/test.docx", 15*time.Minute)
	if err != nil {
		t.Fatalf("PresignPUT: %v", err)
	}
	parsedPut, err := url.Parse(putURL)
	if err != nil {
		t.Fatalf("parse put url: %v", err)
	}
	if parsedPut.Host != "127.0.0.1:9000" {
		t.Fatalf("put host = %q, want %q", parsedPut.Host, "127.0.0.1:9000")
	}

	getURL, err := p.PresignGET(context.Background(), "templates/test.docx", 15*time.Minute)
	if err != nil {
		t.Fatalf("PresignGET: %v", err)
	}
	parsedGet, err := url.Parse(getURL)
	if err != nil {
		t.Fatalf("parse get url: %v", err)
	}
	if parsedGet.Host != "127.0.0.1:9000" {
		t.Fatalf("get host = %q, want %q", parsedGet.Host, "127.0.0.1:9000")
	}
}
