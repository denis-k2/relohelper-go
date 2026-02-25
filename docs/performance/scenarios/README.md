# Load Testing Scenarios

This directory contains canonical load testing scenarios.

## Principles

- Scenarios define workload type (e.g., `baseline-read`, `write-mix`)
- Scenarios are version-agnostic
- Historical runs preserve scenario snapshots
- Scenarios may evolve, but past runs remain reproducible

## Each Scenario Should Define

- Target endpoints
- Request distribution
- Concurrency model
- Test duration
- Think time (if applicable)

Scenarios should be kept minimal, focused, and clearly named.
