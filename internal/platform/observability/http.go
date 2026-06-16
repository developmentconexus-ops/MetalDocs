package observability

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/requesttrace"

	"go.opentelemetry.io/otel/trace"
)

type routeMetrics struct {
	requests   uint64
	errors     uint64
	durationMs uint64
	mu         sync.Mutex
	samples    []uint64
	cursor     int
}

// SchedulerMetricsProvider is satisfied by any scheduler whose per-job counters
// should appear in the /api/v1/metrics payload. Defined here so the platform
// package stays free of module imports (REQ-TOP-2); the jobs package implements
// it implicitly via duck typing.
type SchedulerMetricsProvider interface {
	SchedulerMetrics() map[string]any
}

// DBPoolStatsProvider exposes database connection pool stats for the metrics endpoint.
type DBPoolStatsProvider interface {
	DBPoolStats() map[string]any
}

type HTTPObservability struct {
	runtimeProvider  RuntimeStatusProvider
	schedulerMetrics SchedulerMetricsProvider
	dbPool           DBPoolStatsProvider
	userIDResolver   func(*http.Request) string
	mu               sync.RWMutex
	byKey            map[string]*routeMetrics
}

type metricItem struct {
	Route           string `json:"route"`
	Method          string `json:"method"`
	Requests        uint64 `json:"requests"`
	Errors          uint64 `json:"errors"`
	DurationTotalMs uint64 `json:"duration_total_ms"`
	AvgDurationMs   uint64 `json:"avg_duration_ms"`
	P50DurationMs   uint64 `json:"p50_duration_ms"`
	P95DurationMs   uint64 `json:"p95_duration_ms"`
	P99DurationMs   uint64 `json:"p99_duration_ms"`
}

// NewHTTPObservability builds the metrics/logging middleware. userIDResolver
// is REQUIRED (not an optional setter) so the composition root cannot silently
// omit it and degrade every request to "anonymous" in the access log — an
// audit-attribution loss. The callback is how this platform package stays free
// of module imports (REQ-TOP-2); a nil resolver, or one returning empty, logs
// "anonymous". It is invoked per request and must not be swapped after the
// middleware sees traffic (the field is not synchronized).
func NewHTTPObservability(userIDResolver func(*http.Request) string, runtimeProvider ...RuntimeStatusProvider) *HTTPObservability {
	var provider RuntimeStatusProvider
	if len(runtimeProvider) > 0 {
		provider = runtimeProvider[0]
	}
	return &HTTPObservability{
		runtimeProvider: provider,
		userIDResolver:  userIDResolver,
		byKey:           make(map[string]*routeMetrics),
	}
}

// SetSchedulerMetrics wires a scheduler whose counters will appear as a top-level
// "scheduler" key in MetricsHandler's response. Call once during startup, before
// ListenAndServe — same set-once pattern as runtimeProvider.
func (o *HTTPObservability) SetSchedulerMetrics(s SchedulerMetricsProvider) {
	o.schedulerMetrics = s
}

// SetDBPool wires a DB pool stats provider. Call once during startup, before
// ListenAndServe. Absent → "db_pool" key omitted from MetricsHandler response.
func (o *HTTPObservability) SetDBPool(p DBPoolStatsProvider) {
	o.dbPool = p
}

