# Checklist Completed Items Archive

Archived on 2026-03-23.

## Source: `server/CHECKLIST.md`

### Fallback Runtime and Reliability

- Add runtime prompt profile routing (`schema_first` / `safety_complex_first`).
- Add client complex hint support and auto-complex threshold (>=10 ingredients).
- Add one-pass safety repair flow for recoverable schema issues.
- Add runtime `/no_think` control and bounded token output (`LLM_DISABLE_THINKING_TAG`, `LLM_MAX_TOKENS`).

### Observability and Operations

- Add `GET /recipes/health/llm` endpoint exposing runtime config and fallback counters.
- Add machine-friendly LLM metrics endpoint (`GET /recipes/metrics/llm`).

### Ingredient Governance and Metadata

- Resolve or create missing ingredients during LLM recipe persistence.
- Add background judge-triggered enrichment for newly created ingredients (category/coverage/quality/metadata when enabled).

### Judge Model (Non-User-Derived Quality/Metadata)

- Add lightweight judge-model pass for ingredient metadata inference on newly created ingredients.
- Add lightweight judge-model pass for secondary recipe quality score (alongside deterministic score).
- Persist judge outputs with confidence + trace fields and keep deterministic score as fallback.
- Derive initial metadata/quality calibration priors from `datasets/raw/server-lib` and `datasets/derived/server-lib`.

## Source: `web/CHECKLIST.md`

### Search UX and Controls

- Send `complex` hint from recipes search page payload.
- Auto-enable complex mode in UI when 10+ ingredients are entered.

## Source: `llm/CHECKLIST.md`

### Judge Model and Metadata Enrichment (V1+)

- Add judge pass for secondary recipe quality scoring (non-user-derived), persisted alongside deterministic score.
- Keep deterministic non-LLM score path as mandatory fallback when judge model is unavailable.
