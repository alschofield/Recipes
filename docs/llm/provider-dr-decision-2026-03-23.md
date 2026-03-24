# Backup Provider DR Decision (2026-03-23)

Decision: **keep a hosted OpenAI-compatible backup provider** as disaster-recovery lane.

## Rationale

- Primary serving direction remains self-hosted (`qwen3:8b`) for cost control.
- Hosted fallback provides operational continuity when self-hosted inference is degraded.
- Existing server contract already supports provider swap through env vars (`LLM_BASE_URL`, `LLM_MODEL`, `LLM_API_KEY`) without code changes.

## Operating mode

- Backup provider is not default.
- Enable only during DR or staged verification events.
- Keep credentials in secret manager and rotate on normal security cycle.

## Runbook tie-in

- Rollout controls: `LLM_FALLBACK_CANARY_PERCENT`, `LLM_FALLBACK_DISABLED`.
- Rollback/DR verification path is tracked in `llm/PRODUCTION-READINESS.md` and `docs/ops/operations-runbook.md`.
