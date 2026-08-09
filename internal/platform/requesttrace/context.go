package requesttrace

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type contextKey struct{}

// WithTraceID returns a copy of ctx carrying traceID, provided traceID
// normalizes to a valid value; otherwise ctx is returned unchanged.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if traceID, ok := Normalize(traceID); ok {
		return context.WithValue(ctx, contextKey{}, traceID)
	}
	return ctx
}

// FromContext returns the normalized trace ID stored in ctx by WithTraceID,
// and whether one was present.
func FromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	traceID, ok := ctx.Value(contextKey{}).(string)
	if !ok {
		return "", false
	}
	traceID, ok = Normalize(traceID)
	return traceID, ok
}

// Resolve returns the trace ID carried by ctx, or generates a fresh UUID when
// none is present.
func Resolve(ctx context.Context) string {
	if traceID, ok := FromContext(ctx); ok {
		return traceID
	}
	return uuid.NewString()
}

// Normalize trims raw and validates it as a trace ID: non-empty, at most 128
// bytes, and printable ASCII only. Returns ("", false) when raw fails validation.
func Normalize(raw string) (string, bool) {
	traceID := strings.TrimSpace(raw)
	if traceID == "" || len(traceID) > 128 {
		return "", false
	}
	for _, r := range traceID {
		if r < 0x21 || r > 0x7e {
			return "", false
		}
	}
	return traceID, true
}
