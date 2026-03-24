# `/recipes/search` Load Baseline

Status: active baseline template (update when infra or fallback strategy changes)

## Goal

Measure endpoint behavior under concurrent traffic with and without fallback pressure.

## Runner

Use the built-in load tool:

```bash
cd server
go run ./cmd/search-load -scenario fallback-heavy -requests 300 -concurrency 12
go run ./cmd/search-load -scenario db-only -requests 300 -concurrency 12
```

Optional flags:

- `-url` custom endpoint (default `http://localhost:8081/recipes/search`)
- `-token` bearer token if gateway/auth requires it
- `-timeout` per-request timeout

## Scenarios

- `fallback-heavy`: uncommon ingredient set + `strict` + `complex=true` to stress fallback path.
- `db-only`: common pantry set + `dbOnly=true` to isolate DB/search path.

## Output Fields

- `throughput_rps`
- `success`, `failures`, `non200`
- `latency_ms` (`p50`, `p95`, `p99`)

## Baseline Capture Template

Record each run in release notes / ops logs:

```text
date:
environment:
scenario:
requests:
concurrency:
throughput_rps:
success/failures/non200:
latency_ms_p50:
latency_ms_p95:
latency_ms_p99:
notes:
```

## Gate Suggestion

Use these as initial checks until stricter SLOs are set:

- `non200 = 0`
- `failures = 0`
- `p95` should not regress by more than 20% versus last approved baseline for same scenario.
