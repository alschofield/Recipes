# LLM Workspace

This folder is organized by operational lane so serving, evaluation, training, and judge-model work stay compartmentalized.

## Structure

- `evals/` - evaluation datasets, harness, and result artifacts.
- `train/` - training plans, run records, and trainer-compose setup.
- `judge/` - lightweight judge-model plans, prompts, schemas, and data priors.
- `CHECKLIST.md` - LLM source-of-truth backlog.
- `PRODUCTION-READINESS.md` - hard gates before production switch.

Reference contract docs:

- `CHECKLIST.md`
- `../docs/llm/fallback-contract.md`
- `../docs/llm/local-docker-setup.md`
- `../docs/llm/finetune-model-shortlist.md`
- `../docs/llm/eval-scorecard-template.md`
- `../docs/llm/data-provenance-manifest-template.md`
- `../docs/llm/serving-options.md`
- `../docs/llm/serving-infra-blueprints.md`
- `train/qlora/pilot-plan.md`
- `judge/ingredient-metadata-plan.md`
- `judge/recipe-quality-plan.md`
- `judge/README.md`
- `PRODUCTION-READINESS.md`

Local quick start:

- `docker compose up -d ollama`
- `docker exec -it recipes_ollama ollama pull qwen3:8b`
- Then follow `../docs/llm/local-docker-setup.md`

Training container quick start:

- `docker compose -f llm/train/docker-compose.train.yml up -d`
- `docker exec -it recipes_llm_trainer bash`
