// Package health provides health check utilities and handlers for monitoring service health
package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Status represents the health status of a component
type Status string

const (
	StatusUp      Status = "up"
	StatusDown    Status = "down"
	StatusWarning Status = "warning"
)

// CheckResult represents the result of a health check
type CheckResult struct {
	Status    Status                 `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"`
}

// Check represents a health check function
type Check func(ctx context.Context) CheckResult

// Checker manages health checks for various components
type Checker struct {
	checks   map[string]Check
	mutex    sync.RWMutex
	timeout  time.Duration
	metadata map[string]interface{}
}

// NewChecker creates a new health checker
func NewChecker() *Checker {
	return &Checker{
		checks:   make(map[string]Check),
		timeout:  30 * time.Second,
		metadata: make(map[string]interface{}),
	}
}

// SetTimeout sets the timeout for health checks
func (hc *Checker) SetTimeout(timeout time.Duration) {
	hc.timeout = timeout
}

// SetMetadata sets metadata for the health checker
func (hc *Checker) SetMetadata(key string, value interface{}) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	hc.metadata[key] = value
}

// AddCheck adds a health check
func (hc *Checker) AddCheck(name string, check Check) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	hc.checks[name] = check
}

// RemoveCheck removes a health check
func (hc *Checker) RemoveCheck(name string) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	delete(hc.checks, name)
}

// CheckHealth performs all health checks
func (hc *Checker) CheckHealth(ctx context.Context) HealthResponse {
	hc.mutex.RLock()
	checks := make(map[string]Check)
	for name, check := range hc.checks {
		checks[name] = check
	}
	metadata := make(map[string]interface{})
	for key, value := range hc.metadata {
		metadata[key] = value
	}
	hc.mutex.RUnlock()

	results := make(map[string]CheckResult)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Run checks concurrently
	for name, check := range checks {
		wg.Add(1)
		go func(name string, check Check) {
			defer wg.Done()
			
			// Create context with timeout
			checkCtx, cancel := context.WithTimeout(ctx, hc.timeout)
			defer cancel()
			
			start := time.Now()
			result := check(checkCtx)
			result.Duration = time.Since(start)
			result.Timestamp = start
			
			mu.Lock()
			results[name] = result
			mu.Unlock()
		}(name, check)
	}

	wg.Wait()

	// Determine overall status
	overallStatus := StatusUp
	for _, result := range results {
		if result.Status == StatusDown {
			overallStatus = StatusDown
			break
		} else if result.Status == StatusWarning && overallStatus == StatusUp {
			overallStatus = StatusWarning
		}
	}

	return HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Checks:    results,
		Metadata:  metadata,
		System:    getSystemInfo(),
	}
}

// HealthResponse represents the overall health response
type HealthResponse struct {
	Status    Status                    `json:"status"`
	Timestamp time.Time                 `json:"timestamp"`
	Checks    map[string]CheckResult    `json:"checks"`
	Metadata  map[string]interface{}    `json:"metadata"`
	System    map[string]interface{}    `json:"system"`
}

// Handler returns a Gin handler for health checks
func (hc *Checker) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		response := hc.CheckHealth(c.Request.Context())
		
		statusCode := http.StatusOK
		switch response.Status {
		case StatusDown:
			statusCode = http.StatusServiceUnavailable
		case StatusWarning:
			statusCode = http.StatusOK // Warning still returns 200
		}
		
		c.JSON(statusCode, response)
	}
}

// SimpleHandler returns a simple health check handler
func SimpleHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "up",
			"timestamp": time.Now(),
		})
	}
}

// ReadinessHandler returns a readiness check handler
func (hc *Checker) ReadinessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		response := hc.CheckHealth(c.Request.Context())
		
		// For readiness, we're more strict - any down component means not ready
		statusCode := http.StatusOK
		if response.Status != StatusUp {
			statusCode = http.StatusServiceUnavailable
		}
		
		c.JSON(statusCode, response)
	}
}

// LivenessHandler returns a liveness check handler (simple uptime check)
func LivenessHandler() gin.HandlerFunc {
	startTime := time.Now()
	
	return func(c *gin.Context) {
		uptime := time.Since(startTime)
		
		c.JSON(http.StatusOK, gin.H{
			"status":    "up",
			"timestamp": time.Now(),
			"uptime":    uptime.String(),
		})
	}
}

// Common health check functions

// DatabaseCheck creates a database health check
func DatabaseCheck(db interface{ Health() error }) Check {
	return func(ctx context.Context) CheckResult {
		start := time.Now()
		
		if err := db.Health(); err != nil {
			return CheckResult{
				Status:    StatusDown,
				Message:   "Database connection failed",
				Details:   map[string]interface{}{"error": err.Error()},
				Timestamp: start,
			}
		}
		
		return CheckResult{
			Status:    StatusUp,
			Message:   "Database connection healthy",
			Timestamp: start,
		}
	}
}

