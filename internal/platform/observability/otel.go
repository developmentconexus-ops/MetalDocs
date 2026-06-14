package observability

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// SetupOTel installs a real OpenTelemetry tracer provider ONLY when the operator
// configured an exporter (OTEL_TRACES_EXPORTER or OTEL_EXPORTER_OTLP_ENDPOINT
// set). Unconfigured ⇒ no SDK install, the global tracer stays the no-op default,
// and the otel chain link is skipped — zero overhead, byte-identical boot. This
// env-gated, inert-by-default behaviour is REQUIRED per Wave Z spec Z-1: the
// single-host default deployment (D-1) ships no collector, so OTel must cost
// nothing until a second host or an external-tenant SLA turns it on (REQ-OBS-3).
//
// The W3C TraceContext + Baggage composite propagator is installed so inbound
// `traceparent` headers are honoured and outbound calls carry the active trace
// (REQ-OBS-2). The existing X-Trace-Id path keeps working: httpObs bridges to the
// active span's TraceID when one is present (see Wrap), and falls back to
// requesttrace otherwise.
//
// serviceName sets the resource service.name attribute; if the OTEL_SERVICE_NAME
// environment variable is set (trimmed, non-empty) it overrides serviceName.
//
// Returns a shutdown func (always non-nil — a no-op when disabled) and whether
// the SDK was installed. autoexport reads OTEL_TRACES_EXPORTER (otlp|console|
// none); "console" writes spans to stdout.
func SetupOTel(ctx context.Context, serviceName string) (shutdown func(context.Context) error, enabled bool, err error) {
	noop := func(context.Context) error { return nil }

	// OTEL_TRACES_EXPORTER=none (the OTel-spec "explicitly off" value, matched
	// case-insensitively + trimmed) short-circuits before any SDK install —
	// return the no-op provider even when an OTLP endpoint is also set. The
	// empty-exporter + empty-endpoint case stays off too (unconfigured default).
	exporter := strings.ToLower(strings.TrimSpace(os.Getenv("OTEL_TRACES_EXPORTER")))
	if exporter == "none" {
		return noop, false, nil
	}
	if exporter == "" && os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		return noop, false, nil
	}

	exp, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return noop, false, err
	}

	if envName := strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME")); envName != "" {
		serviceName = envName
	}
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes("", attribute.String("service.name", serviceName)),
	)
	if err != nil && !errors.Is(err, resource.ErrSchemaURLConflict) {
		return noop, false, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, true, nil
}

// OTelMiddleware returns the chain link that wraps the handler in an
// otelhttp.Handler, naming each server span by a low-cardinality route. It is
// only invoked when SetupOTel reported enabled=true; when OTel is off, main.go
// passes nil for this link and buildChain skips it (zero overhead).
//
// Span name resolution mirrors httpObs's routeLabel: prefer r.Pattern (Go 1.22
// ServeMux), which otelhttp re-evaluates after the inner mux matches. Auth/iam
// middlewares between this link and the mux re-create the request via
// WithContext, so r.Pattern often does not propagate back to the instance
// otelhttp holds; in that case fall back to normalizeRoute(path) — the same
// hand-maintained collapse used for metrics — instead of the raw URL, keeping
// span cardinality bounded regardless (REQ-OBS-1).
func OTelMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, "http.server",
			otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
				if r.Pattern != "" {
					if _, p, ok := strings.Cut(r.Pattern, " "); ok && p != "" {
						return r.Method + " " + p
					}
					return r.Pattern
				}
				return r.Method + " " + normalizeRoute(strings.ReplaceAll(r.URL.Path, "\n", ""))
			}),
		)
	}
}
