// Package middleware holds platform HTTP middleware that composes the
// outermost layers of the API chain (target architecture §2.1).
package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/requesttrace"
)

// Recovery is the outermost middleware (REQ-MW-1, REQ-MW-2). It establishes
// the request trace ID before anything else runs, and converts a downstream
// panic into a trace-ID-tagged log line plus a 500 problem+json response
// instead of a killed connection. http.ErrAbortHandler is re-panicked per the
// net/http contract (intentional connection abort, not a defect).
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID, ok := requesttrace.Normalize(r.Header.Get("X-Trace-Id"))
		if !ok {
			traceID = requesttrace.Resolve(r.Context())
		}
		r = r.WithContext(requesttrace.WithTraceID(r.Context(), traceID))

		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			if recErr, ok := rec.(error); ok && errors.Is(recErr, http.ErrAbortHandler) {
				panic(rec)
			}
			slog.Error("http panic recovered",
				"trace_id", traceID,
				"method", r.Method,
				"path", r.URL.Path,
				"panic", fmt.Sprint(rec),
				"stack", string(debug.Stack()),
			)
			// Best effort: if the handler already started writing the
			// response this produces a corrupt body, which is still
			// preferable to a silently dropped connection.
			_ = problem.Write(w, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Internal server error"))
		}()
		next.ServeHTTP(w, r)
	})
}
