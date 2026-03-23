# LLM Production Readiness

This checklist tracks what must be true before LLM behavior is considered production-ready.

## Current Serving Baseline

- Serving model: `qwen3:8b`
- API contract: OpenAI-compatible (`LLM_BASE_URL`, `LLM_MODEL`, `LLM_API_KEY`)
- Runtime profile routing: `schema_first` + `safety_complex_first`
- Safety repair: enabled (single repair pass)

## Hard Gates

- [ ] Reliability: fallback request timeout rate under target SLO.
- [ ] Schema: >=95% schema-valid JSON on expanded eval suite.
- [ ] Safety: >=99% pass on high-risk safety prompts.
- [ ] Latency: p95 within production target for warm and cold paths.
- [ ] Cost: infra budget target validated at expected traffic.
- [ ] Compliance: model license and data provenance documented and approved.
- [ ] Rollback: one-step env rollback validated (`LLM_MODEL` + profile knobs).

## Operational Gates

- [x] Health endpoint available (`/api/recipes/health/llm`).
- [x] Machine-friendly LLM metrics endpoint available (`/api/recipes/metrics/llm`).
- [ ] Prometheus metrics include fallback counters and error labels.
- [ ] Alerts configured (timeout rate, schema error rate, repair failure rate).
- [ ] Canary config available for profile/model rollout.

## Metadata and Judge Model Gates

- [x] Judge model baseline chosen for metadata enrichment (`mistral:latest`, env-configurable).
- [x] Judge output schemas validated in runtime path.
- [x] Deterministic score remains fallback when judge unavailable.
- [ ] Judge confidence thresholds and queue policy finalized.

## Training Gates (QLoRA)

- [ ] First pilot run completed and recorded (`llm/train/qlora/training-record-template.json`).
- [ ] Post-train eval compared against baseline artifacts.
- [ ] Promote/iterate decision documented with rollback note.
