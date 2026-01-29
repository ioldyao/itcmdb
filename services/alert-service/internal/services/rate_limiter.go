package services

import (
	"sync"
	"time"
)

// RateLimiter 速率限制器（令牌桶算法）
type RateLimiter struct {
	rate       int           // 每秒生成的令牌数
	capacity   int           // 桶容量
	tokens     int           // 当前令牌数
	lastRefill time.Time     // 上次填充时间
	mu         sync.Mutex
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(rate, capacity int) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		capacity:   capacity,
		tokens:     capacity,
		lastRefill: time.Now(),
	}
}

// Allow 检查是否允许操作
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 填充令牌
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)
	tokensToAdd := int(elapsed.Seconds() * float64(rl.rate))

	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.capacity {
			rl.tokens = rl.capacity
		}
		rl.lastRefill = now
	}

	// 检查是否有可用令牌
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// GetTokens 获取当前令牌数
func (rl *RateLimiter) GetTokens() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.tokens
}
