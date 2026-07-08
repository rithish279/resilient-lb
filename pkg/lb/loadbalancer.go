package lb

import (
	"math"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/rithish279/resilient-lb/pkg/chaos"
)

type Algorithm int

const (
	RoundRobin Algorithm = iota
	LeastConnections
	Weighted
)

type LoadBalancer struct {
	backends  	[]*Backend
	slots	  	[]*Backend	
	algorithm 	Algorithm
	counter   	atomic.Uint32
	chaosEngine	*chaos.Engine
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func New(backends []*Backend, algorithm Algorithm, engine *chaos.Engine) *LoadBalancer {
	lb := &LoadBalancer{
		backends: backends,
		algorithm: algorithm,
		chaosEngine: engine,
	}

	for _, b := range backends {
		for i := 0; i < b.Weight; i++ {
			lb.slots = append(lb.slots, b)
		}
	}
	return lb
}

func (lb *LoadBalancer) pickRoundRobin() *Backend {
	start := lb.counter.Add(1) % uint32(len(lb.backends))
	for i := 0; i < len(lb.backends); i++ {
		idx := (start + uint32(i)) % uint32(len(lb.backends))
		if lb.backends[idx].Healthy.Load() {
			return lb.backends[idx]
		}
	}
	return nil
}

func (lb *LoadBalancer) pickLeastConnections() *Backend {
	var best *Backend
	min := int64(math.MaxInt64)

	for _, b := range lb.backends {
		if !b.Healthy.Load() {
			continue
		}
		if conns := b.ActiveConns(); conns < min {
			min = conns
			best = b
		}
	}
	return best
}

func (lb *LoadBalancer) pickWeighted() *Backend {
	if (len(lb.slots) == 0) {
		return lb.pickRoundRobin()
	}
	for i := 0; i < len(lb.slots); i++ {
		idx := lb.counter.Add(1) % uint32(len(lb.slots))
		if lb.slots[idx].Healthy.Load() {
			return lb.slots[idx]
		}
	}
	return nil
}

func (lb *LoadBalancer) pick() *Backend {
	switch lb.algorithm {
	case LeastConnections:
		return lb.pickLeastConnections()
	case Weighted:
		return lb.pickWeighted()
	default:
		return lb.pickRoundRobin()
	}
}

// applyCI applies the chaos config
func applyCI(w http.ResponseWriter, cfg *chaos.Config) bool {
	switch cfg.Type {
	case chaos.FailureLatency:
		time.Sleep(time.Duration(cfg.LatencyMs) * time.Millisecond)
		return false // forwarding but, delayed

	case chaos.FailureError:
		code := cfg.ErrorCode
		if (code == 0) {
			code = http.StatusServiceUnavailable
		}
		http.Error(w, "chaos: injected error", code)
		return true

	case chaos.FailureDrop: // then close connection without responding
		if hijacker, ok := w.(http.Hijacker); ok {
			conn, _, _ := hijacker.Hijack()
			if (conn != nil) {
				conn.Close()
			}
		}
		return true

	case chaos.FailureKillSwitch:
		http.Error(w, "chaos: backend unavailable", http.StatusServiceUnavailable)
		return true
	}
	return false
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := lb.pick()
	if backend == nil {
		http.Error(w, "no healthy backends available", http.StatusServiceUnavailable)
		return
	}

	if !backend.CircuitBreaker.Allow() {
		http.Error(w, "circuit breaker open", http.StatusServiceUnavailable)
		return
	}

	if (lb.chaosEngine != nil) {
		if cfg := lb.chaosEngine.ShouldInject(backend.URL.String()); cfg != nil {
			if handled := applyCI(w, cfg); handled {
				return
			}
		}
	}

	backend.IncrementConns()
	defer backend.DecrementConns()

	rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
	backend.proxy.ServeHTTP(rw, r)

	if (rw.statusCode >= 500) {
		backend.CircuitBreaker.Failure()
	} else {
		backend.CircuitBreaker.Success()
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
