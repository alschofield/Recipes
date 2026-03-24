# V1 Training Corpus Contract

This contract defines minimum quality gates for large-scale dataset curation used by eval and fine-tuning lanes.

## Scope

- Primary lane: `recipe-query` corpus (ingredient-driven request patterns).
- Optional lane: `sft-json-contract` corpus (system/user/assistant JSON-contract examples).
- Provenance source of truth: `llm/train/datasets/provenance-manifest.v1.json`.

## Required record fields

For `recipe-query` records:

- `id` (stable unique id)
- `lane` (must be `recipe-query`)
- `sourceLane`
- `sourcePath`
- `sourceRecordId`
- `cuisine` (normalized lowercase string, can be empty)
- `ingredients` (normalized, deduped string array)
- `ingredientCount` (integer)
- `queryText` (human-readable request)
- `split` (`train`, `validation`, or `test`)

## Dataset acceptance thresholds (V1)

- Volume:
  - `recipe-query` corpus size >= `50,000` records.
  - Validation split >= `10%` and test split >= `10%`.
- Field integrity:
  - `id`, `sourceLane`, and `queryText` present in `100%` of records.
  - `ingredientCount` equals `len(ingredients)` in `100%` of records.
  - Ingredient lists are deduped and normalized to lowercase ASCII tokens.
- Quality:
  - Duplicate signature rate <= `5%` after dedup pass.
  - Records with ingredient count `<3` or `>30` are excluded by default.
  - At least `20` cuisine labels with >= `100` records each (including `unknown` when unlabeled).
- Provenance and rights:
  - Every source lane appears in `provenance-manifest.v1.json`.
  - Any lane with unresolved legal status is `eval-only` until legal signoff.
  - `approvedForFineTune=true` requires `commercialUseAllowed=true`.

## Dedup policy

- Compute a deterministic signature from normalized cuisine + sorted normalized ingredients.
- Keep first-seen record by source priority and drop subsequent signature collisions.
- Record all drop reasons in curation summary report.

## Split policy

- Deterministic hash-based split by signature.
- Default ratios:
  - Train: `70%`
  - Validation: `15%`
  - Test: `15%`

## Output artifacts

- Curated corpus JSONL (gitignored): `llm/train/datasets/raw/recipe-query-corpus.v1.jsonl`
- Curation summary report: `llm/train/datasets/reports/recipe-query-corpus.v1.summary.json`
- Provenance validation report: `llm/train/datasets/reports/provenance-report.json`

## Commands

```bash
python llm/train/datasets/curate_recipe_query_corpus.py

python llm/train/datasets/validate_provenance.py \
  --manifest llm/train/datasets/provenance-manifest.v1.json \
  --denylist llm/train/datasets/source-denylist.txt \
  --out llm/train/datasets/reports/provenance-report.json
```

## Promotion rule (V1)

- A curated dataset can move from draft to pilot use only when:
  - curation summary meets the acceptance thresholds,
  - provenance validation passes,
  - legal/compliance lane marks source lanes as approved for intended usage.

## Commercial-use guardrail

- For profit/commercial fine-tuning, run `validate_commercial_eligibility.py` and require a `pass` result.
- Only lanes with both `commercialUseAllowed=true` and `approvedForFineTune=true` are eligible.
