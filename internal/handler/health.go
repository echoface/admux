package handler

import (
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	appCtx       interface{} // Will be *adxserver.AdxServerContext
	isHealthy    bool
	healthyMutex sync.RWMutex
	startTime    time.Time

	// Metrics
	healthCheckTotal     prometheus.Counter
	healthCheckDuration  prometheus.Histogram
	lastHealthCheckTime  prometheus.Gauge
}

// HealthStatus represents the health status response
type HealthStatus struct {
	Status     string                 `json:"status"`
	Timestamp  time.Time              `json:"timestamp"`
	Uptime     string                 `json:"uptime"`
	Version    string                 `json:"version,omitempty"`
	Components map[string]ComponentStatus `json:"components,omitempty"`
	Checks     map[string]bool        `json:"checks,omitempty"`
}

// ComponentStatus represents the status of a component
type ComponentStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	LastCheck time.Time `json:"last_check"`
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(appCtx interface{}, registry prometheus.Registerer) *HealthHandler {
	if registry == nil {
		registry = prometheus.DefaultRegisterer
	}

	hh := &HealthHandler{
		appCtx:    appCtx,
		isHealthy: true,
		startTime: time.Now(),
	}

	// Initialize metrics
	hh.healthCheckTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "health_check_requests_total",
			Help: "Total number of health check requests",
		},
	)

	hh.healthCheckDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "health_check_duration_seconds",
			Help:    "Duration of health checks",
			Buckets: prometheus.DefBuckets,
		},
	)

	hh.lastHealthCheckTime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "health_check_last_time_seconds",
			Help: "Unix timestamp of the last health check",
		},
	)

	// Register metrics
	registry.MustRegister(hh.healthCheckTotal)
	registry.MustRegister(hh.healthCheckDuration)
	registry.MustRegister(hh.lastHealthCheckTime)

	return hh
}

// HealthCheck handles the main health check endpoint
func (hh *HealthHandler) HealthCheck(c *gin.Context) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		hh.healthCheckDuration.Observe(duration)
		hh.lastHealthCheckTime.SetToCurrentTime()
		hh.healthCheckTotal.Inc()
	}()

	// Perform comprehensive health checks
	healthy := hh.performHealthChecks()

	status := "healthy"
	httpStatus := http.StatusOK
	if !healthy {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	response := HealthStatus{
		Status:    status,
		Timestamp: time.Now(),
		Uptime:    time.Since(hh.startTime).String(),
		Components: hh.getComponentStatus(),
		Checks: map[string]bool{
			"database":     hh.checkDatabase(),
			"redis":        hh.checkRedis(),
			"external_apis": hh.checkExternalAPIs(),
			"memory":       hh.checkMemory(),
		},
	}

	// Add version info if available
	if version, err := hh.getVersion(); err == nil {
		response.Version = version
	}

	c.JSON(httpStatus, response)
}

// LivenessProbe handles Kubernetes liveness probe
func (hh *HealthHandler) LivenessProbe(c *gin.Context) {
	hh.healthyMutex.RLock()
	alive := hh.isHealthy
	hh.healthyMutex.RUnlock()

	if alive {
		c.JSON(http.StatusOK, gin.H{
			"status": "alive",
			"timestamp": time.Now(),
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "dead",
			"timestamp": time.Now(),
		})
	}
}

// ReadinessProbe handles Kubernetes readiness probe
func (hh *HealthHandler) ReadinessProbe(c *gin.Context) {
	ready := hh.checkReadiness()

	status := http.StatusOK
	responseStatus := "ready"

	if !ready {
		status = http.StatusServiceUnavailable
		responseStatus = "not_ready"
	}

	c.JSON(status, gin.H{
		"status":    responseStatus,
		"timestamp": time.Now(),
		"checks": map[string]bool{
			"ssps_ready":   hh.checkSSPsReady(),
			"bidders_ready": hh.checkBiddersReady(),
			"config_loaded": hh.checkConfigLoaded(),
		},
	})
}

// SetHealthyStatus sets the overall health status
func (hh *HealthHandler) SetHealthyStatus(healthy bool) {
	hh.healthyMutex.Lock()
	defer hh.healthyMutex.Unlock()
	hh.isHealthy = healthy
}

// IsHealthy returns the current health status
func (hh *HealthHandler) IsHealthy() bool {
	hh.healthyMutex.RLock()
	defer hh.healthyMutex.RUnlock()
	return hh.isHealthy
}

// performHealthChecks performs all health checks
func (hh *HealthHandler) performHealthChecks() bool {
	// Basic health check
	hh.healthyMutex.RLock()
	basicHealthy := hh.isHealthy
	hh.healthyMutex.RUnlock()

	if !basicHealthy {
		return false
	}

	// Check critical components
	checks := map[string]bool{
		"database":      hh.checkDatabase(),
		"redis":         hh.checkRedis(),
		"memory":        hh.checkMemory(),
	}

	// If any critical check fails, the service is unhealthy
	for _, check := range checks {
		if !check {
			return false
		}
	}

	return true
}

// getComponentStatus returns the status of all components
func (hh *HealthHandler) getComponentStatus() map[string]ComponentStatus {
	return map[string]ComponentStatus{
		"database": {
			Status:     hh.getStatusString(hh.checkDatabase()),
			LastCheck:  time.Now(),
		},
		"redis": {
			Status:     hh.getStatusString(hh.checkRedis()),
			LastCheck:  time.Now(),
		},
		"external_apis": {
			Status:     hh.getStatusString(hh.checkExternalAPIs()),
			LastCheck:  time.Now(),
		},
		"memory": {
			Status:     hh.getStatusString(hh.checkMemory()),
			LastCheck:  time.Now(),
		},
	}
}

// Individual health check methods

func (hh *HealthHandler) checkDatabase() bool {
	// TODO: Implement actual database health check
	// For now, return true as we don't have a database yet
	return true
}

func (hh *HealthHandler) checkRedis() bool {
	// TODO: Implement actual Redis health check
	// For now, return true as Redis check will be implemented later
	return true
}

func (hh *HealthHandler) checkExternalAPIs() bool {
	// TODO: Implement external API health checks
	// This would check if SSP and DSP endpoints are reachable
	return true
}

func (hh *HealthHandler) checkMemory() bool {
	// Simple memory check - ensure we're not under extreme memory pressure
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Check if we're using more than 90% of allocated memory
	if m.Alloc > m.Sys*9/10 {
		return false
	}

	return true
}

func (hh *HealthHandler) checkReadiness() bool {
	// Check if the application is ready to serve traffic
	return hh.checkSSPsReady() && hh.checkBiddersReady() && hh.checkConfigLoaded()
}

func (hh *HealthHandler) checkSSPsReady() bool {
	// TODO: Check if SSP adapters are properly configured and ready
	// For now, return true
	return true
}

func (hh *HealthHandler) checkBiddersReady() bool {
	// TODO: Check if DSP bidders are properly configured and ready
	// For now, return true
	return true
}

func (hh *HealthHandler) checkConfigLoaded() bool {
	// TODO: Check if configuration is properly loaded
	// For now, return true
	return true
}

func (hh *HealthHandler) getStatusString(healthy bool) string {
	if healthy {
		return "healthy"
	}
	return "unhealthy"
}

func (hh *HealthHandler) getVersion() (string, error) {
	// TODO: Return actual version information
	// This could be set during build time using ldflags
	return "v1.0.0-dev", nil
}