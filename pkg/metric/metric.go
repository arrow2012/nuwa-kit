package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// HTTPRequestsTotal counts the total number of HTTP requests.
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nuwa_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration measures the duration of HTTP requests.
	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "nuwa_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// BusinessOpsTotal counts specific business operations (e.g., login, create_user).
	BusinessOpsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nuwa_business_ops_total",
			Help: "Total number of business operations",
		},
		[]string{"type", "status"},
	)

	// ExternalAPIDuration measures latency of external dependencies
	ExternalAPIDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "nuwa_external_api_duration_seconds",
			Help:    "Duration of external API calls in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10},
		},
		[]string{"service", "operation", "status"},
	)

	// EventBusPublished stats
	EventBusPublished = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nuwa_event_published_total",
			Help: "Total number of events published",
		},
		[]string{"topic"},
	)

	// EventBusConsumed stats
	EventBusConsumed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nuwa_event_consumed_total",
			Help: "Total number of events consumed",
		},
		[]string{"topic", "status"},
	)
)

func init() {
	// Register metrics with Prometheus's default registry.
	prometheus.MustRegister(HTTPRequestsTotal)
	prometheus.MustRegister(HTTPRequestDuration)
	prometheus.MustRegister(BusinessOpsTotal)
	prometheus.MustRegister(ExternalAPIDuration)
	prometheus.MustRegister(EventBusPublished)
	prometheus.MustRegister(EventBusConsumed)
}
