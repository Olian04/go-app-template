// [[ when (modeIs "http") ]]
package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Option func(*registryOptions)

type registryOptions struct {
	namespace   string
	constLabels map[string]string
}

func WithNamespace(s string) Option {
	return func(o *registryOptions) { o.namespace = s }
}

// WithConstLabels sets static metric labels (Prometheus naming rules apply).
func WithConstLabels(m map[string]string) Option {
	return func(o *registryOptions) { o.constLabels = m }
}

// Registry owns the process metric set: Go/process/build-info collectors plus
// the request rate, error, and duration series ("RED") the HTTP middleware feeds.
type Registry struct {
	reg       *prometheus.Registry
	requests  *prometheus.CounterVec
	duration  *prometheus.HistogramVec
	inFlight  prometheus.Gauge
}

func NewRegistry(opts ...Option) (*Registry, error) {
	var o registryOptions
	for _, opt := range opts {
		opt(&o)
	}

	if o.namespace == "" {
		return nil, fmt.Errorf("metrics.NewRegistry: WithNamespace is required")
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewBuildInfoCollector(),
	)

	labels := prometheus.Labels(o.constLabels)

	// method+code labels make error rate a query over the same series as rate.
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   o.namespace,
		Name:        "http_requests_total",
		Help:        "Total HTTP requests handled, by method and response code.",
		ConstLabels: labels,
	}, []string{"method", "code"})

	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   o.namespace,
		Name:        "http_request_duration_seconds",
		Help:        "HTTP request latency in seconds, by method and response code.",
		Buckets:     prometheus.DefBuckets,
		ConstLabels: labels,
	}, []string{"method", "code"})

	inFlight := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   o.namespace,
		Name:        "http_requests_in_flight",
		Help:        "HTTP requests currently being served.",
		ConstLabels: labels,
	})

	reg.MustRegister(requests, duration, inFlight)
	return &Registry{reg: reg, requests: requests, duration: duration, inFlight: inFlight}, nil
}

// ObserveRequest records one completed request against the counter and histogram.
func (r *Registry) ObserveRequest(method string, code int, d time.Duration) {
	codeStr := strconv.Itoa(code)
	r.requests.WithLabelValues(method, codeStr).Inc()
	r.duration.WithLabelValues(method, codeStr).Observe(d.Seconds())
}

func (r *Registry) IncInFlight() {
	r.inFlight.Inc()
}

func (r *Registry) DecInFlight() {
	r.inFlight.Dec()
}

func (r *Registry) Handler() http.Handler {
	return promhttp.HandlerFor(r.reg, promhttp.HandlerOpts{Registry: r.reg})
}
