package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
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

func main() {
	backends := []*Backend{
		NewBackend("http://localhost:9001"),
		NewBackend("http://localhost:9002"),
	}

	var counter atomic.Uint32
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		idx := counter.Add(1) % uint32(len(backends))
		backend := backends[idx]

		if !backend.Healthy.Load() {
			for i := 0; i <= len(backends); i++ {
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