# Architecture

This document describes the architecture and design decisions behind **resilient-lb**, a Go-based load balancer featuring active health checks, circuit breakers, configurable load-balancing algorithms, chaos injection, and Prometheus/Grafana observability.

---

## Overview

```text
                     Client Traffic
                           │
                           ▼
┌─────────────────────────────────────────────────────┐
│                 Load Balancer (:8080)               │
│                                                     │
│  1. Select backend                                  │
│  2. Check circuit breaker                           │
│  3. Apply chaos injection (if active)               │
│  4. Forward request                                 │
│  5. Record metrics & circuit outcome                │
└──────────────────────┬──────────────────────────────┘
                       │
          ┌────────────┼────────────┐
          ▼            ▼            ▼
      backend1     backend2     backend3


┌─────────────────────────────────────────────────────┐
│                  Admin API (:8888)                  │
│                                                     │
│  /backends   /chaos/*   /metrics                    │
└─────────────────────────────────────────────────────┘
```

---

## Components

| Component | Responsibility |
|-----------|----------------|
| **Load Balancer** | Accepts client requests and routes them using the configured balancing algorithm. |
| **Backend Pool** | Stores backend state including health, active connections, weight, and circuit breaker status. |
| **Reverse Proxy** | Forwards requests to backend servers using Go's `httputil.ReverseProxy`. |
| **Health Checker** | Periodically probes each backend and updates its health status. |
| **Circuit Breaker** | Prevents requests from reaching backends that are repeatedly failing. |
| **Chaos Engine** | Injects latency, HTTP errors, dropped requests, or kill-switch failures during experiments. |
| **Admin API** | Exposes backend state, chaos controls, and Prometheus metrics. |
| **Observability** | Publishes metrics to Prometheus for visualization in Grafana. |

---

## Request Path

Every request follows the same processing pipeline inside
`LoadBalancer.ServeHTTP`.

1. **Backend Selection**
   - A backend is selected using the configured algorithm:
     - Round Robin
     - Least Connections
     - Weighted Round Robin

2. **Circuit Breaker Check**
   - If the selected backend's circuit breaker is **Open**, the request is rejected immediately.
   - The request never reaches the backend or chaos engine.

3. **Chaos Injection**
   - If an active experiment targets the selected backend, the request may:
     - return an injected HTTP error
     - be delayed
     - be dropped
     - trigger a simulated kill switch

4. **Reverse Proxy**
   - Requests that survive previous stages are forwarded to the backend.

5. **Outcome Recording**
   - Prometheus metrics are updated.
   - The circuit breaker records success or failure.
   - Connection counters are adjusted.

---

# Design Decisions

## Two-Port Architecture

Client traffic and operational endpoints are intentionally separated.

| Port | Purpose |
|------|----------|
| **8080** | Client requests |
| **8888** | Administration, metrics, chaos API |

Keeping the management interface separate prevents operational endpoints from mixing with production traffic. In a real deployment, the admin port would typically be restricted to an internal network or protected behind authentication.

---

## Chaos is Part of the Request Pipeline

The chaos engine is integrated directly into the request path rather than existing as an external simulator.

```
Backend Selection
        │
        ▼
Circuit Breaker
        │
        ▼
Chaos Injection
        │
        ▼
Reverse Proxy
```

This design ensures that simulated failures are processed exactly like genuine backend failures.

For example, an injected HTTP 503 response is indistinguishable from a real backend-generated 503 to the circuit breaker. This allows resilience mechanisms to be validated under realistic fault conditions.

---

## Independent Health Checks and Circuit Breakers

Health checking and circuit breaking measure different aspects of backend health.

| | Health Checker | Circuit Breaker |
|---|---|---|
| Trigger | Periodic timer | Live requests |
| Frequency | Every 5 seconds | Every request |
| Bypasses chaos | Yes | No |
| Detects | Backend reachability | Request failures |
| Purpose | Determine availability | Prevent repeated failures |

A backend may be reachable while still failing client requests.

For example, during a chaos experiment the health checker continues to report the backend as healthy because it bypasses chaos, while the circuit breaker opens after repeated injected failures.

This behavior is demonstrated in **EXP-001** of the Chaos Engineering Runbook.

---

## Metrics Consistency

Runtime metrics are updated immediately whenever backend state changes.

Examples include:

- Active connections
- Circuit breaker transitions
- Request totals
- Request duration
- Backend health

Metrics are updated even on early-return paths (such as circuit breaker rejection) so that Prometheus always reflects the current runtime state exposed by the Admin API.

---

## Observability

```
Load Balancer
      │
      ▼
   metrics
      │
      ▼
  Prometheus
      │
      ▼
    Grafana
```

The project includes a fully containerized observability stack.

- Prometheus scrapes metrics every 5 seconds.
- Grafana dashboards and data sources are automatically provisioned.
- Running `docker-compose up` creates a ready-to-use monitoring environment without additional configuration.

---

## Future Improvements

- Retry policies with exponential backoff for transient backend failures.
- Consistent hashing for sticky session support.
- Passive health checking based on live request failures in addition to active probes.
- OpenTelemetry tracing for end-to-end request visibility.