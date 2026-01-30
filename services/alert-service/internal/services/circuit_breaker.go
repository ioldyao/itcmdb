package services

import (
	"errors"
	"sync"
	"time"
)

// CircuitState 断路器状态
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker 断路器实现
type CircuitBreaker struct {
	maxFailures  int
	resetTimeout time.Duration
	state        CircuitState
	failures     int
	lastFailTime time.Time
	mu           sync.RWMutex
}

// NewCircuitBreaker 创建断路器
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}

// Call 执行操作，如果断路器打开则返回错误
func (cb *CircuitBreaker) Call(fn func() error) error {
	// 第一步：检查断路器状态（持有锁）
	cb.mu.Lock()
	if cb.state == StateOpen {
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			// 从Open转换到HalfOpen
			cb.state = StateHalfOpen
			cb.failures = 0
		} else {
			cb.mu.Unlock()
			return errors.New("circuit breaker is open")
		}
	}
	currentState := cb.state
	cb.mu.Unlock()

	// 第二步：在锁外执行操作（避免阻塞其他goroutine）
	err := fn()

	// 第三步：根据执行结果更新状态（持有锁）
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailTime = time.Now()

		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		}
		return err
	}

	// 成功则重置
	if currentState == StateHalfOpen {
		cb.state = StateClosed
	}
	cb.failures = 0
	return nil
}

// GetState 获取当前状态
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset 重置断路器
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failures = 0
}
