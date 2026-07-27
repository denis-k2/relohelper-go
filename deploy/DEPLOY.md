# VPS Deploy with Docker Compose and Caddy

This deploy setup is intended for an Ubuntu VPS and keeps the public surface area small:

- public: Caddy on `80/443`
- public API: `https://relohelper.pro`
- internal only: PostgreSQL, Prometheus, Grafana, API metrics listener

## Files

- Deploy compose: `deploy/docker-compose.yml`
- Caddy config: `deploy/Caddyfile`
- API image: `ghcr.io/denis-k2/relohelper-go`
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
- `RELOHELPER_IMAGE_TAG`
- `RELOHELPER_DB_MAX_OPEN_CONNS`
- `RELOHELPER_LIMITER_RPS`
- `RELOHELPER_LIMITER_BURST`
- `RELOHELPER_AUTH_ENABLED`
- `RELOHELPER_LIMITER_ENABLED`
- `GRAFANA_ADMIN_USER`
- `GRAFANA_ADMIN_PASSWORD`
- SMTP settings if email delivery is required

Use `edge` only for testing the latest `main` build. For a stable deployment,
pin `RELOHELPER_IMAGE_TAG` to a release such as `0.5.0` or to an immutable
`sha-<commit>` tag.

## GHCR access

The API image contains no secrets. If the GHCR package is public, the VPS can
pull it without authentication.

If the package is private, log in once on the VPS with a GitHub token that has
the `read:packages` permission:

```bash
echo "$GHCR_TOKEN" | docker login ghcr.io -u denis-k2 --password-stdin
```

Do not store `GHCR_TOKEN` in the project `.env` file.

## GitHub Actions deployment

The `Deploy` workflow provides a manually confirmed production deployment:

1. Open `Actions -> Deploy`.
2. Select `Run workflow`.
3. Keep the workflow branch set to `main`.
4. Enter an image tag such as `edge`, `0.5.0`, or `sha-<commit>`.
5. Run the workflow.

The workflow updates the repository on the VPS, pulls the selected image,
starts the Compose stack, waits for `/readyz`, and reports the deployed version.
It does not automatically roll back database migrations.

Configure these repository or `production` environment variables in
`Settings -> Secrets and variables -> Actions`:

- `VPS_HOST`: VPS IPv4 address or DNS name
- `VPS_PORT`: SSH port; optional, defaults to `22`
- `VPS_USER`: SSH deployment user
- `VPS_DEPLOY_PATH`: absolute path to the repository on the VPS

Configure these secrets:

- `VPS_SSH_PRIVATE_KEY`: private key used only by GitHub Actions
- `VPS_SSH_KNOWN_HOSTS`: verified SSH host key entry for the VPS

Create a dedicated deployment key locally:

```bash
ssh-keygen -t ed25519 -C github-actions-relohelper -f ./relohelper_deploy_key
```

Append `relohelper_deploy_key.pub` to the deployment user's
`~/.ssh/authorized_keys` on the VPS. Store the contents of
`relohelper_deploy_key` in `VPS_SSH_PRIVATE_KEY`.

Obtain the host key entry:

```bash
ssh-keyscan -H -p 22 <your_vps_ip>
```

Verify its fingerprint against the VPS host key before storing the output in
`VPS_SSH_KNOWN_HOSTS`. The deployment user must have access to the repository
directory and permission to run Docker without `sudo`.

## Run on VPS

From the repository root:

```bash
docker compose --env-file .env -f deploy/docker-compose.yml pull
docker compose --env-file .env -f deploy/docker-compose.yml up -d
```

The VPS downloads the image built by GitHub Actions. It does not compile the
Go project or retain a Go builder image.

Verify the version embedded in the API image:

```bash
docker compose --env-file .env -f deploy/docker-compose.yml \
  run --rm --no-deps api /app/api -version
```

## Updating

Pull the current deployment files, select the image tag in `.env`, and apply
the update:

```bash
git pull --ff-only
docker compose --env-file .env -f deploy/docker-compose.yml pull
docker compose --env-file .env -f deploy/docker-compose.yml up -d
```

Compose recreates the API container when its image changes. PostgreSQL data
remains in the existing named volume.

## Rollback

Set `RELOHELPER_IMAGE_TAG` in `.env` to the previously working release or
`sha-<commit>` tag, then run:

```bash
docker compose --env-file .env -f deploy/docker-compose.yml pull api
docker compose --env-file .env -f deploy/docker-compose.yml up -d api
```

This rolls back the API image only. Database migrations are not automatically
reverted; schema-changing releases require a compatible migration plan and a
verified backup.

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
- Docker container logs are rotated with:
  - `max-size=10m`
  - `max-file=5`
- Prometheus retention is limited to:
  - `7d`
  - `1GB`
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
