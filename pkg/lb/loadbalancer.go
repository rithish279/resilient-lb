package lb

import (
	"net/http"
	"sync/atomic"
)

type LoadBalancer struct {
	backends []*Backend
	counter  atomic.Uint32
}

func New(backends []*Backend) *LoadBalancer {
	return &LoadBalancer{backends: backends}
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	idx := lb.counter.Add(1) % uint32(len(lb.backends))
	backend := lb.backends[idx]

	if !backend.Healthy.Load() {
		for i := 0; i < len(lb.backends); i++ {
			idx = (idx + 1) % uint32(len(lb.backends))
			if lb.backends[idx].Healthy.Load() {
				backend = lb.backends[idx]
				break
			}
		}
	}

	if !backend.Healthy.Load() {
		http.Error(w, "no healthy backends available", http.StatusServiceUnavailable)
		return
	}

	backend.proxy.ServeHTTP(w, r)
}