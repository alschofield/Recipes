# Ingredient Judge Plan

## Goal

Enrich newly created ingredients with machine-generated metadata without relying on user votes.

## Trigger

- Run judge pass for ingredients created by LLM fallback when metadata fields are missing.

## Candidate models (lightweight)

- Primary: `mistral:latest` (current baseline in runtime).
- Fallback A: `qwen3:4b` (lower footprint option when latency/cost pressure is high).
- Fallback B: hosted OpenAI-compatible small model lane (disaster-recovery path only).

Selection policy:

- Prefer local-first low-cost model when schema and confidence thresholds pass.
- Keep candidate list aligned with `llm/PRODUCTION-READINESS.md` and `docs/llm/serving-options.md`.

## Output fields

- `category` (single primary category)
- `aliasSuggestions` (normalized list)
- `allergenHints` (standard allergen labels when applicable)
- `riskHints` (handling notes: raw poultry, shellfish, etc.)
- `confidence` (0-1)
- `evidence` (short explanation string)

## Policy

- Never overwrite human-approved canonical labels automatically.
- Upsert low-risk fields directly when confidence >= threshold.
- Queue candidate review when confidence is below threshold.

## Validation

- Validate output against `schemas/ingredient-metadata-output.schema.json`.
- Reject and requeue on schema failure.

## Data-driven priors

- Use `llm/judge/data-priors.summary.json` generated from `datasets/raw/server-lib` + `datasets/derived/server-lib`.
- Coverage/quality pattern from canonical seed:
  - coverage `2` mean quality `~0.288`
  - coverage `3` mean quality `~0.355`
  - coverage `4` mean quality `~0.506`
  - coverage `5` mean quality `~0.785`
- Apply stricter auto-approve threshold for low-coverage ingredients.
