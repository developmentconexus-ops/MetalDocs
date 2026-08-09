package idempotency

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"regexp"

	"metaldocs/internal/platform/problem"
)

var uuidRE = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

const maxIdempotencyRequestBodyBytes int64 = 1 << 20

// Idempotency-Key header contract sentinels. The wire rule is UUID-everywhere
// (F-QA4-6): the same shape is enforced by Require below AND by the handlers
// that manage their own replay slots instead of being wrapped by it. Handlers
// map these to the same 400 problem+json Require emits — see ValidateKey.
var (
	// ErrKeyRequired signals a missing (or blank) Idempotency-Key header.
	ErrKeyRequired = errors.New("idempotency: Idempotency-Key header required")
	// ErrKeyInvalid signals a present but non-UUID Idempotency-Key header.
	ErrKeyInvalid = errors.New("idempotency: Idempotency-Key must be a UUID")
)

// IsValidKey reports whether an idempotency key matches the required UUID shape.
func IsValidKey(key string) bool {
	return uuidRE.MatchString(key)
}

// ValidateKey is the single Idempotency-Key format rule for the whole API.
// Both the Require middleware and the bespoke-replay handlers (approval
// signoff / fast-forward / review-verdict / route-admin / submit) call it, so
// there is exactly one regex and no handler-specific dialect.
//
// It validates only — it does not trim. Callers that accept surrounding
// whitespace must normalise before calling, which keeps middleware parity
// exact (Require passes the raw header value).
func ValidateKey(key string) error {
	if key == "" {
		return ErrKeyRequired
	}
	if !IsValidKey(key) {
		return ErrKeyInvalid
	}
	return nil
}

// RequestHash computes a deterministic hash of method + path + query + body.
// This prevents replaying a key across different resource paths with identical bodies.
func RequestHash(r *http.Request) (string, error) {
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(buf))

	sum := sha256.New()
	sum.Write([]byte(r.Method))
	sum.Write([]byte{'\n'})
	sum.Write([]byte(r.URL.Path))
	sum.Write([]byte{'?'})
	sum.Write([]byte(r.URL.RawQuery))
	sum.Write([]byte{'\n'})
	sum.Write(buf)
	return hex.EncodeToString(sum.Sum(nil)), nil
}

// Option configures Require.
type Option func(*config)

type config struct {
	streamingOptOut func(*http.Request) bool
}

// WithStreamingOptOut skips idempotency wrapping when matcher returns true.
//
// Idempotency buffers the full response for replay, which is incompatible
// with streaming responses (SSE, chunked, long-lived). Routes that need
// streaming must opt out — there is no way to both buffer-for-replay and
// stream incrementally. Opted-out routes get NO idempotency protection;
// callers must not send Idempotency-Key for them, or accept that retries
// can duplicate work.
func WithStreamingOptOut(matcher func(*http.Request) bool) Option {
	return func(c *config) { c.streamingOptOut = matcher }
}

// Require returns middleware that enforces the Idempotency-Key header using
// the two-phase BeginReplay / CompleteReplay / FailReplay protocol.
//
// On a replay hit, the stored response is written and the handler is not called.
// On conflict (same key, different body hash), 422 is returned.
// On a winning claim, the handler runs inside the lifetime of the claim's
// transaction; CompleteReplay records the response on a 2xx, FailReplay
// releases the slot on a panic or non-2xx so the next retry can re-execute.
//
// Streaming contract: the response is buffered for replay storage, so
// wrapped handlers MUST NOT use http.Flusher. Any call to Flush() on the
// wrapped writer panics with a clear directive to use WithStreamingOptOut.
// This fails closed rather than silently buffering or panicking with an
// opaque interface-conversion error.
func Require(store *Store, actorFromCtx func(context.Context) (string, string), opts ...Option) func(http.Handler) http.Handler {
	cfg := config{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.streamingOptOut != nil && cfg.streamingOptOut(r) {
				next.ServeHTTP(w, r)
				return
			}
			serveWithIdempotency(store, actorFromCtx, next, w, r)
		})
	}
}

// serveWithIdempotency runs the two-phase BeginReplay / CompleteReplay /
// FailReplay protocol described on Require, once streaming opt-out has
// already been ruled out by the caller.
func serveWithIdempotency(store *Store, actorFromCtx func(context.Context) (string, string), next http.Handler, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	key := r.Header.Get("Idempotency-Key")
	switch err := ValidateKey(key); {
	case errors.Is(err, ErrKeyRequired):
		writeErrJSON(w, 400, problem.CodeRequestIdempotencyKeyRequired, "Idempotency-Key header required")
		return
	case errors.Is(err, ErrKeyInvalid):
		writeErrJSON(w, 400, problem.CodeRequestIdempotencyKeyInvalid, "Idempotency-Key must be a UUID")
		return
	}

	tenantID, actorID := actorFromCtx(ctx)

	r.Body = http.MaxBytesReader(w, r.Body, maxIdempotencyRequestBodyBytes)
	hash, err := RequestHash(r)
	if err != nil {
		writeRequestHashError(w, err)
		return
	}

	handle, replay, err := store.BeginReplay(ctx, tenantID, actorID, key, hash)
	if errors.Is(err, ErrConflict) {
		// 409, not the pre-ADR-0089 422: the status is bound to
		// conflict.idempotency_key_reused at registration (annex R-8).
		writeErrJSON(w, http.StatusConflict, problem.CodeConflictIdempotencyKeyReused, "key reused with different payload")
		return
	}
	if err != nil {
		slog.ErrorContext(ctx, "idempotency: begin failed",
			"key", key, "tenant", tenantID, "err", err)
		writeErrJSON(w, 500, problem.CodeInternalUnknown, "idempotency check failed")
		return
	}
	if replay != nil {
		writeReplay(w, replay)
		return
	}

	runClaimedHandler(ctx, store, handle, tenantID, key, next, w, r)
}

