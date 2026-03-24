# DB-First + LLM Fallback Contract

Status: v1 locked baseline (update via checklist preflight item)

## Decision

- Primary source is database recipes.
- LLM is fallback only when DB results are low-confidence or empty.

## Fallback Trigger Rules

Trigger LLM generation when either is true:

1. DB returns `0` results.
2. Top DB `matchPercent` is below `0.45` and fewer than `5` results are returned.

Do not trigger fallback if user sets `dbOnly=true`.

## LLM Input Contract

- Normalized ingredient list.
- Search mode (`strict` or `inclusive`).
- Constraints: dietary tags, max prep time, cuisine, servings.
- Optional client hint: `complex=true` to force complex prompt profile.
- Required output schema (JSON only).
- Prompt profile strategy:
  - `schema_first` for default requests.
  - `safety_complex_first` for complex requests.
  - Complex mode auto-enables for >=10 normalized ingredients, even without explicit hint.
- Runtime controls:
  - `LLM_FALLBACK_DISABLED=true` fully disables fallback generation for incident handling.
  - `LLM_FALLBACK_CANARY_PERCENT` limits fallback execution by deterministic canary percentage.
  - `LLM_STRICT_GENERATED_POLICY` controls strict-mode behavior (`none` or `degrade_inclusive`).
  - `LLM_DISABLE_THINKING_TAG=true` prefixes `/no_think` for Qwen3 latency stability.
  - `LLM_MAX_TOKENS` bounds response length for fallback predictability.
  - `LLM_TIMEOUT_SECONDS` and `LLM_REPAIR_TIMEOUT_SECONDS` tune fallback + repair deadlines.
  - Optional judge model controls (`LLM_JUDGE_*`) manage metadata/secondary score enrichment.

## LLM Output Schema (Required)

```json
{
  "recipes": [
    {
      "name": "string",
      "description": "string",
      "ingredients": [
        { "name": "string", "amount": "string", "optional": false }
      ],
      "steps": ["string"],
      "prepMinutes": 30,
      "cookMinutes": 20,
      "difficulty": "easy",
      "cuisine": "string",
      "dietaryTags": ["string"],
      "servings": 2,
      "safetyNotes": ["string"]
    }
  ]
}
```

## Validation and Persistence

- Reject non-JSON or schema-invalid responses.
- Optional safety-repair pass can re-ask the model once to repair schema-invalid output for safety-sensitive prompts.
- Normalize ingredient names before persistence.
- Persist generated recipes with:
  - `source=llm`
  - `generationModel`
  - `generationTimestamp`
  - `promptVersion`
- Mark generated recipes as `reviewable=true`.
- For newly created ingredients, optional judge pass can enrich category/alias/risk metadata.
- Recipe quality uses deterministic scoring first; optional judge score can blend in with confidence gating.

## Response Composition

- Blend DB matches and generated results with deterministic interleave when fallback returns usable generated items.
- Blend policy is controlled by `SEARCH_BLEND_MIN_GENERATED`, `SEARCH_BLEND_MAX_GENERATED_SHARE`, and `SEARCH_BLEND_SEED`.
- Include `source` per recipe (`database` or `llm`).
- Include blend metadata per recipe (`blendSlot`, `rankingReason`).
- Include explanation metadata (`matchedIngredients`, `missingIngredients`).

## Caching Policy

- Cache by normalized ingredient hash + mode + filters.
- Default TTL: 6 hours.
- Bypass cache when `debugNoCache=true` (non-production only).

## Safety Rules

- Block unsafe cooking instructions or dangerous ingredient pairings when detected.
- Include allergy-sensitive reminders when common allergens appear.
- Never claim medical/nutritional certainty.

## Testing Expectations

- Mock provider responses for deterministic tests.
- Validate fallback trigger edge cases.
- Validate schema rejection and recovery path.
- Ensure provenance fields are always present for generated recipes.

## Health and Metrics

- Health endpoint: `GET /recipes/health/llm` (nginx path: `GET /api/recipes/health/llm`).
- Includes runtime config snapshot, fallback counters (`requests`, `success`, `requestErrors`, `timeoutErrors`, `schemaErrors`, `repairsTried`, `repairsOK`), and alert-rate status.
- Metrics endpoint: `GET /recipes/metrics/llm` (nginx path: `GET /api/recipes/metrics/llm`).
- Base service metrics endpoint `GET /recipes/metrics` includes labeled LLM fallback counter samples under `llmFallbackMetrics.samples`.
- Alert thresholds are env-tunable with `LLM_ALERT_TIMEOUT_RATE_THRESHOLD`, `LLM_ALERT_SCHEMA_ERROR_RATE_THRESHOLD`, and `LLM_ALERT_REPAIR_FAIL_RATE_THRESHOLD`.
