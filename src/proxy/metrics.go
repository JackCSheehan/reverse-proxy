package proxy

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics for a particular endpoint
type EndpointMetrics struct {
	RequestCount           prometheus.Counter
	SuccessfulRequestCount prometheus.Counter
	FailedRequestCount     prometheus.Counter
	LastResponseTime       prometheus.Gauge
}

// Calls http.Handle to set up a /metrics endpoint for Prometheus metrics
func RegisterMetricsEndpoint() {
	http.Handle("/metrics", promhttp.Handler())
}

// Helper function to take a "from" resource string and convert it to a valid Prometheus metric
// prefix.
//
// For example, if the "from" resource string is "/home", this function will produce the prefix
// "home_".
func FromResourceToMetricsPrefix(from string) string {
	// Edge case for root path from strings
	if strings.TrimSpace(from) == "/" {
		return "ROOT_"
	}

	// Resources will more than likely have leading slashes, so cut it
	fromWithoutLeadingSlash := strings.TrimPrefix(from, "/")

	// Any other slashes we'll replace to make the resource paths human-readable in the metrics list
	fromWithoutSlashes := strings.Replace(fromWithoutLeadingSlash, "/", "_", -1)

	// Finaly, append an underscore since this string is intended to be used as a prefix for
	// metrics
	return fromWithoutSlashes + "_"
}

// Creates a new EndpointMetrics instance and registers the Prometheus metrics with the given
// registry. Accepts a Config instance which is used for something. prefix is prepended to
// metric names.
func NewEndpointMetrics(config *Config, prefix string) *EndpointMetrics {
	endpointMetrics := EndpointMetrics{
		RequestCount: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: prefix + "request_count",
			},
		),
		SuccessfulRequestCount: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: prefix + "successful_request_count",
				Help: "Total number of successful requests, not counting /metrics requests",
			},
		),
		FailedRequestCount: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: prefix + "failed_request_count",
				Help: "Total number of failed requests, not counting /metrics requests",
			},
		),
		LastResponseTime: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: prefix + "last_response_time",
				Help: "Response time of last request (ms)",
			},
		),
	}

	return &endpointMetrics
}
