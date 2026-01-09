package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestNewTokenBucket(t *testing.T) {
	tb := NewTokenBucket(10, 20)
	
	if tb == nil {
		t.Fatal("NewTokenBucket returned nil")
	}
}

func TestTokenBucket_Wait(t *testing.T) {
	tb := NewTokenBucket(100, 100) // High rate for testing
	
	// First request should succeed immediately
	ctx := context.Background()
	if err := tb.Wait(ctx); err != nil {
		t.Errorf("First Wait() failed: %v", err)
	}
}

func TestTokenBucket_TryAcquire(t *testing.T) {
	tb := NewTokenBucket(100, 10)
	
	// Should be able to acquire initially
	if !tb.TryAcquire() {
		t.Error("TryAcquire() failed on first attempt")
	}
	
	// After exhausting tokens, should eventually get more
	time.Sleep(100 * time.Millisecond)
	successCount := 0
	for i := 0; i < 5; i++ {
		if tb.TryAcquire() {
			successCount++
		}
	}
	
	if successCount == 0 {
		t.Error("No tokens refilled after waiting")
	}
}

func TestTokenBucket_WaitContextCancellation(t *testing.T) {
	tb := NewTokenBucket(0.1, 1) // Very low rate
	
	// Consume token
	ctx := context.Background()
	tb.Wait(ctx)
	
	// Create a cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	
	// Should return context error
	err := tb.Wait(cancelledCtx)
	if err == nil {
		t.Error("Wait() with cancelled context should return error")
	}
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}
