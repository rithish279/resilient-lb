package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rithish279/resilient-lb/pkg/lb"
)

func main() {
	backends := []*lb.Backend{
		lb.NewBackend("http://localhost:9001"),
		lb.NewBackend("http://localhost:9002"),
		lb.NewBackend("http://localhost:9003"),
		lb.NewBackend("http://localhost:9004"),
	}

	lb.StartHealthChecks(backends, 5*time.Second, 2*time.Second)

	balancer := lb.New(backends, lb.RoundRobin)

	http.Handle("/", balancer)

	fmt.Println("load balancer starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}