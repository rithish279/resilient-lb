package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

var requestCount atomic.Uint64

func main() {
	port := flag.String("port", "9001", "port to listen on")
	fail := flag.Bool("fail", false, "always return 500 (for circuit breaker testing)")
	flag.Parse()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)
		log.Printf("[backend:%s] request #%d %s %s", *port, count, r.Method, r.URL.Path)

		if (*fail) {
			http.Error(w, "simulated failure", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "backend:%s  |  request:#%d  |  time:%s\n", *port, count, time.Now().Format(time.RFC3339))
	})

	log.Printf("backend starting on :%s (fail=%v)", *port, *fail)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}