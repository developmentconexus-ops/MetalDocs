package idempotency

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"regexp"
)

var uuidRE = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

const maxIdempotencyRequestBodyBytes int64 = 1 << 20

// IsValidKey reports whether an idempotency key matches the required UUID shape.
func IsValidKey(key string) bool {
	return uuidRE.MatchString(key)
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
			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				writeErrJSON(w, 400, "IDEMPOTENCY_KEY_REQUIRED", "Idempotency-Key header required")
				return
			}
			if !IsValidKey(key) {
				writeErrJSON(w, 400, "IDEMPOTENCY_KEY_INVALID", "Idempotency-Key must be a UUID")
				return
			}

			tenantID, actorID := actorFromCtx(r.Context())

			r.Body = http.MaxBytesReader(w, r.Body, maxIdempotencyRequestBodyBytes)
			hash, err := RequestHash(r)
			if err != nil {
				var maxBytesErr *http.MaxBytesError
				if errors.As(err, &maxBytesErr) {
					writeErrJSON(w, http.StatusRequestEntityTooLarge, "REQUEST_BODY_TOO_LARGE", "request body exceeds 1 MiB limit")
					return
				}
				writeErrJSON(w, 400, "BAD_REQUEST", "cannot read body")
				return
			}

			handle, replay, err := store.BeginReplay(r.Context(), tenantID, actorID, key, hash)
			if errors.Is(err, ErrConflict) {
				writeErrJSON(w, 422, "IDEMPOTENCY_KEY_CONFLICT", "key reused with different payload")
				return
			}
			if err != nil {
				slog.ErrorContext(r.Context(), "idempotency: begin failed",
					"key", key, "tenant", tenantID, "err", err)
				writeErrJSON(w, 500, "INTERNAL", "idempotency check failed")
				return
			}
			if replay != nil {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Idempotent-Replay", "true")
				w.WriteHeader(replay.Status)
				_, _ = w.Write(replay.Body)
				return
			}

			// Winner path: handle owns the slot. The deferred FailReplay
			// releases the slot on panic or non-2xx; CompleteReplay below
			// commits on 2xx and flips `released` so the defer is a no-op.
			released := false
			defer func() {
				if !released {
					if rerr := store.FailReplay(handle, nil); rerr != nil {
						slog.ErrorContext(r.Context(), "idempotency: fail-release error",
							"key", key, "tenant", tenantID, "err", rerr)
					}
				}
			}()

			rec := &responseRecorder{ResponseWriter: w}
			func() {
				defer func() {
					if p := recover(); p != nil {
						slog.ErrorContext(r.Context(), "idempotency: handler panicked",
							"key", key, "tenant", tenantID, "panic", p)
						panic(p)
					}
				}()
				next.ServeHTTP(rec, r)
			}()

			if rec.status >= 200 && rec.status < 300 {
				if err := store.CompleteReplay(handle, rec.status, rec.body.Bytes()); err != nil {
					// Response already shipped; surface the persistence loss
					// loudly so alerting can catch silent idempotency collapse
					// (H1).
					slog.ErrorContext(r.Context(), "idempotency: record failed — retry may duplicate",
						"key", key, "tenant", tenantID, "status", rec.status, "err", err)
				}
				released = true
				return
			}
			// Non-2xx: deliberately fall through so the deferred FailReplay
			// releases the slot.
		})
	}
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

func writeErrJSON(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"code": code, "message": msg})
}
