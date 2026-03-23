# QLoRA Pilot (Qwen3 8B)

This folder tracks the first fine-tuning pilot for `qwen3:8b` focused on:

- preserving strict JSON schema compliance
- improving safety reliability on cooking-risk prompts
- improving complex-recipe execution without regressing schema

## Scope

- Base model: `Qwen3-8B`
- Method: `QLoRA`
- Objective: supervised fine-tune on contract-aligned recipe outputs
- Hardware target: single 12GB VRAM class GPU for development experiments

## Files

- `pilot-plan.md`: gates, dataset contract, and run flow
- `training-record-template.json`: fill one record per run

## Minimum pilot outputs

1. Trained adapter artifact + metadata
2. Eval output under `llm/evals/results/` using the same case files
3. Decision note: promote, iterate, or rollback