func (o *HTTPObservability) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Trace-id resolution order (REQ-OBS-2): an active OTel span (created by
		// the otelhttp link when OTel is configured, seeded from an inbound W3C
		// traceparent if present) wins, so logs/metrics correlate with exported
		// spans. Otherwise fall back to the X-Trace-Id header, then requesttrace.
		// When OTel is off SpanContext is invalid and behaviour is unchanged.
		traceID := ""
		if sc := trace.SpanContextFromContext(r.Context()); sc.IsValid() {
			traceID = sc.TraceID().String()
		}
		if traceID == "" {
			var ok bool
			traceID, ok = requesttrace.Normalize(r.Header.Get("X-Trace-Id"))
			if !ok {
				traceID = requesttrace.Resolve(r.Context())
			}
		}
		r = r.WithContext(withPrincipalSlot(requesttrace.WithTraceID(r.Context(), traceID)))

		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		panicked := true
		defer func() {
			if panicked {
				// A downstream panic is unwinding through us. The outer
				// recovery middleware writes the 500 problem+json; record
				// the request here so panics stay visible in RED metrics
				// (REQ-MW-1). The panic itself is NOT swallowed.
				sw.status = http.StatusInternalServerError
			}

			path := strings.ReplaceAll(r.URL.Path, "\n", "")
			route := routeLabel(r, path)
			method := r.Method
			elapsedMs := time.Since(start).Milliseconds()
			if elapsedMs < 0 {
				elapsedMs = 0
			}
			durationMs := uint64(elapsedMs)
			isError := sw.status >= 400

			m := o.getMetric(route, method)
			atomic.AddUint64(&m.requests, 1)
			if isError {
				atomic.AddUint64(&m.errors, 1)
			}
			atomic.AddUint64(&m.durationMs, durationMs)
			m.record(durationMs)

			// Attribution order: principal slot (set outward by authn when
			// this middleware runs outside auth, REQ-MW-4) → injected
			// resolver (works when wrapped inside auth, e.g. tests).
			userID := "anonymous"
			if id := principalFromContext(r.Context()); id != "" {
				userID = id
			} else if o.userIDResolver != nil {
				if id := strings.TrimSpace(o.userIDResolver(r)); id != "" {
					userID = id
				}
			}
			documentID, profileCode := extractRouteContext(path)

			slog.Info("http_request",
				"trace_id", traceID,
				"user_id", userID,
				"method", method,
				"path", path,
				"route", route,
				"status", sw.status,
				"duration_ms", durationMs,
				"document_id", documentID,
				"profile_code", profileCode,
			)
		}()
		next.ServeHTTP(sw, r)
		panicked = false
	})
}

func (o *HTTPObservability) MetricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			httpresponse.WriteMethodNotAllowed(w, "GET")
			return
		}

		items := o.snapshot()
		payload := map[string]any{"items": items}
		if o.runtimeProvider != nil {
			payload["runtime"] = o.runtimeProvider.RuntimeMetrics(r.Context())
		}
		if o.schedulerMetrics != nil {
			payload["scheduler"] = o.schedulerMetrics.SchedulerMetrics()
		}
		if o.dbPool != nil {
			payload["db_pool"] = o.dbPool.DBPoolStats()
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	})
}

func (o *HTTPObservability) snapshot() []metricItem {
	o.mu.RLock()
	defer o.mu.RUnlock()

	out := make([]metricItem, 0, len(o.byKey))
	for key, m := range o.byKey {
		method, route := splitKey(key)
		req := atomic.LoadUint64(&m.requests)
		errs := atomic.LoadUint64(&m.errors)
		dur := atomic.LoadUint64(&m.durationMs)
		avg := uint64(0)
		if req > 0 {
			avg = dur / req
		}
		p50, p95, p99 := m.percentiles()
		out = append(out, metricItem{
			Route:           route,
			Method:          method,
			Requests:        req,
			Errors:          errs,
			DurationTotalMs: dur,
			AvgDurationMs:   avg,
			P50DurationMs:   p50,
			P95DurationMs:   p95,
			P99DurationMs:   p99,
		})
	}
	return out
}

func (o *HTTPObservability) getMetric(route, method string) *routeMetrics {
	key := method + " " + route
	// Take the read lock first so hot-path lookups do not queue behind writers.
	// We release it before Lock to avoid starving concurrent readers while a new
	// route bucket is being created.
	o.mu.RLock()
	m, ok := o.byKey[key]
	o.mu.RUnlock()
	if ok {
		return m
	}

	o.mu.Lock()
	defer o.mu.Unlock()
	if existing, ok := o.byKey[key]; ok {
		return existing
	}
	created := &routeMetrics{samples: make([]uint64, 0, 200)}
	o.byKey[key] = created
	return created
}

