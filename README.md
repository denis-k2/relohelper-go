**Project Description**  
Relohelper rewritten in Go - a high-performance port of the initial Python prototype. For full description and implementation details, please refer to the [Python version's README](https://github.com/denis-k2/relohelper).

_(Now maintained as primary implementation due to Go's performance advantages)_

## Performance Benchmarking: Go vs Python Implementation

A comparative performance analysis between two implementations of the Relohelper REST API service.

**Implementations Compared**

| Implementation  | Version                                                                 | Key Characteristics                                                                 |
| --------------- | ----------------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| `Relohelper-py` | [v0.2.0](https://github.com/denis-k2/relohelper/releases/tag/v0.2.0)    | FastAPI (latest), async, auth disabled, no observability, no caching, Python 3.12.3 |
| `Relohelper-go` | [v0.1.0](https://github.com/denis-k2/relohelper-go/releases/tag/v0.1.0) | Native HTTP server, auth disabled, no observability, no caching                                        |

**Hardware Configurations**

| Specification   | 1 CPU VPS         | 2 CPU VPS         |
| --------------- | ----------------- | ----------------- |
| vCPUs           | 1                 | 2                 |
| Clock Speed     | 3.3 GHz           | 3.3 GHz           |
| RAM             | 1 GB              | 4 GB              |
| Storage         | NVMe              | NVMe              |
| Network         | 100 Mbps          | 100 Mbps          |
| `Relohelper-py` | [Locust report](https://denis-k2.github.io/Relohelper-Go/LocustReports/py/report-1_cpu_vps(gunicorn_w3).html) | [Locust report](https://denis-k2.github.io/Relohelper-Go/LocustReports/py/report-2_cpu_vps(gunicorn_w3).html) |
| `Relohelper-go` | [Locust report](https://denis-k2.github.io/Relohelper-Go/LocustReports/go/report-1_cpu_vps.html) | [Locust report](https://denis-k2.github.io/Relohelper-Go/LocustReports/go/report-2_cpu_vps.html) |


**Testing Methodology**

- Load testing with [Locust](https://github.com/locustio/locust) framework
- Ramp-up rate of 0.5 users per second
- Fresh [VPS](https://timeweb.cloud) reboot before each test
- Minimal logging during tests
- Authentication was disabled in both implementations to ensure a fair comparison. 
  Although the authentication mechanisms differ (stateful in Go vs. stateless JWT in Python), these were excluded to focus solely on the raw performance aspects of the API processing and database interactions.
- PostgreSQL 16 and Relohelper-Py/Go installed directly on host OS (no Docker)

**Load Testing Overview** (`locust.py`)

- Retrieves the list of country codes at the start
- Simulates virtual users with a random wait time of 1–5 seconds
- Four tasks with weighted distribution:
    - 55.6% (weight 15): `GET` `/cities/{city_id}` (using a random `city_id` from 1 to 534, with all query parameters)
    - 37.0% (weight 10): `GET` `/countries/{country_code}` (using a random country code, with all query parameters)
    - 3.7%  (weight 1):  `GET` `/countries`
    - 3.7%  (weight 1):  `GET` `/cities`

**Execution**
```bash
# Python
gunicorn -w 3 -k uvicorn.workers.UvicornWorker main:app \
  --bind 0.0.0.0:8000

uvicorn main:app --host 0.0.0.0 --port 8000 --log-level critical

# Go
./api -db-dsn=... -limiter-enabled=false \
  db-max-open-conns=100 -db-max-idle-conns=100
```


## Charts
Comparison on 2 CPU VPS.

Detailed Locust reports on the links in the *Hardware Configurations* table or by clicking on the graph.

**Python** (`gunicorn -w 3`)

[![py-2cpu_vps-total_rps](/docs/load-testing/py/reports/2_cpu_vps-total_rps.png)](https://denis-k2.github.io/Relohelper-Go/LocustReports/py/report-2_cpu_vps(gunicorn_w3).html)

**Go**

[![go-2cpu_vps-total_rps](/docs/load-testing/go/reports/2_cpu_vps-total_rps.png)](https://denis-k2.github.io/Relohelper-Go/LocustReports/go/report-2_cpu_vps.html)

**VPS Dashboard**

![2_cpu_vps_6_hours_grade](/docs/load-testing/go/reports/2_cpu_vps_6_hours_grade.jpg)

<details>
  <summary>Comparison on 1 CPU VPS.</summary>

  **Python** (`gonicorn -w 3`)

  [![py-2cpu_vps-total_rps](/docs/load-testing/py/reports/1_cpu_vps-total_rps(gunicorn).png)](https://denis-k2.github.io/Relohelper-Go/LocustReports/py/report-1_cpu_vps(gunicorn_w3).html)

  Supplemental report: [1 CPU VPS `uvicorn`](https://denis-k2.github.io/Relohelper-Go/LocustReports/py/report-1_cpu_vps(uvicorn).html)

  **Go**

  [![go-2cpu_vps-total_rps](/docs/load-testing/go/reports/1_cpu_vps-total_rps.png)](https://denis-k2.github.io/Relohelper-Go/LocustReports/go/report-1_cpu_vps.html)

  **VPS Dashboard**

  ![2_cpu_vps_6_hours_grade](/docs/load-testing/go/reports/1_cpu_vps-stat.jpg)
</details>

## Observations
- **Go outperformed Python by 4x** in requests per second
- Zero errors in Go implementation vs about 0.5% in Python
- The API server is CPU-bound, with the processor reaching its limits while memory usage remains moderate.
- Network saturation occurred first in Go tests (100 Mbps limit)
