package application_test

import (
	"context"
	"testing"

	"metaldocs/internal/modules/documents/application"
)

type fakeProfileDefaultTemplateReader struct {
	id     *string
	status *string
	err    error
	calls  int
}

func (f *fakeProfileDefaultTemplateReader) GetDefaultTemplateVersionID(_ context.Context, _, _ string) (*string, *string, error) {
	f.calls++
	return f.id, f.status, f.err
}

func strptr(v string) *string { return &v }

// TestResolveTemplateVersionID_ShortCircuit proves that when a concrete override
// id is supplied the default-template reader is NOT called (short-circuit), and
// when no override is supplied it IS called. This is the key property that
// prevents the atomic CD-create self-deadlock: the in-tx clone never issues an
// authz-recording GetByCode taxonomy read while holding the advisory lock.
func TestResolveTemplateVersionID_ShortCircuit(t *testing.T) {
	defaultID := "00000000-0000-0000-0000-000000000001"
	defaultStatus := "published"

	t.Run("with concrete override: default NOT fetched", func(t *testing.T) {
		reader := &fakeProfileDefaultTemplateReader{
			id:     &defaultID,
			status: &defaultStatus,
		}
		overrideID := "00000000-0000-0000-0000-000000000099"
		svc := application.NewServiceForTest(reader)
		got, err := svc.ExportedResolveTemplateVersionID(context.Background(), "tenant-a", "ISO", &overrideID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != overrideID {
			t.Errorf("expected override id %q, got %q", overrideID, got)
		}
		if reader.calls != 0 {
			t.Errorf("GetDefaultTemplateVersionID should NOT be called when override is supplied; got %d call(s)", reader.calls)
		}
	})

	t.Run("without override: default IS fetched", func(t *testing.T) {
		reader := &fakeProfileDefaultTemplateReader{
			id:     &defaultID,
			status: &defaultStatus,
		}
		svc := application.NewServiceForTest(reader)
		got, err := svc.ExportedResolveTemplateVersionID(context.Background(), "tenant-a", "ISO", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != defaultID {
			t.Errorf("expected default id %q, got %q", defaultID, got)
		}
		if reader.calls != 1 {
			t.Errorf("GetDefaultTemplateVersionID should be called exactly once when no override; got %d call(s)", reader.calls)
		}
	})
}
