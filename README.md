# resilient-lb

A HTTP load balancer and reverse proxy with a built-in chaos engineering framework, written in Go.

Built to explore load balancing internals, resilience patterns, and how chaos engineering validates self-healing systems — without hiding behind frameworks.

---

## Features

**Load Balancer**
- Three algorithms — Round Robin, Least Connections, Weighted Round Robin
- Active health checking with automatic failover and recovery
- Per-backend circuit breaker (closed → open → half-open) reacting to live traffic failures
- Connection pooling via Go's `httputil.ReverseProxy`

**Chaos Engineering**
- Four failure modes — `latency`, `error`, `drop`, `killswitch`
- Per-backend targeting or global injection, with configurable probability
- Chaos-induced failures feed into the circuit breaker exactly like real failures
- REST API to inject, inspect, and clear chaos at runtime

**Observability**
- Prometheus metrics — request rate, latency (p50/p95/p99), backend health, circuit breaker state, active connections, chaos injection count
- Auto-provisioned Grafana dashboard — zero manual setup, just `docker-compose up`
- Admin API on a separate port, isolated from user traffic

**Tested & Documented**
- Unit tests for circuit breaker state transitions and all three algorithms
- [Chaos runbook](deploy/docs/chaos-runbook.md) with real documented experiments, evidence, and findings

---

## Quick Start (Docker)

```bash
cd docker
docker-compose up --build
```

This starts the load balancer, 3 backend servers, Prometheus, and Grafana — all wired together automatically.

| Service | URL |
|---|---|
| Load balancer | http://localhost:8080 |
| Admin API | http://localhost:8888 |
| Prometheus | http://localhost:9090 |
| Grafana (admin/admin) | http://localhost:3000 |

The Grafana dashboard ("Resilient LB — Overview") and Prometheus data source are pre-configured — no manual setup needed.

---

## Quick Start (Local, no Docker)

```bash
# Terminal 1 & 2 — start two backends
go run cmd/backend/main.go -port 9001
go run cmd/backend/main.go -port 9002

# Terminal 3 — start the load balancer
go run main.go
```

---

## Admin API

| Method | Path | Description |
|---|---|---|
| `GET` | `/backends` | Live backend status: health, active connections, circuit breaker state |
| `GET` | `/metrics` | Prometheus metrics |
| `POST` | `/chaos/inject` | Start a chaos experiment |
| `GET` | `/chaos/status` | Current chaos state |
| `DELETE` | `/chaos/clear` | Stop all chaos injection |

### Example — inject latency

```bash
curl -X POST http://localhost:8888/chaos/inject \
  -H "Content-Type: application/json" \
  -d '{"type":"latency","latency_ms":500,"probability":1.0,"duration_sec":30}'
```

### Example — inject errors targeting one backend

```bash
curl -X POST http://localhost:8888/chaos/inject \
  -H "Content-Type: application/json" \
  -d '{"type":"error","target":"http://backend1:9001","error_code":503,"probability":1.0,"duration_sec":60}'
```

Supported `type` values: `latency`, `error`, `drop`, `killswitch`

---

## Running Tests

```bash
go test ./pkg/lb/... -v
```

Covers circuit breaker state transitions (open/closed/half-open) and correctness of all three load balancing algorithms.

---

## Project Structure
resilient-lb/
├── main.go                    # Entry point
├── cmd/backend/main.go        # Test backend server
├── pkg/
│   ├── lb/                    # Load balancer, algorithms, circuit breaker, health checks
│   ├── chaos/                 # Chaos injection engine
│   └── metrics/               # Prometheus metric definitions
├── api/                       # Admin API handlers
├── docker/                    # Dockerfiles, docker-compose, Grafana provisioning
└── deploy/docs/                # Architecture notes, chaos runbook, evidence

---

## Documentation

- [Architecture](deploy/docs/architecture.md)
- [Chaos Runbook](deploy/docs/chaos-runbook.md) — real experiments with evidence
