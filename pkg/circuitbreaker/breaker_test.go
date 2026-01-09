package circuitbreaker

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	cb := New("test", 5, 60*time.Second)
	
	if cb == nil {
		t.Fatal("New returned nil")
	}
	
	if cb.GetState() != StateClosed {
		t.Errorf("Initial state should be Closed, got %v", cb.GetState())
	}
}

func TestBreaker_AllowRequest(t *testing.T) {
	cb := New("test", 5, 60*time.Second)
	
	// Should allow request when closed
	if !cb.AllowRequest() {
		t.Error("AllowRequest() should return true for closed circuit")
	}
}

func TestBreaker_RecordSuccess(t *testing.T) {
	cb := New("test", 3, 1*time.Second)
	
	// Record success
	cb.RecordSuccess()
	
	if cb.GetState() != StateClosed {
		t.Errorf("State should remain Closed after success, got %v", cb.GetState())
	}
}

func TestBreaker_RecordFailure(t *testing.T) {
	cb := New("test", 3, 1*time.Second)
	
	// First 2 failures should keep circuit closed
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.GetState() != StateClosed {
		t.Errorf("State should be Closed after 2 failures, got %v", cb.GetState())
	}
	
	// 3rd failure should open circuit
	cb.RecordFailure()
	if cb.GetState() != StateOpen {
		t.Errorf("State should be Open after threshold failures, got %v", cb.GetState())
	}
}

func TestBreaker_OpenState(t *testing.T) {
	cb := New("test", 2, 100*time.Millisecond)
	
	// Trigger circuit to open
	cb.RecordFailure()
	cb.RecordFailure()
	
	if cb.GetState() != StateOpen {
		t.Fatal("Circuit should be open")
	}
	
	// Requests should be blocked
	if cb.AllowRequest() {
		t.Error("AllowRequest() should return false when circuit is open")
	}
	
	// After timeout, should transition to half-open
	time.Sleep(150 * time.Millisecond)
	if !cb.AllowRequest() {
		t.Error("AllowRequest() should return true after timeout (half-open)")
	}
	
	if cb.GetState() != StateHalfOpen {
		t.Errorf("State should be HalfOpen after timeout, got %v", cb.GetState())
	}
}

func TestBreaker_IsHealthy(t *testing.T) {
	cb := New("test", 2, 1*time.Second)
	
	// Should be healthy when closed
	if err := cb.IsHealthy(); err != nil {
		t.Errorf("IsHealthy() should return nil when closed: %v", err)
	}
	
	// Open circuit
	cb.RecordFailure()
	cb.RecordFailure()
	
	// Should be unhealthy when open
	if err := cb.IsHealthy(); err == nil {
		t.Error("IsHealthy() should return error when open")
	}
}

func TestState_String(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{StateClosed, "CLOSED"},
		{StateOpen, "OPEN"},
		{StateHalfOpen, "HALF_OPEN"},
	}
	
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Errorf("State.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
