package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"
)

// enabling internal connection pool reusablilty across requests
type Backend struct {
	URL 	*url.URL
	Healthy	atomic.Bool
	proxy	*httputil.ReverseProxy
}

// NewBackend constructs a Backend and wires up its reverse proxy
func NewBackend(rawURL string) *Backend {
	u := mustParseURL(rawURL)
	b := &Backend{
		URL: u,
		proxy: httputil.NewSingleHostReverseProxy(u),
	}
	b.Healthy.Store(true)
	return b
}

func startHealthChecks(backends []*Backend, interval, timeout time.Duration) {
	
	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // No redirects in Health checks
		},
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			for _, b := range backends {
				checkHealth(b, client)
			}
		}
	}()
}

func checkHealth(b *Backend, client *http.Client) {
	url := b.URL.String() + "/"
	resp, err := client.Get(url)

	// Healthy only on clean 200 OK
	isHealthy := err == nil && resp.StatusCode == http.StatusOK
	if resp != nil {
		resp.Body.Close()
	}

	// log only on transistions, not every tick
	prev := b.Healthy.Swap(isHealthy)
	if prev == isHealthy {
		return
	}

	if isHealthy {
		log.Printf("[health] backend %s is UP", b.URL)
	} else {
		log.Printf("[health] backend %s is DOWN", b.URL)
	}
}

func main() {
	backends := []*Backend{
		NewBackend("http://localhost:9001"),
		NewBackend("http://localhost:9002"),
	}

	// Health checks should start before traffic
	startHealthChecks(backends, 5*time.Second, 2*time.Second)

	var counter atomic.Uint32
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		idx := counter.Add(1) % uint32(len(backends))
		backend := backends[idx]

		if !backend.Healthy.Load() {
			for i := 0; i < len(backends); i++ {
				idx = (idx + 1) % uint32(len(backends))
				if backends[idx].Healthy.Load() {
					backend = backends[idx]
					break
				}
			}
		}

		if !backend.Healthy.Load() {
			http.Error(w, "no healthy backends available", http.StatusServiceUnavailable)
			return 
		}

		backend.proxy.ServeHTTP(w, r)
	})

	fmt.Println("load balancer starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func mustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return u
}