package lb

import (
	"sync"
	"time"

)

type cbState int

/*
cbClosed   -> normal
cbOpen     -> reject requests (fail)
cbHalfOpen -> testing (backend recovery)
*/
const (
	cbClosed cbState = iota
	cbOpen
	cbHalfOpen
)

type CircuitBreaker struct {
	mu 			sync.Mutex
	state 		cbState
	failures	int
	thresold	int
	timeout		time.Duration
	lastFailure	time.Time
}

func NewCircuitBreaker(thresold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		thresold: thresold,
		timeout: timeout,
	}
}


func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case cbClosed:
		return true
	case cbOpen:
		if (time.Since(cb.lastFailure) >= cb.timeout) {
			cb.state = cbHalfOpen
			return true
		}
		return false
	case cbHalfOpen:
		return true
	}
	return false
}

// Records a successful request
func (cb *CircuitBreaker) Success() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.state = cbClosed
}

// Records a failed request
func (cb *CircuitBreaker) Failure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	if (cb.failures >= cb.thresold) {
		cb.state = cbOpen
	}
}

func (cb *CircuitBreaker) State() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case cbOpen:
		return "open"
	case cbHalfOpen:
		return "half-open"
	default:
		return "closed"
	}
}