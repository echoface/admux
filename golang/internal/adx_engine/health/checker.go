package health

import (
	"sync"
	"time"
)

// HealthStatus 健康状态
type HealthStatus struct {
	Healthy       bool
	FailureCount  int
	SuccessCount  int
	LastCheck     time.Time
	LastError     error
}

// HealthChecker 健康检查器
type HealthChecker struct {
	checkInterval    time.Duration
	failureThreshold int
	successThreshold int
	mu               sync.RWMutex
	status           map[string]*HealthStatus
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(checkInterval time.Duration, failureThreshold, successThreshold int) *HealthChecker {
	return &HealthChecker{
		checkInterval:    checkInterval,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		status:           make(map[string]*HealthStatus),
	}
}

// UpdateHealthStatus 更新健康状态
func (hc *HealthChecker) UpdateHealthStatus(id string, success bool, err error) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	status, exists := hc.status[id]
	if !exists {
		status = &HealthStatus{
			Healthy:   true,
			LastCheck: time.Now(),
		}
		hc.status[id] = status
	}

	status.LastCheck = time.Now()

	if success {
		status.SuccessCount++
		status.FailureCount = 0
		status.LastError = nil

		if status.SuccessCount >= hc.successThreshold {
			status.Healthy = true
		}
	} else {
		status.FailureCount++
		status.SuccessCount = 0
		status.LastError = err

		if status.FailureCount >= hc.failureThreshold {
			status.Healthy = false
		}
	}
}

// IsHealthy 检查是否健康
func (hc *HealthChecker) IsHealthy(id string) bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	status, exists := hc.status[id]
	if !exists {
		return true // 默认健康
	}

	return status.Healthy
}

// GetHealthStatus 获取健康状态
func (hc *HealthChecker) GetHealthStatus(id string) *HealthStatus {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	status, exists := hc.status[id]
	if !exists {
		return &HealthStatus{
			Healthy:   true,
			LastCheck: time.Now(),
		}
	}

	// 返回副本以避免外部修改
	return &HealthStatus{
		Healthy:      status.Healthy,
		FailureCount: status.FailureCount,
		SuccessCount: status.SuccessCount,
		LastCheck:    status.LastCheck,
		LastError:    status.LastError,
	}
}

// GetAllHealthStatus 获取所有健康状态
func (hc *HealthChecker) GetAllHealthStatus() map[string]*HealthStatus {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	// 返回副本以避免外部修改
	result := make(map[string]*HealthStatus)
	for id, status := range hc.status {
		result[id] = &HealthStatus{
			Healthy:      status.Healthy,
			FailureCount: status.FailureCount,
			SuccessCount: status.SuccessCount,
			LastCheck:    status.LastCheck,
			LastError:    status.LastError,
		}
	}

	return result
}

// CleanupStaleStatus 清理过期的状态
func (hc *HealthChecker) CleanupStaleStatus(staleDuration time.Duration) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	now := time.Now()
	for id, status := range hc.status {
		if now.Sub(status.LastCheck) > staleDuration {
			delete(hc.status, id)
		}
	}
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	state         CircuitState
	failureCount  int
	successCount  int
	lastStateTime time.Time
	mu            sync.RWMutex
}

// CircuitState 熔断器状态
type CircuitState int

const (
	StateClosed CircuitState = iota   // 关闭状态（正常）
	StateOpen                         // 打开状态（熔断）
	StateHalfOpen                     // 半开状态（试探）
)

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		state:         StateClosed,
		lastStateTime: time.Now(),
	}
}

// Allow 检查是否允许请求
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// 检查是否应该进入半开状态
		if time.Since(cb.lastStateTime) > 60*time.Second {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = StateHalfOpen
			cb.lastStateTime = time.Now()
			cb.mu.Unlock()
			cb.mu.RLock()
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess 记录成功
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		cb.successCount++
		cb.failureCount = 0
	case StateHalfOpen:
		cb.successCount++
		if cb.successCount >= 3 {
			cb.state = StateClosed
			cb.lastStateTime = time.Now()
			cb.successCount = 0
		}
	}
}

// RecordFailure 记录失败
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		cb.failureCount++
		cb.successCount = 0
		if cb.failureCount >= 5 {
			cb.state = StateOpen
			cb.lastStateTime = time.Now()
		}
	case StateHalfOpen:
		cb.state = StateOpen
		cb.lastStateTime = time.Now()
	}
}