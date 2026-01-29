package services

import (
	"fmt"
	"math"
	"time"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries     int           // 最大重试次数
	InitialBackoff time.Duration // 初始退避时间
	MaxBackoff     time.Duration // 最大退避时间
	Multiplier     float64       // 退避倍数
}

// DefaultRetryConfig 默认重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		Multiplier:     2.0,
	}
}

// RetryWithExponentialBackoff 使用指数退避重试
func RetryWithExponentialBackoff(config *RetryConfig, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			// 计算退避时间
			backoff := time.Duration(float64(config.InitialBackoff) * math.Pow(config.Multiplier, float64(attempt-1)))
			if backoff > config.MaxBackoff {
				backoff = config.MaxBackoff
			}
			time.Sleep(backoff)
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err
	}

	return fmt.Errorf("failed after %d retries: %w", config.MaxRetries, lastErr)
}
