package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lb_requests_total",
		Help: "Total number of requests handled  by the load balancer",
	}, []string{"backend", "status"})

	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "lb_request_duration_seconds",
		Help: "Request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"backend"})

	ActiveConnections = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "lb_active_connections",
		Help: "Number of active connections per backend",
	}, []string{"backend"})

	BackendHealth = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "lb_backend_health",
		Help: "Backend health state (1 = health, 0 = unhealthy)",
	}, []string{"backend"})

	CircuitBreakerState = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "lb_circuit_breaker_open",
		Help: "Circuit breaker state (1 = open, 0 = closed)",
	}, []string{"backend"})

	ChaosInjection = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lb_chaos_injections_total",
		Help: "Total number of chaos injections applied",
	}, []string{"type"})
)