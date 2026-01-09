package circuitbreaker

import (
	"fmt"
	"sync"
	"time"
)

// State 熔断器状态
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// String 状态字符串
func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// Breaker 熔断器
type Breaker struct {
	name             string
	failureThreshold int
	timeout          time.Duration

	state       State
	failures    int
	successes   int
	lastFailure time.Time
	mu          sync.RWMutex

	onStateChange func(from, to State)
}

// New 创建熔断器
func New(name string, failureThreshold int, timeout time.Duration) *Breaker {
	return &Breaker{
		name:             name,
		failureThreshold: failureThreshold,
		timeout:          timeout,
		state:            StateClosed,
	}
}

// AllowRequest 是否允许请求
func (cb *Breaker) AllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.transitionTo(StateHalfOpen)
			return true
		}
		return false
	case StateHalfOpen:
		return true
	}
	return false
}

// RecordSuccess 记录成功
func (cb *Breaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	if cb.state == StateHalfOpen {
		cb.successes++
		if cb.successes >= 3 {
			cb.transitionTo(StateClosed)
		}
	}
}

// RecordFailure 记录失败
func (cb *Breaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	if cb.failures >= cb.failureThreshold {
		cb.transitionTo(StateOpen)
	}
}

// transitionTo 状态转换
func (cb *Breaker) transitionTo(newState State) {
	if cb.state != newState {
		oldState := cb.state
		cb.state = newState
		cb.failures = 0
		cb.successes = 0

		if cb.onStateChange != nil {
			go cb.onStateChange(oldState, newState)
		}
	}
}

// IsHealthy 是否健康
func (cb *Breaker) IsHealthy() error {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if cb.state == StateOpen {
		return fmt.Errorf("circuit breaker %s is open", cb.name)
	}
	return nil
}

// GetState 获取当前状态
func (cb *Breaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}
