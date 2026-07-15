package lb

import (
	"testing"
)

func newTestBackend(url string, weight int) *Backend {
	b := NewBackend(url, weight)
	b.Healthy.Store(true)
	return b
}

func TestRoundRobin_DistributesEvenly(t *testing.T) {
	backends := []*Backend{
		newTestBackend("http://backend1", 1),
		newTestBackend("http://backend2", 1),
	}
	lb := New(backends, RoundRobin, nil)

	counts := map[string]int{}
	for i := 0; i < 100; i++ {
		b := lb.pick()
		if b == nil {
			t.Fatal("pick() returned nil, expected a backend")
		}
		counts[b.URL.String()]++
	}

	for url, count := range counts {
		if count < 40 || count > 60 {
			t.Errorf("expected roughly even distribution, backend %s got %d/100", url, count)
		}
	}
}

func TestRoundRobin_SkipsUnhealthy(t *testing.T) {
	b1 := newTestBackend("http://backend1", 1)
	b2 := newTestBackend("http://backend2", 1)
	b2.Healthy.Store(false)

	lb := New([]*Backend{b1, b2}, RoundRobin, nil)

	for i := 0; i < 10; i++ {
		b := lb.pick()
		if b == nil {
			t.Fatal("pick() returned nil, expected healthy backend")
		}
		if b.URL.String() != "http://backend1" {
			t.Fatalf("expected only backend1 to be picked, got %s", b.URL.String())
		}
	}
}

func TestRoundRobin_ReturnsNilWhenAllUnhealthy(t *testing.T) {
	b1 := newTestBackend("http://backend1", 1)
	b1.Healthy.Store(false)

	lb := New([]*Backend{b1}, RoundRobin, nil)

	if b := lb.pick(); b != nil {
		t.Fatalf("expected nil when all backends unhealthy, got %v", b.URL)
	}
}

func TestLeastConnections_PicksLowestConns(t *testing.T) {
	b1 := newTestBackend("http://backend1", 1)
	b2 := newTestBackend("http://backend2", 1)

	b1.IncrementConns()
	b1.IncrementConns()
	b2.IncrementConns()

	lb := New([]*Backend{b1, b2}, LeastConnections, nil)

	picked := lb.pick()
	if picked.URL.String() != "http://backend2" {
		t.Fatalf("expected backend2 (fewer connections), got %s", picked.URL.String())
	}
}

func TestWeighted_RespectsRatio(t *testing.T) {
	b1 := newTestBackend("http://backend1", 1)
	b2 := newTestBackend("http://backend2", 3)

	lb := New([]*Backend{b1, b2}, Weighted, nil)

	counts := map[string]int{}
	for i := 0; i < 400; i++ {
		b := lb.pick()
		counts[b.URL.String()]++
	}

	// backend2 has 3x weight, expect roughly 25%/75% split
	ratio := float64(counts["http://backend2"]) / float64(counts["http://backend1"])
	if ratio < 2.0 || ratio > 4.0 {
		t.Errorf("expected backend2 to get roughly 3x traffic of backend1, got ratio %.2f", ratio)
	}
}