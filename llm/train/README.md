# LLM Training Workspace

This folder is for model training and adapter lifecycle work.

## Structure

- `qlora/` - QLoRA pilot plans, run records, and adapter metadata.
- `datasets/` - versioned training/validation/test manifests (add as created).
- `artifacts/` - local adapter outputs (gitignored in normal workflow).

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
