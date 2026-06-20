package middleware

import (
	"bufio"
	"net"
	"net/http"
	"strings"

	"metaldocs/internal/platform/problem"
)

// MethodNotAllowedJSON converts the stdlib text/plain 405 and 404 responses
// that Go 1.22 method-routed ServeMux emits (via http.Error) into RFC 9457
// application/problem+json, preserving the Allow header on a 405 (D-03: no
// bare-status error responses). Method-prefixed patterns reject the request
// before any handler runs, so an in-body guard cannot fix the envelope — a
// response-rewriting interceptor at the mux boundary fixes every route at once.
//
// Hand-coded 404/405 responses already use application/problem+json, so they
// never match the text/plain guard and pass through untouched.
func MethodNotAllowedJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ic := &problemInterceptor{ResponseWriter: w}
		next.ServeHTTP(ic, r)
		if !ic.intercepted {
			return
		}
		// The mux already set Allow (on 405) into w.Header(); problem.Write
		// only overwrites Content-Type, so the header is preserved.
		switch ic.status {
		case http.StatusMethodNotAllowed:
			_ = problem.Write(w, problem.New(http.StatusMethodNotAllowed, problem.CodeMethodNotAllowed, "Method not allowed"))
		default:
			_ = problem.Write(w, problem.New(http.StatusNotFound, problem.CodeNotFound, "Not found"))
		}
	})
}

// problemInterceptor swallows a stdlib text/plain 404/405 (status + body) so
// the surrounding middleware can re-emit it as problem+json. Any other response
// passes straight through to the underlying writer.
type problemInterceptor struct {
	http.ResponseWriter
	intercepted bool
	status      int
}

func (w *problemInterceptor) WriteHeader(code int) {
	if (code == http.StatusNotFound || code == http.StatusMethodNotAllowed) &&
		strings.HasPrefix(w.Header().Get("Content-Type"), "text/plain") {
		w.intercepted = true
		w.status = code
		return // swallow; MethodNotAllowedJSON re-emits as problem+json
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *problemInterceptor) Write(b []byte) (int, error) {
	if w.intercepted {
		return len(b), nil // discard the stdlib text/plain body
	}
	return w.ResponseWriter.Write(b)
}

// Flush / Hijack delegate to the underlying writer when supported so streaming
// (SSE) and connection-upgrade (websocket) routes are unaffected by the wrap.
func (w *problemInterceptor) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *problemInterceptor) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}
