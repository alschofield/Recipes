# LLM Training Workspace

This folder is for model training and adapter lifecycle work.

## Structure

- `qlora/` - QLoRA pilot plans, run records, and adapter metadata.
- `datasets/` - versioned training/validation/test manifests (add as created).
- `artifacts/` - local adapter outputs (gitignored in normal workflow).

## Dataset controls

- `datasets/provenance-manifest.v1.json` - source/license/provenance registry.
- `datasets/source-denylist.txt` - blocked sources for training use.
- `datasets/validate_provenance.py` - validates manifest + denylist constraints.
- `datasets/validate_commercial_eligibility.py` - blocks non-commercial/non-approved source lanes for profit use.
- `datasets/dedup_recipes_jsonl.py` - removes near-identical training rows.
- `datasets/check_safety_policy_jsonl.py` - blocks unsafe instruction phrasing.
- `datasets/curate_recipe_query_corpus.py` - builds a large deduped query corpus from raw lanes.
- `datasets/curate_commercial_recipe_query_corpus.py` - builds a large commercial-safe query corpus from first-party seed data.
- `datasets/build_commercial_sft_contract_dataset.py` - converts commercial query corpus into system/user/assistant JSON-contract SFT splits.

Example commands:

```bash
python llm/train/datasets/validate_provenance.py \
  --manifest llm/train/datasets/provenance-manifest.v1.json \
  --denylist llm/train/datasets/source-denylist.txt \
  --out llm/train/datasets/reports/provenance-report.json

python llm/train/datasets/dedup_recipes_jsonl.py \
  --in llm/train/datasets/raw/train.jsonl \
  --out llm/train/datasets/processed/train.dedup.jsonl \
  --report llm/train/datasets/reports/train.dedup.report.json

python llm/train/datasets/check_safety_policy_jsonl.py \
  --in llm/train/datasets/processed/train.dedup.jsonl \
  --out llm/train/datasets/reports/train.safety.report.json

python llm/train/datasets/curate_recipe_query_corpus.py \
  --out-jsonl llm/train/datasets/raw/recipe-query-corpus.v1.jsonl \
  --out-report llm/train/datasets/reports/recipe-query-corpus.v1.summary.json

python llm/train/datasets/curate_commercial_recipe_query_corpus.py \
  --out-jsonl llm/train/datasets/raw/commercial-recipe-query-corpus.v1.jsonl \
  --out-report llm/train/datasets/reports/commercial-recipe-query-corpus.v1.summary.json

python llm/train/datasets/validate_commercial_eligibility.py \
  --manifest llm/train/datasets/provenance-manifest.v1.json \
  --in llm/train/datasets/raw/commercial-recipe-query-corpus.v1.jsonl \
  --out llm/train/datasets/reports/commercial-eligibility-report.v1.json

python llm/train/datasets/build_commercial_sft_contract_dataset.py \
  --in-jsonl llm/train/datasets/raw/commercial-recipe-query-corpus.v1.jsonl \
  --out-train llm/train/datasets/processed/commercial-sft-json-contract.v1.train.jsonl \
  --out-validation llm/train/datasets/processed/commercial-sft-json-contract.v1.validation.jsonl \
  --out-test llm/train/datasets/processed/commercial-sft-json-contract.v1.test.jsonl \
  --out-report llm/train/datasets/reports/commercial-sft-json-contract.v1.summary.json
```

Contract reference: `docs/llm/v1-training-corpus-contract.md`.

## Runtime model

- Run training in a dedicated container/job, separate from serving containers.
- Keep training data and artifacts on mounted volumes.
- Evaluate every run with `llm/evals/` before promotion.

## Quick start

From repo root:

```bash
docker compose -f llm/train/docker-compose.train.yml up -d
docker exec -it recipes_llm_trainer bash
```
