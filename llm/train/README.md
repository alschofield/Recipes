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
- `datasets/dedup_recipes_jsonl.py` - removes near-identical training rows.
- `datasets/check_safety_policy_jsonl.py` - blocks unsafe instruction phrasing.

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
```

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
