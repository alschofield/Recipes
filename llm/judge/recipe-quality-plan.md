# Recipe Quality Judge Plan

## Goal

Generate a secondary quality score for LLM recipes using a cheap/light judge model.

## Inputs

- Generated recipe JSON
- Normalized ingredient list used for the request
- Existing deterministic score components (if available)

## Output fields

- `overallScore` (0-1)
- `coherenceScore` (0-1)
- `safetyCompletenessScore` (0-1)
- `techniqueScore` (0-1)
- `confidence` (0-1)
- `notes` (short explanation)

## Policy

- Keep deterministic score as required fallback.
- Do not replace deterministic score if judge output is invalid or low confidence.
- Persist both scores and their provenance for auditability.

## Validation

- Validate output against `schemas/recipe-quality-output.schema.json`.
- Track drift between judge and deterministic scores over time.

## Data-driven priors

- Use `llm/judge/data-priors.summary.json` generated from `datasets/raw/server-lib` + `datasets/derived/server-lib`.
- Kaggle train distribution (`39,774` recipes) indicates ingredient count median `10` and p90 `17`.
- Use these priors to calibrate complexity/coherence expectations in judge scoring.
