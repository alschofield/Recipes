# Session Handoff - 2026-03-23

This snapshot captures current LLM/search productionization progress so no context is lost between sessions.

## Completed in This Session Block

- Root Docker compose now includes Ollama service and persistent model volume.
- Recipes server LLM defaults pinned for local self-hosting (`qwen3:8b`).
- Runtime profile routing implemented:
  - `schema_first` default profile
  - `safety_complex_first` complex profile
  - explicit `complex=true` request hint
  - auto-complex for >=10 normalized ingredients
- Fallback runtime controls added:
  - `LLM_DISABLE_THINKING_TAG`
  - `LLM_MAX_TOKENS`
  - `LLM_TIMEOUT_SECONDS`
  - `LLM_REPAIR_TIMEOUT_SECONDS`
- LLM fallback observability added:
  - `GET /recipes/health/llm`
  - nginx route `GET /api/recipes/health/llm`
  - counters for requests/success/errors/repairs
- Machine-friendly LLM metrics endpoint added:
  - `GET /recipes/metrics/llm`
  - nginx route `GET /api/recipes/metrics/llm`
- Judge-model baseline integration added (env-gated):
  - ingredient metadata enrichment for newly created ingredients
  - secondary recipe quality scoring blended with deterministic score by confidence threshold
  - judge health/metrics exposed in LLM health endpoint
- Web recipes page now forwards complex mode and auto-enables it at 10+ ingredients.

## Current Runtime Snapshot

- Health endpoint returns `status=ok` and runtime config/counters.
- Warm inclusive fallback path can return mixed source results (`database` + `llm`).
- Strict sparse queries may still return zero if generated recipes fail strict compatibility filters.

## Key Open Risks

- Need Prometheus-style metric export in base `/recipes/metrics` labels (currently separate `/recipes/metrics/llm` endpoint).
- Need production ingredient governance mode (`auto_create` vs `queue_only`) to control auto-created ingredients.

## Immediate Next Actions

1. Add ingredient governance mode env switch and production default.
2. Add calibration dataset/run for judge model reliability thresholds.
3. Optionally merge LLM/judge counters into base `/recipes/metrics` endpoint payload.
4. Run first QLoRA pilot and compare against baseline eval artifacts.

## Checklist Map

- Root: `CHECKLIST.md`
- Server backlog: `server/CHECKLIST.md`
- Web backlog: `web/CHECKLIST.md`
- LLM backlog: `llm/CHECKLIST.md`
- Mobile backlog: `mobile/CHECKLIST.md`
