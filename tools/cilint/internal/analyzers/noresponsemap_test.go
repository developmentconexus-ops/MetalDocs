package analyzers_test

import (
	"os"
	"path/filepath"
	"testing"

	"metaldocs/tools/cilint/internal/analyzers"
)

func noResponseMapFixture(t *testing.T, relPath, src string) string {
	t.Helper()
	dir := t.TempDir()
	full := filepath.Join(dir, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(full, []byte(src), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return full
}

func TestNoResponseMap_Positive_DirectLiteral(t *testing.T) {
	src := `package httpdelivery
func (h *Handler) View(w interface{}) {
	writeJSON(w, 200, map[string]any{"id": "x"})
}
`
	path := noResponseMapFixture(t, "internal/modules/foo/delivery/http/handler.go", src)
	findings := analyzers.NoResponseMap([]string{path})
	if len(findings) == 0 {
		t.Fatal("expected finding for direct map[string]any response literal")
	}
}

func TestNoResponseMap_Positive_MapStringString(t *testing.T) {
	// M9 / F9.4: the post-M8 re-audit found map[string]string response literals
	// (documents duplicate/comment/revision-url) that the any-only check let
	// through. The widened gate must flag ANY map[string]<T> reaching a writer.
	src := `package httpdelivery
func (h *Handler) Duplicate(w interface{}) {
	WriteJSON(w, 201, map[string]string{"document_id": "x"})
}
`
	path := noResponseMapFixture(t, "internal/modules/documents/delivery/http/handler.go", src)
	findings := analyzers.NoResponseMap([]string{path})
	if len(findings) == 0 {
		t.Fatal("expected finding for map[string]string response literal (F9.4 widened scope)")
	}
}

func TestNoResponseMap_Negative_NonResponseMapStringString(t *testing.T) {
	// A map[string]string that never reaches a 2xx writer must still pass —
	// the widening is response-safe, scoped to writer-reaching literals only.
	src := `package httpdelivery
func (h *Handler) View(w interface{}) {
	labels := map[string]string{"k": "v"}
	h.metrics.Tag(labels)
}
`
	path := noResponseMapFixture(t, "internal/modules/foo/delivery/http/handler.go", src)
	findings := analyzers.NoResponseMap([]string{path})
	if len(findings) != 0 {
		t.Fatalf("non-response map[string]string must not be flagged, got %d: %+v", len(findings), findings)
	}
}

func TestNoResponseMap_Positive_BuiltThenWrittenLocal(t *testing.T) {
	// The exact laundering the historical grep was blind to.
	src := `package httpdelivery
func (h *Handler) List(w interface{}) {
	page := map[string]any{"items": nil}
	writeJSON(w, 200, page)
}
`
	path := noResponseMapFixture(t, "internal/modules/foo/delivery/http/handler.go", src)
	findings := analyzers.NoResponseMap([]string{path})
	if len(findings) == 0 {
		t.Fatal("expected finding for built-then-written map[string]any local")
	}
}

func TestNoResponseMap_Positive_WriteJSONAlias_PresenceScope(t *testing.T) {
	src := `package presence
func (h *Handler) Stream(w interface{}) {
	WriteJSON(w, 200, map[string]any{"online": true})
}
`
	path := noResponseMapFixture(t, "internal/modules/iam/presence/handler.go", src)
	findings := analyzers.NoResponseMap([]string{path})
	if len(findings) == 0 {
		t.Fatal("expected finding for WriteJSON alias on presence route")
	}
}

func TestNoResponseMap_Negative_TypedBody(t *testing.T) {
	src := `package httpdelivery
type Resp struct{ ID string }
func (h *Handler) View(w interface{}) {
	writeJSON(w, 200, Resp{ID: "x"})
}
`
	path := noResponseMapFixture(t, "internal/modules/foo/delivery/http/handler.go", src)
	findings := analyzers.NoResponseMap([]string{path})
	if len(findings) != 0 {
		t.Fatalf("typed struct body must not be flagged, got %d: %+v", len(findings), findings)
	}
}

func TestNoResponseMap_Negative_NonResponseMap(t *testing.T) {
	// map[string]any that never reaches a 2xx writer (e.g. an audit payload).
	src := `package httpdelivery
func (h *Handler) View(w interface{}) {
	payload := map[string]any{"action": "viewed"}
	h.audit.Write(payload)
}
`
	path := noResponseMapFixture(t, "internal/modules/foo/delivery/http/handler.go", src)
	findings := analyzers.NoResponseMap([]string{path})
	if len(findings) != 0 {
		t.Fatalf("non-response map[string]any must not be flagged, got %d: %+v", len(findings), findings)
	}
}

func TestNoResponseMap_Negative_ExemptHealthFile(t *testing.T) {
	src := `package observability
func live(w interface{}) {
	writeJSON(w, 200, map[string]any{"status": "live"})
}
`
	path := noResponseMapFixture(t, "internal/platform/observability/health.go", src)
	findings := analyzers.NoResponseMap([]string{path})
	if len(findings) != 0 {
		t.Fatalf("recorded health.go exemption must not be flagged, got %d: %+v", len(findings), findings)
	}
}

func TestNoResponseMap_Negative_AllowDirective(t *testing.T) {
	src := `package httpdelivery
func (h *Handler) Webhook(w interface{}) {
	writeJSON(w, 200, map[string]any{"ok": true}) //cilint:allow-responsemap off-spec webhook
}
`
	path := noResponseMapFixture(t, "internal/modules/foo/delivery/http/handler.go", src)
	findings := analyzers.NoResponseMap([]string{path})
	if len(findings) != 0 {
		t.Fatalf("allow directive must suppress finding, got %d: %+v", len(findings), findings)
	}
}

func TestNoResponseMap_OutOfScope_ApplicationLayer(t *testing.T) {
	src := `package application
func (s *Svc) emit(w interface{}) {
	writeJSON(w, 200, map[string]any{"id": "x"})
}
`
	path := noResponseMapFixture(t, "internal/modules/foo/application/svc.go", src)
	findings := analyzers.NoResponseMap([]string{path})
	if len(findings) != 0 {
		t.Fatalf("application layer is out of the registered-route surface, got %d: %+v", len(findings), findings)
	}
}
