# Latency and Cost Validation (2026-03-23)

This is the current verification snapshot for expected request volume assumptions.

## Traffic assumptions (initial launch)

- Daily recipe searches: `30,000`
- Peak-hour share: `20%`
- Fallback trigger rate target: `15%`
- Fallback requests per peak hour: `30,000 * 0.20 * 0.15 = 900`
- Required fallback throughput at peak: `900 / 3600 = 0.25 req/s`

## Current measured fallback profile (reference)

Source: `llm/evals/results/qwen3-8b-profile-compare-repair-20260322/profiles_summary.json`

- `qwen3:8b` + `safety_complex_first`
  - avg latency: `56,210.46 ms`
  - p95 latency: `105,376.77 ms`

Approx effective single-worker throughput from avg latency:

- `1 / 56.21s = 0.0178 req/s`

Required concurrent generation slots for peak:

- `0.25 / 0.0178 = 14.0` slots (rounded up)

## Cost posture (self-hosted lane)

- Current stack is self-hosted-first, so direct request billing is avoided.
- Cost is dominated by GPU uptime and capacity sized to required concurrent slots.
- With current latency profile, infra sizing must support about `14` concurrent fallback slots during peak assumptions.

## Verification outcome

- Latency and throughput are verified against an explicit demand assumption.
- Result: current profile is **quality-viable but capacity-expensive** at launch assumptions.
- Action: reduce p95 and timeout rate before production promotion, or lower fallback trigger rate via DB/ranking improvements.
