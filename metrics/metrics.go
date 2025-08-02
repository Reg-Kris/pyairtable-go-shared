// Package metrics provides Prometheus metrics utilities and helpers
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Registry holds all metrics
type Registry struct {
	registry *prometheus.Registry
	
	// HTTP metrics
	HTTPRequestsTotal     *prometheus.CounterVec
	HTTPRequestDuration   *prometheus.HistogramVec
	HTTPRequestSize       *prometheus.HistogramVec
	HTTPResponseSize      *prometheus.HistogramVec
	
	// Database metrics
	DatabaseConnectionsActive *prometheus.GaugeVec
	DatabaseConnectionsIdle   *prometheus.GaugeVec
	DatabaseQueryDuration     *prometheus.HistogramVec
	DatabaseQueriesTotal      *prometheus.CounterVec
	
	// Cache metrics
	CacheOperationsTotal   *prometheus.CounterVec
	CacheOperationDuration *prometheus.HistogramVec
	CacheHitRatio          *prometheus.GaugeVec
	
	// Business metrics
	UsersTotal        *prometheus.GaugeVec
	WorkspacesTotal   *prometheus.GaugeVec
	RecordsTotal      *prometheus.GaugeVec
	APICallsTotal     *prometheus.CounterVec
	
	// System metrics
	CPUUsage    *prometheus.GaugeVec
	MemoryUsage *prometheus.GaugeVec
	GoroutineCount *prometheus.Gauge
}

// New creates a new metrics registry
func New(namespace string) *Registry {
	registry := prometheus.NewRegistry()
	
	r := &Registry{
		registry: registry,
		
		// HTTP metrics
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		
		HTTPRequestSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_size_bytes",
				Help:      "HTTP request size in bytes",
				Buckets:   prometheus.ExponentialBuckets(1024, 2, 10),
			},
			[]string{"method", "endpoint"},
		),
		
		HTTPResponseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_response_size_bytes",
				Help:      "HTTP response size in bytes",
				Buckets:   prometheus.ExponentialBuckets(1024, 2, 10),
			},
			[]string{"method", "endpoint"},
		),
		
		// Database metrics
		DatabaseConnectionsActive: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "database_connections_active",
				Help:      "Number of active database connections",
			},
			[]string{"database"},
		),
		
		DatabaseConnectionsIdle: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "database_connections_idle",
				Help:      "Number of idle database connections",
			},
			[]string{"database"},
		),
		
		DatabaseQueryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "database_query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"operation", "table"},
		),
		
		DatabaseQueriesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "database_queries_total",
				Help:      "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		),
		
		// Cache metrics
		CacheOperationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_operations_total",
				Help:      "Total number of cache operations",
			},
			[]string{"operation", "result"},
		),
		
		CacheOperationDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "cache_operation_duration_seconds",
				Help:      "Cache operation duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"operation"},
		),
		
		CacheHitRatio: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "cache_hit_ratio",
				Help:      "Cache hit ratio",
			},
			[]string{"cache_type"},
		),
		
		// Business metrics
		UsersTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "users_total",
				Help:      "Total number of users",
			},
			[]string{"status"},
		),
		
		WorkspacesTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "workspaces_total",
				Help:      "Total number of workspaces",
			},
			[]string{"status"},
		),
		
		RecordsTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "records_total",
				Help:      "Total number of records",
			},
			[]string{"workspace_id"},
		),
		
		APICallsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "api_calls_total",
				Help:      "Total number of API calls",
			},
			[]string{"endpoint", "method", "status"},
		),
		
		// System metrics
		CPUUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "cpu_usage_percent",
				Help:      "CPU usage percentage",
			},
			[]string{"core"},
		),
		
		MemoryUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "memory_usage_bytes",
				Help:      "Memory usage in bytes",
			},
			[]string{"type"},
		),
		
		GoroutineCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "goroutines_total",
				Help:      "Total number of goroutines",
			},
		),
	}
	
	// Register all metrics
	r.registerMetrics()
	
	return r
}

// registerMetrics registers all metrics with the registry
func (r *Registry) registerMetrics() {
	// HTTP metrics
	r.registry.MustRegister(r.HTTPRequestsTotal)
	r.registry.MustRegister(r.HTTPRequestDuration)
	r.registry.MustRegister(r.HTTPRequestSize)
	r.registry.MustRegister(r.HTTPResponseSize)
	
	// Database metrics
	r.registry.MustRegister(r.DatabaseConnectionsActive)
	r.registry.MustRegister(r.DatabaseConnectionsIdle)
	r.registry.MustRegister(r.DatabaseQueryDuration)
	r.registry.MustRegister(r.DatabaseQueriesTotal)
	
	// Cache metrics
	r.registry.MustRegister(r.CacheOperationsTotal)
	r.registry.MustRegister(r.CacheOperationDuration)
	r.registry.MustRegister(r.CacheHitRatio)
	
	// Business metrics
	r.registry.MustRegister(r.UsersTotal)
	r.registry.MustRegister(r.WorkspacesTotal)
	r.registry.MustRegister(r.RecordsTotal)
	r.registry.MustRegister(r.APICallsTotal)
	
	// System metrics
	r.registry.MustRegister(r.CPUUsage)
	r.registry.MustRegister(r.MemoryUsage)
	r.registry.MustRegister(r.GoroutineCount)
}

