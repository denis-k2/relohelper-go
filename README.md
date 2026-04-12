# Relohelper (Go Edition)

Relohelper is a REST API service rewritten in Go as a high-performance successor to the original Python prototype.

The Go implementation is now the primary and actively developed version of the project.

Starting from [`v0.3.0`](https://github.com/denis-k2/relohelper-go/releases/tag/v0.3.0), the project also includes a built-in dashboard UI on top of the API. The UI is intended as a reference client for comparing cities by cost of living, climate, and quality of life, while the API remains the core product surface.

## About the Project

This project focuses on:

- Clean API design
- Structured architecture
- Performance-oriented implementation
- Reproducible load testing
- A reference dashboard UI built on top of the API

The original Python implementation can be found [here](https://github.com/denis-k2/relohelper).

## Documentation

Project documentation is located in the `docs/` directory:

- `docs/architecture/` — system structure and design
- `docs/performance/` — load testing methodology and benchmark runs

Performance comparison between Go and Python implementations is documented [here](./docs/performance/runs/2025-04-22_compare-go-v0.1.0_py-v0.3.0/README.md).

## Public Access

- Dashboard UI: [https://relohelper.pro/](https://relohelper.pro/)
- Swagger UI: [https://relohelper.pro/swagger](https://relohelper.pro/swagger)
- `/metrics` is internal-only
- Grafana is available through SSH tunneling only

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

## VPS Deploy with Docker Compose and Caddy

For production-like deployment on an Ubuntu VPS, use the dedicated deploy stack.

### Files

- VPS compose: `deploy/docker-compose.yml`
- Caddy config: `deploy/Caddyfile`
- Deploy guide: `deploy/DEPLOY.md`
- API image build: `Dockerfile`
- Env template: `.env.example`

See [deploy/DEPLOY.md](./deploy/DEPLOY.md) for:

- DNS setup
- firewall requirements
- environment variables
- startup commands
- public and internal service exposure
- secure Grafana access via SSH tunnel

Run the VPS stack from the repository root with:

```bash
docker compose --env-file .env -f deploy/docker-compose.yml up -d --build
```

## Recommendation by Use Case

- Local development:
  - PostgreSQL separately
  - API via `go run`
  - monitoring via `monitoring/docker-compose.yml` when needed
- VPS / production-like deploy:
  - `deploy/docker-compose.yml`
  - `deploy/Caddyfile`

More detailed project description will be added as development progresses.
