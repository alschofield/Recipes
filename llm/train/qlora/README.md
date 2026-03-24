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
- `create_training_record.py`: helper to assemble a filled run record from eval artifacts
- `train_qlora.py`: non-interactive QLoRA training entrypoint for commercial SFT splits

## Create a run record

```bash
python llm/train/qlora/create_training_record.py \
  --template llm/train/qlora/training-record-template.json \
  --run-id qlora-pilot-001 \
  --date 2026-03-23 \
  --dataset-name recipes-sft \
  --dataset-version v1 \
  --train-count 7000 \
  --val-count 1500 \
  --test-count 1500 \
  --dataset-hash sha256:replace \
  --adapter-path llm/train/qlora/artifacts/qlora-pilot-001 \
  --logs-path llm/train/qlora/logs/qlora-pilot-001 \
  --eval-output-path llm/evals/results/qlora-pilot-001 \
  --eval-summary llm/evals/results/qlora-pilot-001/safety_complex_first/summary.json \
  --baseline-profile-summary llm/evals/results/qwen3-8b-profile-compare-repair-20260322/profiles_summary.json \
  --baseline-profile safety_complex_first \
  --timeout-rate-percent 0.5 \
  --decision-outcome iterate \
  --decision-notes "initial pilot record" \
  --out llm/train/qlora/runs/qlora-pilot-001.json
```

## Minimum pilot outputs

1. Trained adapter artifact + metadata
2. Eval output under `llm/evals/results/` using the same case files
3. Decision note: promote, iterate, or rollback

## Overnight launch (inside trainer container)

```bash
python -m pip install --upgrade pip
python -m pip install "transformers>=4.45" "trl>=0.12" "peft>=0.12" "accelerate>=0.34" "bitsandbytes>=0.43" sentencepiece

nohup python llm/train/qlora/train_qlora.py \
  --model-name "Qwen/Qwen2.5-7B-Instruct" \
  --output-dir "llm/train/qlora/artifacts/qlora-pilot-001" \
  --max-seq-length 2048 \
  --batch-size 1 \
  --grad-accum 16 \
  --epochs 2 \
  > llm/train/qlora/logs/qlora-pilot-001.train.log 2>&1 &
```
