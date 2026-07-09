package lb

import (
	"log"
	"net/http"
	"time"

	"github.com/rithish279/resilient-lb/metrics"
)

func StartHealthChecks(backends []*Backend, interval, timeout time.Duration) {

	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
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

	isHealthy := err == nil
	if resp != nil {
		resp.Body.Close()
	}

	prev := b.Healthy.Swap(isHealthy)
	if prev == isHealthy {
		return
	}

	if isHealthy {
		log.Printf("[health] backend %s is UP", b.URL)
	} else {
		log.Printf("[health] backend %s is DOWN", b.URL)
	}

	healthVal := 0.0
	if (isHealthy) {
		healthVal = 1.0
	}
	metrics.BackendHealth.WithLabelValues(b.URL.String()).Set(healthVal)
}
