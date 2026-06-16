package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rithish279/resilient-lb/pkg/api"
	"github.com/rithish279/resilient-lb/pkg/lb"
)

func main() {
	backends := []*lb.Backend{
		lb.NewBackend("http://localhost:9001", 1),
		lb.NewBackend("http://localhost:9002", 2),
	}

	lb.StartHealthChecks(backends, 5*time.Second, 2*time.Second)

	balancer := lb.New(backends, lb.RoundRobin)

	// Port 8080 => Traffic
	go func() {
		fmt.Println("Load balancer starting on :8080")
		log.Fatal(http.ListenAndServe(":8080", balancer))
	} ()

	// Port 8888 => Admin
	adminMux := http.NewServeMux()
	adminHandler := api.NewAdminHandler(backends)
	adminHandler.RegisterRoutes(adminMux)

	fmt.Print("admin api starting on :8888")
	log.Fatal(http.ListenAndServe(":8888", adminMux))
}