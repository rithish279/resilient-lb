# resilient-lb

A HTTP load balancer and reverse proxy with a built-in chaos engineering framework, written in Go.

## Features
- Multiple load balancing algorithms — Round Robin, Least Connections, Weighted Round Robin
- Active health checking with automatic failover and recovery (in progress)
- Admin API on a separate port for live backend status
- Chaos engineering framework (in progress)

## Run

```bash
# Start backends
go run cmd/backend/main.go -port 9001
go run cmd/backend/main.go -port 9002

# Start load balancer
go run main.go
```

## Admin API

```bash
# Backend status
curl http://localhost:8888/backends
```