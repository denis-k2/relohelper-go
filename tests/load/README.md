# Locust Load Tests

These load tests exercise the main public city and country API routes using a realistic mix of list, detail, and batch requests.

## Environment Setup

Create and activate a virtual environment from the repository root:

```bash
uv venv .venv
source .venv/bin/activate
```

Install dependencies:

```bash
uv pip install -r tests/load/requirements.txt
```

## Running Locust

Start the API first in load-test mode, then run Locust from the repository root.

For example:

```bash
make run/api/load
```

The load-test target disables:

- authentication
- rate limiting

Web UI example:

```bash
locust -f tests/load/locustfile.py --host=http://127.0.0.1:4000
```

The Locust UI is available at `http://0.0.0.0:8089` by default.

Headless example:

```bash
locust -f tests/load/locustfile.py --host=http://127.0.0.1:4000 --headless --users 20 --spawn-rate 5 --run-time 2m
```

## Base URL

The load tests expect the API base URL to be provided via Locust `--host`, for example:

```bash
--host=http://127.0.0.1:4000
```

## Scenario Mix

The test bootstrap fetches real data once at startup from:

- `GET /countries`
- `GET /cities`

It then randomly selects valid `country_code` and `geoname_id` values for all subsequent requests.
