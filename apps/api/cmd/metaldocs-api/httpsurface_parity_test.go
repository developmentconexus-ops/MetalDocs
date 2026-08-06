package main

import (
	"net/http"
	"sort"
	"testing"

	"metaldocs/internal/platform/httprouter"
)

// surfaceRecordingMux implements httprouter.Muxer over a real *http.ServeMux,
// recording every registered pattern before delegating. It is a LOCAL copy
// of permissions_test.go's recordingMux (that file's methods, at
// permissions_test.go:234-257) rather than a dependency on it: that type is
// deleted in Task 15, at which point this file is rewritten anyway (see
// mountAllCodegenHandlers below), and until then keeping this file
// self-contained means Task 15 can delete recordingMux without also having
// to touch this file's plumbing.
//
// Until Task 14 lands httprouter.NewRecorder, this stands in for it.
type surfaceRecordingMux struct {
	mux      *http.ServeMux
	patterns []string
}

func newSurfaceRecordingMux() *surfaceRecordingMux {
	return &surfaceRecordingMux{mux: http.NewServeMux()}
}

func (r *surfaceRecordingMux) Handle(pattern string, handler http.Handler) {
	r.patterns = append(r.patterns, pattern)
	r.mux.Handle(pattern, handler)
}

func (r *surfaceRecordingMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	r.patterns = append(r.patterns, pattern)
	r.mux.HandleFunc(pattern, handler)
}

func (r *surfaceRecordingMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *surfaceRecordingMux) Patterns() []string {
	return r.patterns
}

var _ httprouter.Muxer = (*surfaceRecordingMux)(nil)

// mountAllCodegenHandlers constructs one instance of every module handler
// with nil inner services and mounts them all onto mux. Nil services are
// safe: registration hands method *values* to the mux and never invokes
// them — buildTestRouteHandlers (permissions_test.go) and buildRouter
// (router.go) are production's own construction/mount path, reused here
// rather than re-implemented, so this test observes exactly what main()
// mounts.
//
// This helper is superseded by Task 15's publisher list — at that point
// rewrite this test to iterate `publishers` instead, per that task's own
// commit note.
func mountAllCodegenHandlers(t *testing.T, mux httprouter.Muxer) {
	t.Helper()
	buildRouter(mux, buildTestRouteHandlers())
}

// TestGeneratedKeysMatchCodegenPatterns proves the generator's key formula
// and oapi-codegen's HandlerWithOptions registration formula produce the same
// bytes for every operation that goes through codegen. It is exhaustive by
// construction: it walks httpSurface, not a sampled list.
//
// The boot assertion (Task 15) proves the same thing on every real boot;
// this test proves it in CI without a database.
func TestGeneratedKeysMatchCodegenPatterns(t *testing.T) {
	rec := newSurfaceRecordingMux()
	mountAllCodegenHandlers(t, rec) // defined above; nil services are safe at mount time

	mounted := map[string]bool{}
	for _, p := range rec.Patterns() {
		mounted[p] = true
	}

	var missing []string
	for key := range httpSurface {
		if !mounted[key] {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("%d declared keys were not registered by codegen with the same bytes:\n%v", len(missing), missing)
	}
}