// routeLabel returns the low-cardinality route key for metrics/logs. Go 1.22's
// ServeMux sets r.Pattern to the matched method+pattern (e.g.
// "GET /api/v1/iam/users/{user_id}/roles") after dispatch; this defer runs after
// next.ServeHTTP, so when the request reached the mux unmodified the pattern is
// available and collapses every path variable for free (F-17 cardinality fix).
// When an intermediate middleware replaced the request (so Pattern did not
// propagate back) it is empty, and we fall back to the hand-maintained
// normalizeRoute — identical to the pre-OTel behaviour.
func routeLabel(r *http.Request, path string) string {
	if r.Pattern != "" {
		if _, p, ok := strings.Cut(r.Pattern, " "); ok && p != "" {
			return p
		}
		return r.Pattern
	}
	return normalizeRoute(path)
}

func normalizeRoute(path string) string {
	if strings.HasPrefix(path, "/api/v1/documents/") {
		parts := strings.Split(path, "/")
		if len(parts) > 4 {
			parts[4] = "{documentId}"
			if len(parts) > 5 && parts[5] == "versions" {
				return "/api/v1/documents/{documentId}/versions"
			}
			return strings.Join(parts, "/")
		}
	}
	if strings.HasPrefix(path, "/api/v1/document-profiles/") {
		parts := strings.Split(path, "/")
		if len(parts) > 4 {
			parts[4] = "{profileCode}"
			return strings.Join(parts, "/")
		}
	}
	if strings.HasPrefix(path, "/api/v1/workflow/documents/") && strings.HasSuffix(path, "/transitions") {
		return "/api/v1/workflow/documents/{documentId}/transitions"
	}
	if strings.HasPrefix(path, "/api/v1/iam/users/") && strings.HasSuffix(path, "/roles") {
		return "/api/v1/iam/users/{user_id}/roles"
	}
	if strings.HasPrefix(path, "/api/v1/iam/users/") && strings.HasSuffix(path, "/reset-password") {
		return "/api/v1/iam/users/{user_id}/reset-password"
	}
	if strings.HasPrefix(path, "/api/v1/iam/users/") && strings.HasSuffix(path, "/unlock") {
		return "/api/v1/iam/users/{user_id}/unlock"
	}
	return path
}

func splitKey(key string) (method, route string) {
	parts := strings.SplitN(key, " ", 2)
	if len(parts) != 2 {
		return "UNKNOWN", key
	}
	return parts[0], parts[1]
}

func extractRouteContext(path string) (documentID string, profileCode string) {
	if strings.HasPrefix(path, "/api/v1/documents/") {
		parts := strings.Split(path, "/")
		if len(parts) > 4 {
			documentID = parts[4]
		}
	}
	if strings.HasPrefix(path, "/api/v1/document-profiles/") {
		parts := strings.Split(path, "/")
		if len(parts) > 4 {
			profileCode = parts[4]
		}
	}
	return documentID, profileCode
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		// Header() is called before Write; the result here is always the set value.
		if v := w.Header().Get("Status"); v != "" {
			if code, err := strconv.Atoi(v); err == nil {
				w.status = code
			}
		}
		if w.status == 0 {
			w.status = http.StatusOK
		}
	}
	return w.ResponseWriter.Write(b)
}

func (w *statusWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Unwrap exposes the wrapped ResponseWriter so http.ResponseController and
// libraries that follow the Unwrap chain (e.g. coder/websocket's hijacker)
// can reach the underlying http.Hijacker/Flusher. Without this, WebSocket
// upgrades through the observability middleware fail with 501 (A1).
func (w *statusWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (m *routeMetrics) record(durationMs uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.samples) < cap(m.samples) {
		m.samples = append(m.samples, durationMs)
		return
	}
	m.samples[m.cursor] = durationMs
	m.cursor = (m.cursor + 1) % len(m.samples)
}

func (m *routeMetrics) percentiles() (uint64, uint64, uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.samples) == 0 {
		return 0, 0, 0
	}
	cpy := make([]uint64, len(m.samples))
	copy(cpy, m.samples)
	sort.Slice(cpy, func(i, j int) bool { return cpy[i] < cpy[j] })
	return percentile(cpy, 0.50), percentile(cpy, 0.95), percentile(cpy, 0.99)
}

func percentile(values []uint64, p float64) uint64 {
	if len(values) == 0 {
		return 0
	}
	if p <= 0 {
		return values[0]
	}
	if p >= 1 {
		return values[len(values)-1]
	}
	idx := int(float64(len(values)-1) * p)
	return values[idx]
}
