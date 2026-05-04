package application

import (
	"context"
	"errors"
	"testing"

	"metaldocs/internal/modules/documents/domain"
)

type stubTemplateReader struct {
	snap domain.TemplateSnapshot
	err  error
}

func (s stubTemplateReader) LoadForSnapshot(_ context.Context, _, _ string) (domain.TemplateSnapshot, error) {
	return s.snap, s.err
}

func TestResolveTemplate_ReturnsSnapshotAndPlaceholders(t *testing.T) {
	want := domain.TemplateSnapshot{
		PlaceholderSchemaJSON: []byte(`[{"id":"author","required":true}]`),
	}
	svc := NewSnapshotService(stubTemplateReader{snap: want}, nil)

	got, phs, err := svc.ResolveTemplate(context.Background(), "tenant-1", "tmpl-1")
	if err != nil {
		t.Fatalf("ResolveTemplate: %v", err)
	}
	if string(got.PlaceholderSchemaJSON) != string(want.PlaceholderSchemaJSON) {
		t.Fatalf("snapshot mismatch")
	}
	if len(phs) != 1 || phs[0].ID != "author" {
		t.Fatalf("placeholders = %+v", phs)
	}
}

func TestResolveTemplate_ReaderError(t *testing.T) {
	svc := NewSnapshotService(stubTemplateReader{err: errors.New("boom")}, nil)
	if _, _, err := svc.ResolveTemplate(context.Background(), "t", "x"); err == nil {
		t.Fatal("expected error")
	}
}
