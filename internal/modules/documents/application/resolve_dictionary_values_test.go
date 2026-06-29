package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"metaldocs/internal/modules/documents/domain"
	templatesdomain "metaldocs/internal/modules/templates/domain"
)

type fakeDictReader struct {
	val   string
	found bool
	err   error
}

func (f fakeDictReader) Lookup(_ context.Context, _, _ string) (string, bool, error) {
	return f.val, f.found, f.err
}

type fakeSnapReader struct{ schema []byte }

func (f fakeSnapReader) LoadForSnapshot(_ context.Context, _, _ string) (domain.TemplateSnapshot, error) {
	return domain.TemplateSnapshot{PlaceholderSchemaJSON: f.schema}, nil
}

func schemaJSON(t *testing.T, phs []templatesdomain.Placeholder) []byte {
	t.Helper()
	b, err := json.Marshal(phs)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestResolveDictionaryValues(t *testing.T) {
	schema := schemaJSON(t, []templatesdomain.Placeholder{
		{ID: "d1", Type: templatesdomain.PHDictionary, Name: "company_name"},
		{ID: "u1", Type: templatesdomain.PHText, Label: "free"},
	})

	t.Run("found pins by placeholder id", func(t *testing.T) {
		svc := &Service{
			snapshotSvc: NewSnapshotService(fakeSnapReader{schema: schema}),
		}
		svc.WithDictionaryReader(fakeDictReader{val: "ACME", found: true})
		got, err := svc.ResolveDictionaryValues(context.Background(), "t1", "v1")
		if err != nil {
			t.Fatal(err)
		}
		if got["d1"] != "ACME" || len(got) != 1 {
			t.Fatalf("want {d1:ACME}, got %v", got)
		}
	})

	t.Run("missing token -> ErrDictionaryTokenMissing", func(t *testing.T) {
		svc := &Service{
			snapshotSvc: NewSnapshotService(fakeSnapReader{schema: schema}),
		}
		svc.WithDictionaryReader(fakeDictReader{found: false})
		_, err := svc.ResolveDictionaryValues(context.Background(), "t1", "v1")
		if !errors.Is(err, domain.ErrDictionaryTokenMissing) {
			t.Fatalf("want ErrDictionaryTokenMissing, got %v", err)
		}
	})

	t.Run("reader error propagates (not mis-mapped to missing)", func(t *testing.T) {
		svc := &Service{
			snapshotSvc: NewSnapshotService(fakeSnapReader{schema: schema}),
		}
		boom := errors.New("authz denied")
		svc.WithDictionaryReader(fakeDictReader{err: boom})
		_, err := svc.ResolveDictionaryValues(context.Background(), "t1", "v1")
		if errors.Is(err, domain.ErrDictionaryTokenMissing) || !errors.Is(err, boom) {
			t.Fatalf("want raw reader error, got %v", err)
		}
	})
}
