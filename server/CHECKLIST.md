# Server Checklist

Goal: make recipe search + LLM fallback production-ready with deterministic behavior, measurable quality, and low-ops controls.

Only active/open work is listed here. Completed items are archived in `../docs/archive/checklist-completed-items.md`.

## Search Blend and Ranking

- [ ] Define final blend policy for DB + LLM results (minimum generated count, max generated share, fallback behavior) (ref: `docs/server/search-contract.md`, `docs/llm/fallback-contract.md`).
  Model: `CODEX_HIGH`
- [ ] Add deterministic interleave/shuffle strategy with stable seed support for pagination consistency.
  Model: `CODEX_HIGH`
- [ ] Add blend metadata fields in API response (`source`, `blendSlot`, `rankingReason`).
  Model: `FREE_BALANCED`
- [ ] Add tests for blend determinism, source-mix guarantees, and pagination stability.
  Model: `CODEX_HIGH`

## Fallback Runtime and Reliability

- [ ] Add canary toggle + emergency fallback-disable switch for production incidents (ref: `docs/ops/deployment-plan.md`).
  Model: `CODEX_HIGH`

## Observability and Operations

- [ ] Export fallback counters into `/recipes/metrics` (Prometheus-style labels), not health-only JSON (ref: `docs/llm/fallback-contract.md`).
  Model: `FREE_BALANCED`
- [ ] Add request-id linked logs for fallback lifecycle states (triggered, provider_call, repaired, persisted, skipped).
  Model: `FREE_BALANCED`
- [ ] Define alert thresholds for fallback timeout rate, schema-error rate, and repair-fail rate.
  Model: `FREE_BALANCED`

## Ingredient Governance and Metadata

- [ ] Add env-controlled ingredient policy mode (`auto_create`, `queue_only`) for production governance (ref: `docs/server/ingredient-governance.md`).
  Model: `CODEX_HIGH`
- [ ] Add review queue SLA and dashboards for unresolved ingredient candidates.
  Model: `FREE_BALANCED`

## Judge Model (Non-User-Derived Quality/Metadata)

- [ ] Add calibration tests comparing judge scores to deterministic baseline + manual spot checks (ref: `llm/judge/calibration-template.json`).
  Model: `CODEX_HIGH`

## API/Contract Hardening

- [ ] Keep `docs/server/search-contract.md` and `docs/llm/fallback-contract.md` in lockstep with handler behavior.
  Model: `FREE_FAST`
- [ ] Add strict-mode policy for generated results (explicitly return none vs degrade to inclusive fallback).
  Model: `CODEX_HIGH`
- [ ] Add endpoint-level load test baseline for `/recipes/search` including fallback-heavy traffic.
  Model: `CODEX_HIGH`
