package ratelimit

import (
	"context"
	"math"
	"sync"
	"time"
)

// TokenBucket 令牌桶限流器
type TokenBucket struct {
	rate       float64 // tokens per second
	burst      int     // max tokens
	tokens     float64 // current tokens
	lastUpdate time.Time
	mu         sync.Mutex
}

// NewTokenBucket 创建令牌桶
func NewTokenBucket(rate float64, burst int) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		burst:      burst,
		tokens:     float64(burst),
		lastUpdate: time.Now(),
	}
}

// Wait 等待获取令牌
func (tb *TokenBucket) Wait(ctx context.Context) error {
	tb.mu.Lock()

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(tb.lastUpdate).Seconds()
	tb.tokens = math.Min(float64(tb.burst), tb.tokens+elapsed*tb.rate)
	tb.lastUpdate = now

	if tb.tokens >= 1 {
		tb.tokens--
		tb.mu.Unlock()
		return nil
	}

	// Calculate wait time
	waitDuration := time.Duration((1-tb.tokens)/tb.rate*1000) * time.Millisecond
	tb.mu.Unlock()

	select {
	case <-time.After(waitDuration):
		tb.mu.Lock()
		tb.tokens = 0
		tb.lastUpdate = time.Now()
		tb.mu.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TryAcquire 尝试获取令牌（非阻塞）
func (tb *TokenBucket) TryAcquire() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill
	now := time.Now()
	elapsed := now.Sub(tb.lastUpdate).Seconds()
	tb.tokens = math.Min(float64(tb.burst), tb.tokens+elapsed*tb.rate)
	tb.lastUpdate = now

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}
