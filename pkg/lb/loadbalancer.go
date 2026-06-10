package lb

import (
	"math"
	"net/http"
	"sync/atomic"
)

type Algorithm int

const (
	RoundRobin Algorithm = iota
	LeastConnections
	Weighted
)

type LoadBalancer struct {
	backends  []*Backend
	slots	  []*Backend	
	algorithm Algorithm
	counter   atomic.Uint32
}

func New(backends []*Backend, algorithm Algorithm) *LoadBalancer {
	lb := &LoadBalancer{
		backends: backends,
		algorithm: algorithm,
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

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := lb.pick()
	if backend == nil {
		http.Error(w, "no healthy backends available", http.StatusServiceUnavailable)
		return
	}
}
