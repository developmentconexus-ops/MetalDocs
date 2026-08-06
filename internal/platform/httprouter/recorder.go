package httprouter

import "net/http"

// Recorder is a Muxer that records every pattern registered through it and
// delegates to a real muxer.
//
// It runs in PRODUCTION startup, not only in a test: the composition root
// mounts every publisher through one Recorder each and feeds the result to
// assertSurface. The recorder wraps the real mount, the assertion runs on the
// real boot, and the check IS the invariant rather than a mirror of it.
type Recorder struct {
	inner    Muxer
	patterns []string
}

// NewRecorder wraps inner. A nil inner is a wiring bug and panics.
func NewRecorder(inner Muxer) *Recorder {
	if inner == nil {
		panic("httprouter: NewRecorder called with a nil Muxer")
	}
	return &Recorder{inner: inner}
}

func (r *Recorder) Handle(pattern string, handler http.Handler) {
	r.patterns = append(r.patterns, pattern)
	r.inner.Handle(pattern, handler)
}

func (r *Recorder) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	r.patterns = append(r.patterns, pattern)
	r.inner.HandleFunc(pattern, handler)
}

func (r *Recorder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.inner.ServeHTTP(w, req)
}

// Patterns returns every pattern registered through this recorder, in
// registration order.
func (r *Recorder) Patterns() []string {
	out := make([]string, len(r.patterns))
	copy(out, r.patterns)
	return out
}