// Handler returns the Prometheus metrics handler
func (r *Registry) Handler() http.Handler {
	return promhttp.HandlerFor(r.registry, promhttp.HandlerOpts{})
}

// HTTP Metrics helpers

// RecordHTTPRequest records HTTP request metrics
func (r *Registry) RecordHTTPRequest(method, endpoint string, statusCode int, duration time.Duration, requestSize, responseSize int64) {
	status := strconv.Itoa(statusCode)
	
	r.HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	r.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
	r.HTTPRequestSize.WithLabelValues(method, endpoint).Observe(float64(requestSize))
	r.HTTPResponseSize.WithLabelValues(method, endpoint).Observe(float64(responseSize))
}

// Database Metrics helpers

// RecordDatabaseConnections records database connection metrics
func (r *Registry) RecordDatabaseConnections(database string, active, idle int) {
	r.DatabaseConnectionsActive.WithLabelValues(database).Set(float64(active))
	r.DatabaseConnectionsIdle.WithLabelValues(database).Set(float64(idle))
}

// RecordDatabaseQuery records database query metrics
func (r *Registry) RecordDatabaseQuery(operation, table, status string, duration time.Duration) {
	r.DatabaseQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
	r.DatabaseQueriesTotal.WithLabelValues(operation, table, status).Inc()
}

// Cache Metrics helpers

// RecordCacheOperation records cache operation metrics
func (r *Registry) RecordCacheOperation(operation, result string, duration time.Duration) {
	r.CacheOperationsTotal.WithLabelValues(operation, result).Inc()
	r.CacheOperationDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordCacheHitRatio records cache hit ratio
func (r *Registry) RecordCacheHitRatio(cacheType string, ratio float64) {
	r.CacheHitRatio.WithLabelValues(cacheType).Set(ratio)
}

// Business Metrics helpers

// RecordUserCount records user count metrics
func (r *Registry) RecordUserCount(status string, count int) {
	r.UsersTotal.WithLabelValues(status).Set(float64(count))
}

// RecordWorkspaceCount records workspace count metrics
func (r *Registry) RecordWorkspaceCount(status string, count int) {
	r.WorkspacesTotal.WithLabelValues(status).Set(float64(count))
}

// RecordRecordCount records record count metrics
func (r *Registry) RecordRecordCount(workspaceID string, count int) {
	r.RecordsTotal.WithLabelValues(workspaceID).Set(float64(count))
}

// RecordAPICall records API call metrics
func (r *Registry) RecordAPICall(endpoint, method, status string) {
	r.APICallsTotal.WithLabelValues(endpoint, method, status).Inc()
}

// System Metrics helpers

// RecordCPUUsage records CPU usage metrics
func (r *Registry) RecordCPUUsage(core string, usage float64) {
	r.CPUUsage.WithLabelValues(core).Set(usage)
}

// RecordMemoryUsage records memory usage metrics
func (r *Registry) RecordMemoryUsage(memType string, usage int64) {
	r.MemoryUsage.WithLabelValues(memType).Set(float64(usage))
}

// RecordGoroutineCount records goroutine count
func (r *Registry) RecordGoroutineCount(count int) {
	r.GoroutineCount.Set(float64(count))
}

// Middleware returns a Gin middleware for recording HTTP metrics
func (r *Registry) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Process request
		c.Next()
		
		// Record metrics
		duration := time.Since(start)
		method := c.Request.Method
		endpoint := c.FullPath()
		statusCode := c.Writer.Status()
		requestSize := c.Request.ContentLength
		responseSize := int64(c.Writer.Size())
		
		if requestSize < 0 {
			requestSize = 0
		}
		if responseSize < 0 {
			responseSize = 0
		}
		
		r.RecordHTTPRequest(method, endpoint, statusCode, duration, requestSize, responseSize)
	}
}

// Timer helps measure operation duration
type Timer struct {
	start time.Time
}

// NewTimer creates a new timer
func NewTimer() *Timer {
	return &Timer{start: time.Now()}
}

// Duration returns the elapsed duration
func (t *Timer) Duration() time.Duration {
	return time.Since(t.start)
}

// ObserveHistogram records the duration in a histogram
func (t *Timer) ObserveHistogram(histogram prometheus.Observer) {
	histogram.Observe(t.Duration().Seconds())
}

// Global registry instance
var defaultRegistry *Registry

// SetDefault sets the default metrics registry
func SetDefault(registry *Registry) {
	defaultRegistry = registry
}

// GetDefault returns the default metrics registry
func GetDefault() *Registry {
	return defaultRegistry
}