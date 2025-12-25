package server

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// httpMetrics holds Prometheus metrics for HTTP server monitoring.
type httpMetrics struct {
	requestsTotal     *prometheus.CounterVec
	requestDuration   *prometheus.HistogramVec
	requestSize       *prometheus.HistogramVec
	responseSize      *prometheus.HistogramVec
	activeConnections prometheus.Gauge
}

// newHTTPMetrics creates a new set of HTTP metrics registered with the provided registry.
// If constLabels is provided, these labels will be added to all metrics.
func newHTTPMetrics(registry *prometheus.Registry, constLabels map[string]string) *httpMetrics {
	factory := promauto.With(registry)

	return &httpMetrics{
		requestsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "http_server_requests_total",
				Help:        "Total number of HTTP requests received",
				ConstLabels: constLabels,
			},
			[]string{"method", "path", "status"},
		),
		requestDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "http_server_request_duration_seconds",
				Help:        "Time taken to process HTTP requests",
				Buckets:     prometheus.DefBuckets,
				ConstLabels: constLabels,
			},
			[]string{"method", "path", "status"},
		),
		requestSize: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "http_server_request_size_bytes",
				Help:        "Size of HTTP request bodies",
				Buckets:     prometheus.ExponentialBuckets(1, 10, 8), // 1B to 10MB
				ConstLabels: constLabels,
			},
			[]string{"method", "path"},
		),
		responseSize: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "http_server_response_size_bytes",
				Help:        "Size of HTTP response bodies",
				Buckets:     prometheus.ExponentialBuckets(1, 10, 8), // 1B to 10MB
				ConstLabels: constLabels,
			},
			[]string{"method", "path", "status"},
		),
		activeConnections: factory.NewGauge(
			prometheus.GaugeOpts{
				Name:        "http_server_active_connections",
				Help:        "Number of active HTTP connections",
				ConstLabels: constLabels,
			},
		),
	}
}

// metricsMiddleware returns a gin middleware for metrics collection.
func (m *httpMetrics) metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Increment active connections
		m.activeConnections.Inc()
		defer m.activeConnections.Dec()

		// Record request size
		if c.Request.ContentLength > 0 {
			m.requestSize.WithLabelValues(c.Request.Method, c.FullPath()).Observe(float64(c.Request.ContentLength))
		}

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get response info
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		path := c.FullPath()

		// Handle cases where path might be empty (e.g., 404s)
		if path == "" {
			path = c.Request.URL.Path
		}

		// Record metrics
		m.requestsTotal.WithLabelValues(method, path, status).Inc()
		m.requestDuration.WithLabelValues(method, path, status).Observe(duration)

		// Record response size
		responseSize := float64(c.Writer.Size())
		if responseSize > 0 {
			m.responseSize.WithLabelValues(method, path, status).Observe(responseSize)
		}
	}
}

// createMetricsMiddleware creates a metrics middleware function.
// This is used by the server to add metrics
func createMetricsMiddleware(
	registry *prometheus.Registry,
	constLabels map[string]string,
) gin.HandlerFunc {
	metrics := newHTTPMetrics(registry, constLabels)
	return metrics.metricsMiddleware()
}
