package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// BroadcastMetrics 广播相关指标
type BroadcastMetrics struct {
	// 请求相关指标
	TotalRequests prometheus.Counter
	SuccessResponses prometheus.Counter
	FailedResponses  prometheus.Counter

	// 延迟相关指标
	ResponseLatency prometheus.Histogram

	// 并发相关指标
	ActiveBidders prometheus.Gauge
	ConcurrentRequests prometheus.Gauge

	// 重试相关指标
	RetryCount prometheus.Counter

	// 健康状态相关指标
	HealthyBidders prometheus.Gauge
	UnhealthyBidders prometheus.Gauge

	// 熔断器相关指标
	CircuitBreakerOpen prometheus.Gauge
	CircuitBreakerHalfOpen prometheus.Gauge
	CircuitBreakerClosed prometheus.Gauge
}

// NewBroadcastMetrics 创建广播指标实例
func NewBroadcastMetrics(namespace, subsystem string) *BroadcastMetrics {
	return &BroadcastMetrics{
		TotalRequests: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "total_requests",
			Help:      "Total number of broadcast requests",
		}),
		SuccessResponses: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "success_responses",
			Help:      "Number of successful bid responses",
		}),
		FailedResponses: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "failed_responses",
			Help:      "Number of failed bid responses",
		}),
		ResponseLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "response_latency_seconds",
			Help:      "Response latency in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 2.0},
		}),
		ActiveBidders: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "active_bidders",
			Help:      "Number of active bidders",
		}),
		ConcurrentRequests: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "concurrent_requests",
			Help:      "Number of concurrent broadcast requests",
		}),
		RetryCount: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "retry_count",
			Help:      "Total number of retries",
		}),
		HealthyBidders: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "healthy_bidders",
			Help:      "Number of healthy bidders",
		}),
		UnhealthyBidders: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "unhealthy_bidders",
			Help:      "Number of unhealthy bidders",
		}),
		CircuitBreakerOpen: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "circuit_breaker_open",
			Help:      "Circuit breaker open state",
		}),
		CircuitBreakerHalfOpen: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "circuit_breaker_half_open",
			Help:      "Circuit breaker half-open state",
		}),
		CircuitBreakerClosed: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "circuit_breaker_closed",
			Help:      "Circuit breaker closed state",
		}),
	}
}

// RecordRequest 记录请求指标
func (m *BroadcastMetrics) RecordRequest() {
	m.TotalRequests.Inc()
	m.ConcurrentRequests.Inc()
}

// RecordResponse 记录响应指标
func (m *BroadcastMetrics) RecordResponse(success bool, latency float64) {
	m.ConcurrentRequests.Dec()
	m.ResponseLatency.Observe(latency)

	if success {
		m.SuccessResponses.Inc()
	} else {
		m.FailedResponses.Inc()
	}
}

// RecordRetry 记录重试指标
func (m *BroadcastMetrics) RecordRetry() {
	m.RetryCount.Inc()
}

// UpdateBidderHealth 更新bidder健康指标
func (m *BroadcastMetrics) UpdateBidderHealth(healthyCount, unhealthyCount int) {
	m.HealthyBidders.Set(float64(healthyCount))
	m.UnhealthyBidders.Set(float64(unhealthyCount))
}

// UpdateCircuitBreakerState 更新熔断器状态指标
func (m *BroadcastMetrics) UpdateCircuitBreakerState(open, halfOpen, closed bool) {
	if open {
		m.CircuitBreakerOpen.Set(1)
		m.CircuitBreakerHalfOpen.Set(0)
		m.CircuitBreakerClosed.Set(0)
	} else if halfOpen {
		m.CircuitBreakerOpen.Set(0)
		m.CircuitBreakerHalfOpen.Set(1)
		m.CircuitBreakerClosed.Set(0)
	} else if closed {
		m.CircuitBreakerOpen.Set(0)
		m.CircuitBreakerHalfOpen.Set(0)
		m.CircuitBreakerClosed.Set(1)
	}
}