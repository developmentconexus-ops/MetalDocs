package problem

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"metaldocs/internal/platform/requesttrace"
)

// THE SINGLE TERMINAL PROBLEM WRITER (G-07 / R-ERR-1).
//
// Every MetalDocs HTTP error response is serialized here and nowhere else.
// `write` is unexported on purpose: a delivery package cannot reach the
// serialization step at all, so a local `writeProblem` clone is not merely
// banned by lint, it is unrepresentable. The cilint `problemwriter` analyzer
// closes the remaining hole (hand-rolling the media type).
//
// What callers keep owning: BUILDING the *Problem (New / NewFor / a module's
// error-translation seam, R-ERR-2). What they no longer own: status/header/body
// emission, trace correlation, and error logging.
//
// LOGGING POLICY (deliberate, and deliberately not symmetric between 4xx/5xx).
// HTTPObservability.Wrap already emits exactly one access-log line per request
// carrying method, route, status and trace id, so re-logging every 4xx here
// would double-log the same fact. The writer therefore adds a line only where
// the access log is insufficient:
//
//	>= 500          -> slog.Error. The client body is intentionally generic, so
//	                   without this line the cause is lost (the "silent 500"
//	                   defect in audit pass06-08 §7.3).
//	4xx with cause  -> slog.Debug. Client-caused; the access log already
//	                   records it, the cause is diagnostic-only.
//	4xx no cause    -> no line. Access log covers it.
//	< 400           -> slog.Warn. A Problem with a success status is a
//	                   programming defect; surface it without dropping the body.
//
// Only the fields listed below are logged. Request bodies, headers, cookies and
// query strings are never logged here: they are the carriers of secrets, and a
// writer that sees every failing request is the worst possible place to leak.
// p.Detail is safe to log because it is, by construction, already being sent to
// the client.

// Respond is the canonical terminal emission point for an HTTP error.
//
// It writes the RFC 9457 application/problem+json representation of p exactly
// once (media type, status, body) and applies the logging policy above. r is
// required for request/trace correlation and is read-only; a nil r is tolerated
// (non-HTTP-scoped callers) at the cost of correlation. The caller must return
// immediately after Respond — it is terminal.
func Respond(w http.ResponseWriter, r *http.Request, p *Problem) {
	respond(w, r, p, nil)
}

// RespondCause behaves exactly like Respond and additionally records cause in
// the server-side log line. cause is NEVER serialized into the response: it is
// the internal diagnosis, p is the client contract.
func RespondCause(w http.ResponseWriter, r *http.Request, p *Problem, cause error) {
	respond(w, r, p, cause)
}

// fallbackBody is the pre-serialized payload used when p itself cannot be
// marshalled. It is a literal (not a marshal of a fallback Problem) so that the
// failure path cannot recurse into the failure it is handling.
const fallbackBody = `{"title":"Internal server error","status":500,"code":"internal.unknown"}`

func respond(w http.ResponseWriter, r *http.Request, p *Problem, cause error) {
	ctx := context.Background()
	method, path := "", ""
	if r != nil {
		ctx = r.Context()
		method = r.Method
		if r.URL != nil {
			path = r.URL.Path
		}
	}

	if p == nil {
		// A nil Problem is a caller defect, not a client error: never emit a
		// 200 with an empty body because of it.
		slog.ErrorContext(ctx, "problem: nil problem passed to writer",
			"trace_id", requesttrace.Resolve(ctx), "method", method, "path", path)
		p = NewFor(CodeInternalUnknown, "Internal server error")
	}

	body, marshalErr := json.Marshal(p)

	logProblem(ctx, p, cause, marshalErr, method, path)

	w.Header().Set("Content-Type", "application/problem+json")
	if marshalErr != nil {
		// Fail closed: the caller asked for an error response, so an error
		// response is what goes on the wire — never a bare 200.
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fallbackBody))
		return
	}
	w.WriteHeader(p.Status)
	_, _ = w.Write(body)
}

// logProblem applies the package logging policy documented above. It is the
// only place in MetalDocs that logs an outbound problem, so 4xx and 5xx cannot
// drift apart per call site.
func logProblem(ctx context.Context, p *Problem, cause, marshalErr error, method, path string) {
	attrs := []any{
		"trace_id", requesttrace.Resolve(ctx),
		"status", p.Status,
		"code", p.Code.String(),
		"title", p.Title,
		"method", method,
		"path", path,
	}
	if cause != nil {
		attrs = append(attrs, "err", cause.Error())
	}
	if marshalErr != nil {
		attrs = append(attrs, "marshal_err", marshalErr.Error())
		slog.ErrorContext(ctx, "problem: serialization failed, emitting fallback", attrs...)
		return
	}

	switch {
	case p.Status >= http.StatusInternalServerError:
		slog.ErrorContext(ctx, "http error response", attrs...)
	case p.Status >= http.StatusBadRequest:
		if cause != nil {
			slog.DebugContext(ctx, "http error response", attrs...)
		}
	default:
		slog.WarnContext(ctx, "problem: emitted with non-error status", attrs...)
	}
}
