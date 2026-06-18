# City Comparison Query Analysis

Date: 2026-06-18

## Scope

This run inspected the SQL query path used by the dashboard city comparison
flow for a maximum detailed batch of 20 cities:

- `GetCitiesByIDs`
- `GetNumbeoCostByCityIDs`
- `GetNumbeoCityIndicesByCityIDs`
- `GetAvgClimateByCityIDs`
- country-level follow-up queries

The analysis used local PostgreSQL with:

```sql
EXPLAIN (ANALYZE, BUFFERS)
```

The city batch used the 20 most populated cities in the local dataset. Timings
below are PostgreSQL execution times from the local dev database; they are query
level timings, not end-to-end HTTP latency.

## Findings

The largest avoidable cost was in `GetAvgClimateByCityIDs`.

The previous query converted each climate row to JSON, expanded it with
`jsonb_each`, grouped by metric key, and then used a per-city subplan to build
the final object. On a 20-city batch this produced repeated scans over the same
CTE data.

Observed local execution time:

- Before: about `114 ms`
- After: about `7 ms`

The optimized query now aggregates each climate metric explicitly with
`jsonb_agg(... ORDER BY month)` in a single grouped pass over `avg_climate`.

## Query Timings

| Query | Before | After | Status |
| --- | ---: | ---: | --- |
| `GetCitiesByIDs` | `3.6 ms` | unchanged | Already fast on current dataset |
| `GetNumbeoCostByCityIDs` | `41-54 ms` | unchanged | Investigated; no safe simple SQL win found |
| `GetNumbeoCityIndicesByCityIDs` | `0.8 ms` | unchanged | Already fast |
| `GetAvgClimateByCityIDs` | `114 ms` | `7.3 ms` | Optimized |
| `GetCountriesByCodes` | `0.3 ms` | unchanged | Already fast |
| `GetNumbeoCountryIndicesByCodes` | `0.5 ms` | unchanged | Already fast |
| `GetLegatumCountryIndicesByCodes` | `9.0 ms` | unchanged | Acceptable for current dataset |

## `GetNumbeoCostByCityIDs` Variants

`GetNumbeoCostByCityIDs` remains the most expensive comparison query in this
local dataset, mostly due to building detailed cost payloads for roughly 1,100
cost rows in a 20-city batch.

Additional variants were checked:

| Variant | Observed time | Result |
| --- | ---: | --- |
| Current query with `jsonb_*` | `41-54 ms` | Baseline |
| `json_*` instead of `jsonb_*` | `44 ms` | No meaningful improvement |
| Per-city lateral aggregate | `35 ms` | Some improvement, but more complex query shape |
| Temporary covering index on `(geoname_id, param_id)` | `44 ms` | No meaningful improvement |
| Raw sorted rows without JSON aggregation | `39 ms` | Still similar; would move response assembly to Go |
| Numeric sort keys inside JSON aggregation | `31 ms` | Faster, but changes JSON array order |

None of these showed a clear, safe improvement comparable to the climate query
change. Ordering by numeric category/param ids was faster, but it changed the
JSON array order compared with the existing API response, so it was not applied.
