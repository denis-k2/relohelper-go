# VPS Deploy with Docker Compose and Caddy

This deploy setup is intended for an Ubuntu VPS and keeps the public surface area small:

- public: Caddy on `80/443`
- public API: `https://relohelper.pro`
- internal only: PostgreSQL, Prometheus, Grafana, API metrics listener

## Files

- Deploy compose: `deploy/docker-compose.yml`
- Caddy config: `deploy/Caddyfile`
- API image build: `Dockerfile`
- Env template: `.env.example`

## DNS

Create an `A` record:

- `relohelper.pro -> <your VPS IPv4 address>`

Wait until DNS resolves before starting Caddy, otherwise automatic HTTPS issuance will fail.

## VPS prerequisites

Install on the VPS:

- Docker Engine
- Docker Compose plugin

Open firewall ports:

- `22/tcp`
- `80/tcp`
- `443/tcp`

## Environment

Create `.env` in the repository root:

```bash
cp .env.example .env
```

Fill in at least:

- `POSTGRES_DB`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`
- `RELOHELPER_DB_MAX_OPEN_CONNS`
- `RELOHELPER_LIMITER_RPS`
- `RELOHELPER_LIMITER_BURST`
- `RELOHELPER_AUTH_ENABLED`
- `RELOHELPER_LIMITER_ENABLED`
- `GRAFANA_ADMIN_USER`
- `GRAFANA_ADMIN_PASSWORD`
- SMTP settings if email delivery is required

## Run on VPS

From the repository root:

```bash
docker compose --env-file .env -f deploy/docker-compose.yml up -d --build
```

## Public URLs

- API: `https://relohelper.pro`
- Swagger UI: `https://relohelper.pro/swagger`
- Healthcheck: `https://relohelper.pro/healthcheck`
- Readiness: `https://relohelper.pro/readyz`

In this deploy mode, Swagger remains public while `/debug/vars` stays disabled.

## Internal-only services

These are reachable only inside the Docker network by default:

- PostgreSQL
- Prometheus
- API metrics listener on `api:4001`

Grafana is not public on the internet, but it is bound to loopback on the VPS only:

- `127.0.0.1:3000 -> grafana:3000`

Prometheus scrapes metrics from:

- `http://api:4001/metrics`

`/metrics` is not exposed on the public API listener and is also blocked at the Caddy layer.

## Verification

After startup, verify:

```bash
curl -I https://relohelper.pro/healthcheck
curl -I https://relohelper.pro/readyz
curl -I https://relohelper.pro/metrics
```

Expected results:

- `/healthcheck` -> `200`
- `/readyz` -> `200` once DB is ready
- `/metrics` -> not public (`404`)

Also verify that these are not exposed publicly:

- `:5432`
- `:9090`

Grafana should only listen on VPS loopback, not on a public interface.

## Accessing Grafana securely

Use SSH tunneling from your local machine:

```bash
ssh -L 3000:localhost:3000 <user>@<your_vps_ip>
```

Then open locally in your browser:

- `http://127.0.0.1:3000`

## Notes

- Because the deploy compose file lives in `deploy/`, start it with `--env-file .env` from the repository root so Compose picks up the root `.env`.
- This deploy stack uses the PostgreSQL 18 container layout and mounts the data volume at `/var/lib/postgresql`.
- API runtime toggles for VPS deploy are controlled from `.env` and injected into the container command by Docker Compose:
  - `RELOHELPER_DB_MAX_OPEN_CONNS`
  - `RELOHELPER_LIMITER_RPS`
  - `RELOHELPER_LIMITER_BURST`
  - `RELOHELPER_AUTH_ENABLED=true|false`
  - `RELOHELPER_LIMITER_ENABLED=true|false`
- After changing these values, apply them with:

```bash
docker compose --env-file .env -f deploy/docker-compose.yml up -d
```

Compose will usually recreate only the `api` container when only its command changes. PostgreSQL, Prometheus, Grafana, and Caddy are not rebuilt or restarted unless their own configuration changes.

## Local development remains unchanged

This deploy stack does not replace the existing local workflow.

For local development, keep using:

- PostgreSQL separately
- `go run ./cmd/api`
- `monitoring/docker-compose.yml` when needed
