# QLoRA Pilot Plan - Qwen3 8B

## Goal

Improve safety + complex pass rates while keeping schema validity high for the existing recipes contract.

## Baseline reference

Use the latest profile compare outputs:

- `llm/evals/results/qwen3-8b-profile-compare-repair-20260322/profiles_summary.json`

## Data contract for SFT examples

Each training example should contain:

- `system`: JSON-only contract role text
- `user`: prompt in the same shape as `llm/evals/run_eval.py` prompt builder
- `assistant`: valid JSON object matching schema in `docs/llm/fallback-contract.md`

Required labeling buckets:

- `schema_hard`: strict field and type compliance
- `safety_hard`: chicken temp, kidney bean boil, allergen notes, storage notes
- `complex_hard`: multi-recipe sequencing and technique diversity

## Suggested first-pass split

- Train: 70%
- Validation: 15%
- Test: 15%

Keep splits deterministic (seeded), and track hash in run metadata.

## QLoRA starter hyperparameters (initial)

- LoRA rank: `16`
- LoRA alpha: `32`
- LoRA dropout: `0.05`
- Sequence length: `4096`
- Effective batch size: `16` (microbatch + grad accumulation)
- Learning rate: `2e-4`
- Scheduler: cosine
- Epochs: `2`
- Weight decay: `0.01`

Tune one variable at a time after the first run.

## Run flow

1. Prepare dataset with provenance metadata.
2. Train adapter on Qwen3 8B with QLoRA.
3. Export adapter + run metadata record.
4. Re-run eval harness:
   - `schema_first`
   - `safety_complex_first`
   - with `--enable-safety-repair`
5. Compare with baseline and decide promote/iterate.

## Promotion gates for pilot success

- Schema pass >= `95%`
- Safety pass >= `99%`
- Complex pass >= `70%`
- No increase in request-timeout rate

## Rollback criteria

- Schema pass drops below baseline by >2 points
- Safety pass does not improve after 2 tuning cycles
- Latency regression >20% p95 with no quality gain

## Tracking

Fill one record per run in `training-record-template.json`.
