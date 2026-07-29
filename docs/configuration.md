# Application Configuration

The API uses environment variables as its canonical runtime configuration
source. It does not load `.env` files itself. Local shells, Docker Compose, and
CI are responsible for supplying the environment.

`RELOHELPER_DB_DSN` is the only required API variable. All other settings have
defaults.

## API Variables

| Variable | Default | Description |
|---|---:|---|
| `RELOHELPER_ENV` | `development` | Runtime environment: `development`, `staging`, or `production` |
| `RELOHELPER_PORT` | `4000` | Public API listener port |
| `RELOHELPER_DB_DSN` | required | PostgreSQL connection string |
| `RELOHELPER_DB_MAX_OPEN_CONNS` | `25` | Maximum PostgreSQL pool size |
| `RELOHELPER_DB_MAX_IDLE_TIME` | `15m` | Maximum idle time for a pooled connection |
| `RELOHELPER_LIMITER_RPS` | `10` | Rate limiter requests per second |
| `RELOHELPER_LIMITER_BURST` | `20` | Rate limiter burst size |
| `RELOHELPER_LIMITER_ENABLED` | `true` | Enable the HTTP rate limiter |
| `RELOHELPER_AUTH_ENABLED` | `true` | Require authentication on protected endpoints |
| `RELOHELPER_METRICS_PORT` | `0` | Dedicated metrics port; `0` serves metrics on the API listener |
| `RELOHELPER_BATCH_MAX_IDS` | `100` | Maximum IDs in a batch request |
| `RELOHELPER_BATCH_MAX_DETAILED_IDS` | `20` | Maximum IDs in a batch request with detailed includes |
| `RELOHELPER_EXCHANGE_RATES_APP_ID` | empty | Open Exchange Rates application ID |
| `RELOHELPER_SMTP_HOST` | empty | SMTP server hostname |
| `RELOHELPER_SMTP_PORT` | `25` | SMTP server port |
| `RELOHELPER_SMTP_USERNAME` | empty | SMTP username |
| `RELOHELPER_SMTP_PASSWORD` | empty | SMTP password |
| `RELOHELPER_SMTP_SENDER` | `Relohelper <no-reply@relohelper.local>` | Email sender |

Invalid values are rejected during startup before the API opens a database
connection or starts listening.

## Local Development

The recommended local setup uses
[`direnv`](https://direnv.net/) to load a project-specific `.envrc`:

```bash
cp .envrc.example .envrc
# Edit .envrc with local credentials.
direnv allow
```

`.envrc` is ignored by Git and the Docker build context. Values containing
spaces or shell metacharacters must be quoted:

```bash
export RELOHELPER_SMTP_SENDER='Relohelper <no-reply@relohelper.local>'
```

Without `direnv`, export the same variables manually in the current shell or
run `source .envrc` before using `make`. The Makefile does not parse secret
files itself.

Run the API with authentication disabled:

```bash
make run/api
```

Use `make run/api/auth` to enable authentication and `make run/api/load` to
disable both authentication and rate limiting for local load tests.

## Authentication

`RELOHELPER_AUTH_ENABLED` controls API access enforcement:

- `true`: Bearer tokens are processed, and city/country detail endpoints
  require an activated user.
- `false`: city/country detail endpoints are public.

The registration, account activation, and token endpoints remain available in
both modes. This setting does not disable the email subsystem itself.

For local development, use:

```bash
make run/api       # authentication disabled
make run/api/auth  # authentication enabled
```

For production, set the value in `/opt/relohelper-go/.env` and recreate only
the API container:

```bash
RELOHELPER_AUTH_ENABLED=false

docker compose --env-file .env -f deploy/docker-compose.yml \
  up -d --no-deps api
```

When authentication and email activation are used, configure the SMTP
variables as well. The sender variable is named `RELOHELPER_SMTP_SENDER`, not
`RELOHELPER_SMTP_FROM`.

## Production

The VPS uses the root `.env` file as a Docker Compose input. Start from
`.env.example`; Compose explicitly passes API settings into the container.

The production stack fixes the API and metrics ports at `4000` and `4001`
because Caddy and Prometheus use those internal addresses.

Environment variables are the canonical configuration interface. Existing
runtime CLI flags remain temporarily available as higher-priority overrides so
the production stack can roll back to `v0.5.0`. New scripts should not depend
on them. After `v0.6.0` is established as the rollback baseline, remove the
Compose CLI overrides first. Remove CLI parsing from the API in a later release.

The permanent operational CLI option is the diagnostic version command:

```bash
/app/api -version
```
