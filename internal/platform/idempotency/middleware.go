package idempotency

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
)

var uuidRE = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// Require returns middleware that enforces the Idempotency-Key header.
// actorFromCtx extracts (tenantID, actorID) from the request context.
// On a replay hit, the stored response is written and the handler is not called.
// On conflict (same key, different body hash), 422 is returned.
// On miss, the handler is called and a 2xx response is recorded.
func Require(store *Store, actorFromCtx func(context.Context) (string, string)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				writeErrJSON(w, 400, "IDEMPOTENCY_KEY_REQUIRED", "Idempotency-Key header required")
				return
			}
			if !uuidRE.MatchString(key) {
				writeErrJSON(w, 400, "IDEMPOTENCY_KEY_INVALID", "Idempotency-Key must be a UUID")
				return
			}

			tenantID, actorID := actorFromCtx(r.Context())

			buf, err := io.ReadAll(r.Body)
			if err != nil {
				writeErrJSON(w, 400, "BAD_REQUEST", "cannot read body")
				return
			}
			_ = r.Body.Close()
			r.Body = io.NopCloser(bytes.NewReader(buf))

			sum := sha256.Sum256(buf)
			hash := hex.EncodeToString(sum[:])

			replay, err := store.CheckReplay(r.Context(), tenantID, actorID, key, hash)
			if errors.Is(err, ErrConflict) {
				writeErrJSON(w, 422, "IDEMPOTENCY_KEY_CONFLICT", "key reused with different payload")
				return
			}
			if err != nil {
				writeErrJSON(w, 500, "INTERNAL", "idempotency check failed")
				return
			}
			if replay != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(replay.Status)
				_, _ = w.Write(replay.Body)
				return
			}

			rec := &responseRecorder{ResponseWriter: w}
			next.ServeHTTP(rec, r)

			if rec.status >= 200 && rec.status < 300 {
				_ = store.RecordReplay(r.Context(), tenantID, actorID, key, hash, rec.status, rec.body.Bytes())
			}
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

func writeErrJSON(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"code": code, "message": msg})
}
