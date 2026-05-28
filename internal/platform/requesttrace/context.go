package requesttrace

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type contextKey struct{}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if traceID, ok := Normalize(traceID); ok {
		return context.WithValue(ctx, contextKey{}, traceID)
	}
	return ctx
}

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

func Resolve(ctx context.Context) string {
	if traceID, ok := FromContext(ctx); ok {
		return traceID
	}
	return uuid.NewString()
}

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
