package application

import (
	"context"
	"errors"
	"testing"

	"metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/documents/infrastructure"
	templatesdomain "metaldocs/internal/modules/templates/domain"
)

type guardWriter struct {
	source       string
	exists       bool
	upsertCalled bool
}

func (g *guardWriter) UpsertValue(context.Context, infrastructure.PlaceholderValue, ...infrastructure.DBTX) error {
	return nil
}
func (g *guardWriter) UpsertAuthorValue(_ context.Context, _ infrastructure.PlaceholderValue, _ ...infrastructure.DBTX) (int64, error) {
	g.upsertCalled = true
	return 1, nil
}
func (g *guardWriter) CurrentSource(_ context.Context, _, _, _ string) (string, bool, error) {
	return g.source, g.exists, nil
}

// newFillInServiceForGuardTest builds a *FillInService with:
//   - a permissive TxRunner (mirrors the existing fillin_service_test.go pattern)
//   - a fakeSchemaReader returning a single text placeholder with ID="ph1"
//   - the supplied writer as the FillInWriter
func newFillInServiceForGuardTest(t *testing.T, w FillInWriter) *FillInService {
	t.Helper()
	schema := []templatesdomain.Placeholder{{ID: "ph1", Type: templatesdomain.PHText}}
	return NewFillInService(
		newTxRunner(newPermissiveFillInDB(t)),
		fakeSchemaReader{placeholders: schema},
		w,
	)
}

func TestSetPlaceholderValue_GovernedSourceRejected(t *testing.T) {
	// Construct the service with a stub schema reader returning a matching placeholder.
	w := &guardWriter{source: "dictionary", exists: true}
	svc := newFillInServiceForGuardTest(t, w)
	err := svc.SetPlaceholderValue(context.Background(), "t1", "actor", "rev1", "ph1", "hack")
	if !errors.Is(err, domain.ErrPlaceholderNotAuthorEditable) {
		t.Fatalf("want ErrPlaceholderNotAuthorEditable, got %v", err)
	}
	if w.upsertCalled {
		t.Fatal("guard must reject before attempting the write")
	}
}
