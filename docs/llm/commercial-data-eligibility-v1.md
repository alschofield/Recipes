# Commercial Data Eligibility (V1)

This document defines which dataset lanes are currently eligible for commercial fine-tuning and profit-oriented product usage.

Important: this is an operational policy document, not legal advice. Final legal decisions must be confirmed by formal review.

## Current decision table

| Source lane id | Source type | Commercial fine-tune | Decision |
|---|---|---|---|
| `canonical-seed-v1` | first-party-derived | Yes | Allowed |
| `internal-synthetic-query-v1` | first-party-derived | Yes | Allowed |
| `kaggle-train-json` | third-party-public | No | Blocked pending legal approval |
| `culinarydb-aliases` | third-party-public | No | Blocked pending legal approval |
| `world-recipes-csv` | third-party-public | No | Blocked pending legal approval |

## Enforcement policy

- Training corpora for commercial use must pass `validate_commercial_eligibility.py`.
- A source lane is commercial-eligible only when both manifest flags are true:
  - `commercialUseAllowed`
  - `approvedForFineTune`
- Any dataset row with unknown `sourceLane` is blocked.
- Any dataset row from a non-approved lane is blocked.

## Required operator commands

```bash
python llm/train/datasets/curate_commercial_recipe_query_corpus.py

python llm/train/datasets/validate_commercial_eligibility.py \
  --manifest llm/train/datasets/provenance-manifest.v1.json \
  --in llm/train/datasets/raw/commercial-recipe-query-corpus.v1.jsonl \
  --out llm/train/datasets/reports/commercial-eligibility-report.v1.json

python llm/train/datasets/build_commercial_sft_contract_dataset.py

python llm/train/datasets/validate_commercial_eligibility.py \
  --manifest llm/train/datasets/provenance-manifest.v1.json \
  --in llm/train/datasets/processed/commercial-sft-json-contract.v1.train.jsonl \
  --out llm/train/datasets/reports/commercial-eligibility-sft-train.v1.json
```

## Notes

- The commercial-safe corpus intentionally excludes third-party recipe text.
- It is generated from internal canonical ingredient seed data and synthetic query templates.
- If legal approves a third-party lane later, update the manifest flags first, then rerun eligibility validation.
