package docgenv2

import (
	"testing"

	"metaldocs/internal/modules/documents/application"
)

func TestNewTemplatesSnapshotReader(t *testing.T) {
	var _ application.SnapshotTemplateReader = (*TemplatesSnapshotReader)(nil)

	r := NewTemplatesSnapshotReader(nil)
	if r == nil {
		t.Fatal("expected reader")
	}
}
