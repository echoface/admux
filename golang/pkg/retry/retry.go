package retry

import (
	"context"
	"math"
	"time"
)

// ErrorType 错误类型
type ErrorType int

const (
	TimeoutError ErrorType = iota // 超时错误
	NetworkError                  // 网络错误
	ProtocolError                 // 协议错误
	RateLimitError                // 限流错误
	InternalError                 // 内部错误
)

// RetryableError 可重试错误
type RetryableError struct {
	Type    ErrorType
	Message string
}

func (e *RetryableError) Error() string {
	return e.Message
}

// IsRetryable 检查错误是否可重试
func (e *RetryableError) IsRetryable() bool {
	return e.Type == TimeoutError || e.Type == NetworkError || e.Type == RateLimitError
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries      int           // 最大重试次数
	InitialDelay    time.Duration // 初始延迟
	MaxDelay        time.Duration // 最大延迟
	BackoffMultiplier float64     // 退避乘数
}

// DefaultRetryConfig 默认重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:       2,
		InitialDelay:     100 * time.Millisecond,
		MaxDelay:         1 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

// RetryableFunc 可重试函数类型
type RetryableFunc[T any] func(ctx context.Context) (T, error)

// Retry 执行带重试的操作
func Retry[T any](ctx context.Context, fn RetryableFunc[T], config *RetryConfig) (T, error) {
	var zero T
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		result, err := fn(ctx)

		if err == nil {
			return result, nil
		}

		// 检查错误是否可重试
		if !isRetryableError(err) {
			return zero, err
		}

		lastErr = err

		// 如果不是最后一次尝试，则等待
		if attempt < config.MaxRetries {
			delay := calculateBackoffDelay(attempt, config)
			select {
			case <-time.After(delay):
				// 继续下一次尝试
			case <-ctx.Done():
				return zero, ctx.Err()
			}
		}
	}

	return zero, lastErr
}

// isRetryableError 检查错误是否可重试
func isRetryableError(err error) bool {
	if retryableErr, ok := err.(*RetryableError); ok {
		return retryableErr.IsRetryable()
	}

	// 默认情况下，网络和超时错误可重试
	// 可以根据具体错误类型进行更精细的判断
	return true
}

// calculateBackoffDelay 计算退避延迟
func calculateBackoffDelay(attempt int, config *RetryConfig) time.Duration {
	delay := time.Duration(float64(config.InitialDelay) * math.Pow(config.BackoffMultiplier, float64(attempt)))

	if delay > config.MaxDelay {
		return config.MaxDelay
	}

	return delay
}

// WithExponentialBackoff 带指数退避的重试包装器
func WithExponentialBackoff[T any](fn RetryableFunc[T], config *RetryConfig) RetryableFunc[T] {
	return func(ctx context.Context) (T, error) {
		return Retry(ctx, fn, config)
	}
}