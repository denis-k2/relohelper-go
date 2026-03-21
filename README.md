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

## Local Development

The existing local workflow stays unchanged:

1. Start PostgreSQL separately.
2. Run the API locally:

```bash
go run ./cmd/api
```

3. If needed, start only the observability stack:

```bash
docker compose -f monitoring/docker-compose.yml up -d
```

This mode is still the recommended setup for everyday local development.

## Full Stack with Docker Compose

For integration testing and VPS-like deployment, the project now includes a full stack compose file in the repository root.

### Files

- Full stack compose: `docker-compose.yml`
- API image build: `Dockerfile`
- Env template: `.env.example`
- Monitoring-only compose: `monitoring/docker-compose.yml`
- Full stack Prometheus config: `monitoring/prometheus/prometheus.compose.yml`

### How to run

1. Create env file from the template:

```bash
cp .env.example .env
```

2. Adjust values in `.env`.

3. Start the full stack:

```bash
docker compose up -d --build
```

The full stack starts:

- `postgres`
- `migrate`
- `api`
- `prometheus`
- `grafana`

### Available URLs

- API: `http://127.0.0.1:4000`
- Healthcheck: `http://127.0.0.1:4000/healthcheck`
- Readiness: `http://127.0.0.1:4000/readyz`
- Metrics: `http://127.0.0.1:4000/metrics`
- Prometheus: `http://127.0.0.1:9090`
- Grafana: `http://127.0.0.1:3000`

## Recommendation by Use Case

- Local development:
  - PostgreSQL отдельно
  - API через `go run`
  - monitoring при необходимости через `monitoring/docker-compose.yml`
- Integration run:
  - root `docker-compose.yml`
- VPS / deploy-like run:
  - root `docker-compose.yml` as the simplest production-like baseline

More detailed project description will be added as development progresses.
