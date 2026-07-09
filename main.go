package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rithish279/resilient-lb/pkg/api"
	"github.com/rithish279/resilient-lb/pkg/chaos"
	"github.com/rithish279/resilient-lb/pkg/lb"
)

func main() {
	backends := []*lb.Backend{
		lb.NewBackend("http://localhost:9001", 1),
		lb.NewBackend("http://localhost:9002", 2),
	}

	lb.StartHealthChecks(backends, 5*time.Second, 2*time.Second)

	chaosEngine := chaos.New()
	balancer := lb.New(backends, lb.RoundRobin, chaosEngine)


	// Port 8080 => Traffic Server
	go func() {
		fmt.Println("Load balancer starting on :8080")
		log.Fatal(http.ListenAndServe(":8080", balancer))
	} ()

	// Port 8888 => Admin server
	adminMux := http.NewServeMux()
	adminMux.Handle("/metrics", promhttp.Handler())
	api.NewAdminHandler(backends).RegisterRoutes(adminMux)
	api.NewChaosHandler(chaosEngine).RegisterRoutes(adminMux)

	fmt.Print("admin api starting on :8888")
	log.Fatal(http.ListenAndServe(":8888", adminMux))
}