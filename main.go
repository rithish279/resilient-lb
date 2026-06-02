package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type Backend struct {
	URL 	*url.URL
	Healthy	atomic.Bool
}

func main() {
	backends := []*Backend{
		{URL: mustParseURL("http://localhost:9001")},
		{URL: mustParseURL("http://localhost:9002")},
	}
	backends[0].Healthy.Store(true)
	backends[1].Healthy.Store(true)

	var counter uint32 = 0
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// simple round-robin: pick next automatically
		idx := atomic.AddUint32(&counter, 1) % uint32(len(backends))
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
			http.Error(w, "No healthy backends", http.StatusServiceUnavailable)
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(backend.URL)
		proxy.ServeHTTP(w, r)
	})

	fmt.Println("Load balancer starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func mustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return u
}