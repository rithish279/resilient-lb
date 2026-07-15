package lb

import (
	"testing"
	"time"
)

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)

	for i := 0; i < 3; i++ {
		if !cb.Allow() {
			t.Fatalf("expected Allow() to be true before threshold reached, iteration %d", i)
		}
		cb.Failure()
	}

	if cb.State() != "open" {
		t.Fatalf("expected state 'open' after %d failures, got %s", 3, cb.State())
	}

	if cb.Allow() {
		t.Fatal("expected Allow() to be false when circuit is open")
	}
}

func TestCircuitBreaker_ClosesOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)

	cb.Failure()
	cb.Failure()
	cb.Success() // resets failure count

	if cb.State() != "closed" {
		t.Fatalf("expected state 'closed' after Success(), got %s", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenAfterTimeout(t *testing.T) {
	cb := NewCircuitBreaker(1, 50*time.Millisecond)

	cb.Failure() // opens after 1 failure
	if cb.State() != "open" {
		t.Fatalf("expected state 'open', got %s", cb.State())
	}

	time.Sleep(60 * time.Millisecond) // wait past timeout

	if !cb.Allow() {
		t.Fatal("expected Allow() to be true after timeout (half-open probe)")
	}

	if cb.State() != "half-open" {
		t.Fatalf("expected state 'half-open', got %s", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenFailureReopens(t *testing.T) {
	cb := NewCircuitBreaker(1, 50*time.Millisecond)

	cb.Failure()
	time.Sleep(60 * time.Millisecond)
	cb.Allow() // transitions to half-open

	cb.Failure() // probe fails

	if cb.State() != "open" {
		t.Fatalf("expected state 'open' after half-open probe fails, got %s", cb.State())
	}
}