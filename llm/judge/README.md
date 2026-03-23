# Judge Model Workspace

This folder contains plans and contracts for lightweight judge-model usage.

## Purpose

- Infer ingredient metadata for newly created canonical ingredients.
- Produce a secondary recipe quality score independent of user voting.
- Keep deterministic server scoring as fallback when judge model is unavailable.

## Initial scope

- Ingredient metadata enrichment (category, alias hints, allergen/risk hints, confidence).
- Recipe quality rubric scoring (technique quality, ingredient coherence, safety completeness).

## Files

- `ingredient-metadata-plan.md`
- `recipe-quality-plan.md`
- `data-patterns-from-server-lib.md`
- `data-priors.summary.json`
- `prompts/ingredient-metadata.prompt.txt`
- `prompts/recipe-quality.prompt.txt`
- `schemas/ingredient-metadata-output.schema.json`
- `schemas/recipe-quality-output.schema.json`
- `calibration-template.json`
