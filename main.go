package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rithish279/resilient-lb/pkg/api"
	"github.com/rithish279/resilient-lb/pkg/chaos"
	"github.com/rithish279/resilient-lb/pkg/lb"
)
func getBackends() []string {
	raw := os.Getenv("BACKENDS")

	if (raw == "") {
		// Default local
		return []string{
			"http://localhost:9001",
			"http://localhost:9002",
		}
	}
	return strings.Split(raw, ",")
}

func main() {
	backendURLs := getBackends()

	backends := make([]*lb.Backend, len(backendURLs))
	for i, url := range backendURLs {
		backends[i] = lb.NewBackend(strings.TrimSpace(url), 1)
	}

	lb.StartHealthChecks(backends, 5*time.Second, 2*time.Second)

	chaosEngine := chaos.New()
	balancer := lb.New(backends, lb.RoundRobin, chaosEngine)

	// Traffic server
	go func() {
		fmt.Println("load balancer starting on :8080")
		log.Fatal(http.ListenAndServe(":8080", balancer))
	} ()

	// Admin server
	adminMux := http.NewServeMux()
	api.NewAdminHandler(backends).RegisterRoutes(adminMux)
	api.NewChaosHandler(chaosEngine).RegisterRoutes(adminMux)
	adminMux.Handle("/metrics", promhttp.Handler())

	fmt.Println("admin api starting on :8888")
	log.Fatal(http.ListenAndServe(":8888", adminMux))
}