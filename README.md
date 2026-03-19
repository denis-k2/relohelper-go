# Relohelper (Go Edition)

Relohelper is a REST API service rewritten in Go as a high-performance successor to the original Python prototype.

The Go implementation is now the primary and actively developed version of the project.

## About the Project

This project focuses on:

- Clean API design
- Structured architecture
- Performance-oriented implementation
- Reproducible load testing

The original Python implementation can be found [here](https://github.com/denis-k2/relohelper).

## Documentation

Project documentation is located in the `docs/` directory:

- `docs/architecture/` — system structure and design
- `docs/performance/` — load testing methodology and benchmark runs

Performance comparison between Go and Python implementations is documented [here](./docs/performance/runs/2025-04-22_compare-go-v0.1.0_py-v0.3.0/README.md).

## Current Status

- Active development in Go
- Performance testing infrastructure established
- Documentation structure organized for future expansion

## Goals

- Continue improving performance and stability
- Maintain reproducible benchmarking
- Incrementally refine architecture and documentation

## Monitoring Stack

The project includes a minimal Prometheus + Grafana stack on top of the built-in observability endpoints.

### What is available

- API: `http://127.0.0.1:4000`
- Healthcheck: `http://127.0.0.1:4000/healthcheck`
- Readiness: `http://127.0.0.1:4000/readyz`
- Prometheus metrics: `http://127.0.0.1:4000/metrics`
- Expvar: `http://127.0.0.1:4000/debug/vars`
- Prometheus UI: `http://127.0.0.1:9090`
- Grafana: `http://127.0.0.1:3000` (`admin` / `admin`)

### How to run locally

1. Start the API locally.
2. Start the monitoring stack:

```bash
docker compose -f monitoring/docker-compose.yml up -d
```

Prometheus scrapes the API at `host.docker.internal:4000`, so the API should be running on the host machine on port `4000`.

### Files

- Compose: `monitoring/docker-compose.yml`
- Prometheus config: `monitoring/prometheus/prometheus.yml`
- Grafana dashboard: `monitoring/grafana/dashboards/relohelper-api-overview.json`

More detailed project description will be added as development progresses.