// CacheCheck creates a cache health check
func CacheCheck(cache interface{ Health(context.Context) error }) Check {
	return func(ctx context.Context) CheckResult {
		start := time.Now()
		
		if err := cache.Health(ctx); err != nil {
			return CheckResult{
				Status:    StatusDown,
				Message:   "Cache connection failed",
				Details:   map[string]interface{}{"error": err.Error()},
				Timestamp: start,
			}
		}
		
		return CheckResult{
			Status:    StatusUp,
			Message:   "Cache connection healthy",
			Timestamp: start,
		}
	}
}

// HTTPCheck creates an HTTP endpoint health check
func HTTPCheck(url string, timeout time.Duration) Check {
	return func(ctx context.Context) CheckResult {
		start := time.Now()
		
		client := &http.Client{Timeout: timeout}
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return CheckResult{
				Status:    StatusDown,
				Message:   "Failed to create HTTP request",
				Details:   map[string]interface{}{"error": err.Error()},
				Timestamp: start,
			}
		}
		
		resp, err := client.Do(req)
		if err != nil {
			return CheckResult{
				Status:    StatusDown,
				Message:   "HTTP request failed",
				Details:   map[string]interface{}{"error": err.Error()},
				Timestamp: start,
			}
		}
		defer resp.Body.Close()
		
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return CheckResult{
				Status:    StatusUp,
				Message:   "HTTP endpoint healthy",
				Details:   map[string]interface{}{"status_code": resp.StatusCode},
				Timestamp: start,
			}
		}
		
		return CheckResult{
			Status:    StatusDown,
			Message:   "HTTP endpoint returned error status",
			Details:   map[string]interface{}{"status_code": resp.StatusCode},
			Timestamp: start,
		}
	}
}

// MemoryCheck creates a memory usage health check
func MemoryCheck(warningThreshold, criticalThreshold uint64) Check {
	return func(ctx context.Context) CheckResult {
		start := time.Now()
		
		var m runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m)
		
		allocated := m.Alloc
		
		status := StatusUp
		message := "Memory usage is normal"
		
		if allocated > criticalThreshold {
			status = StatusDown
			message = "Memory usage is critical"
		} else if allocated > warningThreshold {
			status = StatusWarning
			message = "Memory usage is high"
		}
		
		return CheckResult{
			Status:  status,
			Message: message,
			Details: map[string]interface{}{
				"allocated_bytes":    allocated,
				"warning_threshold":  warningThreshold,
				"critical_threshold": criticalThreshold,
			},
			Timestamp: start,
		}
	}
}

// GoroutineCheck creates a goroutine count health check
func GoroutineCheck(warningThreshold, criticalThreshold int) Check {
	return func(ctx context.Context) CheckResult {
		start := time.Now()
		
		count := runtime.NumGoroutine()
		
		status := StatusUp
		message := "Goroutine count is normal"
		
		if count > criticalThreshold {
			status = StatusDown
			message = "Goroutine count is critical"
		} else if count > warningThreshold {
			status = StatusWarning
			message = "Goroutine count is high"
		}
		
		return CheckResult{
			Status:  status,
			Message: message,
			Details: map[string]interface{}{
				"goroutine_count":    count,
				"warning_threshold":  warningThreshold,
				"critical_threshold": criticalThreshold,
			},
			Timestamp: start,
		}
	}
}

// getSystemInfo returns system information
func getSystemInfo() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]interface{}{
		"go_version":     runtime.Version(),
		"go_os":          runtime.GOOS,
		"go_arch":        runtime.GOARCH,
		"cpu_count":      runtime.NumCPU(),
		"goroutine_count": runtime.NumGoroutine(),
		"memory": map[string]interface{}{
			"allocated_bytes":   m.Alloc,
			"total_alloc_bytes": m.TotalAlloc,
			"sys_bytes":         m.Sys,
			"gc_count":          m.NumGC,
		},
	}
}

// JSONHandler returns the health status as JSON
func (hr *HealthResponse) JSONHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	statusCode := http.StatusOK
	switch hr.Status {
	case StatusDown:
		statusCode = http.StatusServiceUnavailable
	case StatusWarning:
		statusCode = http.StatusOK
	}
	
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(hr)
}

// String returns a string representation of the health status
func (hr *HealthResponse) String() string {
	data, _ := json.MarshalIndent(hr, "", "  ")
	return string(data)
}

// IsHealthy returns true if the overall status is up
func (hr *HealthResponse) IsHealthy() bool {
	return hr.Status == StatusUp
}

// HasWarnings returns true if any check has warnings
func (hr *HealthResponse) HasWarnings() bool {
	if hr.Status == StatusWarning {
		return true
	}
	
	for _, check := range hr.Checks {
		if check.Status == StatusWarning {
			return true
		}
	}
	
	return false
}

// FailedChecks returns a list of failed check names
func (hr *HealthResponse) FailedChecks() []string {
	var failed []string
	
	for name, check := range hr.Checks {
		if check.Status == StatusDown {
			failed = append(failed, name)
		}
	}
	
	return failed
}