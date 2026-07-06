package observability

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// promState holds the Prometheus registry and the request/error/duration
// vectors instrumented directly in the Wrap hot path (REQ-OBS: /metrics is a
// platform concern, no module imports). It is built lazily via
// prometheusState() so callers that never scrape /metrics pay no extra cost
// beyond the vec increments already happening in Wrap.
type promState struct {
	registry        *prometheus.Registry
	requestsTotal   *prometheus.CounterVec
	errorsTotal     *prometheus.CounterVec
	durationSeconds *prometheus.HistogramVec
}

// dbPoolCollector is a custom prometheus.Collector that calls DBPoolStats()
// at scrape time and emits one gauge per numeric stat. Using a Collector
// (rather than pre-registered gauges) means an unwired or nil provider simply
// yields no series, and pool composition can change without a registration
// dance.
type dbPoolCollector struct {
	// providerFn is resolved at scrape time (not registration time) so a
	// SetDBPool call that happens after prometheusState() has already lazily
	// initialized (e.g. a scrape landing before Postgres wiring finishes)
	// still picks up the provider on the next scrape. Avoids holding o.mu
	// during Collect by taking the snapshot read outside the lock.
	providerFn func() DBPoolStatsProvider
}

var _ prometheus.Collector = (*dbPoolCollector)(nil)

func (c *dbPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	// Dynamic metric set: intentionally unchecked collector (valid pattern
	// for value-only descriptors whose label/name set is data-driven).
}

func (c *dbPoolCollector) Collect(ch chan<- prometheus.Metric) {
	provider := c.providerFn()
	if provider == nil {
		return
	}
	stats := provider.DBPoolStats()
	for name, v := range stats {
		f, ok := toFloat64(v)
		if !ok {
			continue
		}
		desc := prometheus.NewDesc(
			"metaldocs_db_pool_"+name,
			"DB connection pool stat: "+name,
			nil, nil,
		)
		ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, f)
	}
}

func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	default:
		return 0, false
	}
}

// prometheusState lazily builds and caches the Prometheus registry + vecs on
// first use (either the first Wrap-recorded request or the first
// PrometheusHandler() call, whichever comes first). o.mu (already used to
// guard byKey) also guards this one-time init.
func (o *HTTPObservability) prometheusState() *promState {
	o.mu.RLock()
	if o.prom != nil {
		p := o.prom
		o.mu.RUnlock()
		return p
	}
	o.mu.RUnlock()

	o.mu.Lock()
	defer o.mu.Unlock()
	if o.prom != nil {
		return o.prom
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	// o.dbPool follows the same set-once-before-traffic pattern as
	// runtimeProvider/schedulerMetrics (see SetDBPool doc comment); read
	// directly here, matching MetricsHandler's unguarded read.
	reg.MustRegister(&dbPoolCollector{providerFn: func() DBPoolStatsProvider {
		return o.dbPool
	}})

	requestsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "metaldocs_http_requests_total",
		Help: "Total number of HTTP requests handled, labeled by route and method.",
	}, []string{"route", "method"})
	errorsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "metaldocs_http_request_errors_total",
		Help: "Total number of HTTP requests that resulted in a >=400 status, labeled by route and method.",
	}, []string{"route", "method"})
	durationSeconds := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "metaldocs_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds, labeled by route and method.",
		Buckets: prometheus.DefBuckets,
	}, []string{"route", "method"})
	reg.MustRegister(requestsTotal, errorsTotal, durationSeconds)

	o.prom = &promState{
		registry:        reg,
		requestsTotal:   requestsTotal,
		errorsTotal:     errorsTotal,
		durationSeconds: durationSeconds,
	}
	return o.prom
}

// PrometheusHandler serves the Prometheus text-exposition format at whatever
// path the caller mounts it. main.go serves it from a top-level dispatch mux
// ahead of the entire API middleware chain (GET /metrics), NOT on the chained
// API mux — so a credential-less Prometheus scraper can reach it and the
// scrape itself is never counted by httpObs.Wrap or throttled by the global
// rate limiter. Contract §3.2: /metrics is not a versioned product route and
// is therefore NOT declared in api/openapi/v1/openapi.yaml. Uses a dedicated
// registry (not the global default) so this package stays composable and
// test-isolated.
func (o *HTTPObservability) PrometheusHandler() http.Handler {
	p := o.prometheusState()
	return promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{})
}