// writeRequestHashError maps a RequestHash failure to its wire response: the
// MaxBytesReader-triggered 413 case, or a generic 400 for any other read error.
func writeRequestHashError(w http.ResponseWriter, err error) {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		writeErrJSON(w, http.StatusRequestEntityTooLarge, problem.CodeRequestBodyTooLarge, "request body exceeds 1 MiB limit")
		return
	}
	writeErrJSON(w, 400, problem.CodeRequestInvalid, "cannot read body")
}

// writeReplay writes a cached response on an idempotency replay hit.
func writeReplay(w http.ResponseWriter, replay *Replay) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Idempotent-Replay", "true")
	w.WriteHeader(replay.Status)
	_, _ = w.Write(replay.Body) //nolint:gosec // G705: replay.Body is a response OUR OWN handlers produced earlier (cached verbatim), served as application/json — not reflectable markup
}

// runClaimedHandler runs next for the caller that won the BeginReplay claim
// (handle owns the slot). The deferred FailReplay releases the slot on panic
// or non-2xx; CompleteReplay commits on 2xx and flips `released` so the
// defer is a no-op.
func runClaimedHandler(ctx context.Context, store *Store, handle *ReplayHandle, tenantID, key string, next http.Handler, w http.ResponseWriter, r *http.Request) {
	released := false
	defer func() {
		if !released {
			if rerr := store.FailReplay(handle, nil); rerr != nil {
				slog.ErrorContext(ctx, "idempotency: fail-release error",
					"key", key, "tenant", tenantID, "err", rerr)
			}
		}
	}()

	rec := &responseRecorder{ResponseWriter: w}
	invokeHandlerWithPanicLogging(ctx, next, rec, r, tenantID, key)

	if rec.status >= 200 && rec.status < 300 {
		if err := store.CompleteReplay(handle, rec.status, rec.body.Bytes()); err != nil {
			// Response already shipped; surface the persistence loss
			// loudly so alerting can catch silent idempotency collapse
			// (H1).
			slog.ErrorContext(ctx, "idempotency: record failed — retry may duplicate",
				"key", key, "tenant", tenantID, "status", rec.status, "err", err)
		}
		released = true
		return
	}
	// Non-2xx: deliberately fall through so the deferred FailReplay
	// releases the slot.
}

// invokeHandlerWithPanicLogging runs next, logging (and re-panicking) any
// handler panic before it unwinds through the idempotency slot bookkeeping
// in runClaimedHandler.
func invokeHandlerWithPanicLogging(ctx context.Context, next http.Handler, rec *responseRecorder, r *http.Request, tenantID, key string) {
	defer func() {
		if p := recover(); p != nil {
			slog.ErrorContext(ctx, "idempotency: handler panicked",
				"key", key, "tenant", tenantID, "panic", p)
			panic(p)
		}
	}()
	next.ServeHTTP(rec, r)
}

type responseRecorder struct {
	http.ResponseWriter
	status      int
	body        bytes.Buffer
	wroteHeader bool
}

func (r *responseRecorder) WriteHeader(code int) {
	if r.wroteHeader {
		return
	}
	r.wroteHeader = true
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(200)
	}
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// Flush fails closed. Idempotency buffers the entire response body for
// replay storage, which is fundamentally incompatible with incremental
// streaming. Rather than silently buffering (breaking SSE) or letting an
// opaque "interface conversion" panic surface, we panic with an explicit
// directive: streaming routes must use WithStreamingOptOut to skip
// idempotency wrapping entirely.
func (r *responseRecorder) Flush() {
	panic("idempotency: handler called Flush() on a wrapped ResponseWriter; " +
		"streaming responses are incompatible with idempotency replay buffering. " +
		"Use idempotency.WithStreamingOptOut(matcher) to skip wrapping for this route.")
}

// writeErrJSON emits an RFC 9457 Problem (application/problem+json) — the single
// API error envelope (AD-2). The code parameter is typed problem.Code so call
// sites must use catalog constants; the codes_catalog_guard_test enforces that
// no raw string literal appears in the code position (F-09, REQ-API-5).
func writeErrJSON(w http.ResponseWriter, status int, code problem.Code, msg string) {
	if err := problem.Write(w, problem.New(status, code, msg)); err != nil {
		slog.Warn("idempotency: problem write failed", "err", err, "status", status, "code", code)
	}
}
