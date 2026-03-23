# LLM Evals

This folder provides a lightweight evaluation harness for OpenAI-compatible chat endpoints.

Use it to compare candidate models for recipe generation before changing defaults.

## Files

- `recipe_cases.jsonl`: baseline recipe-generation test cases.
- `safety_cases.jsonl`: targeted safety-focused prompts.
- `complex_cases.jsonl`: high-complexity recipe prompts with structure thresholds.
- `run_eval.py`: schema/safety/latency scorer.

## Quick start

From repo root:

If using root Docker compose defaults, Ollama is on `http://localhost:11435/v1`.

```bash
python llm/evals/run_eval.py \
  --base-url http://localhost:11435/v1 \
  --api-key local-not-used \
  --model qwen3:8b \
  --prompt-profile schema_first \
  --enable-safety-repair \
  --recipe-cases llm/evals/recipe_cases.jsonl \
  --safety-cases llm/evals/safety_cases.jsonl \
  --complex-cases llm/evals/complex_cases.jsonl \
  --out llm/evals/results/qwen3-8b
```

Run both prompt profiles and auto-rank them:

```bash
python llm/evals/run_eval.py \
  --base-url http://localhost:11435/v1 \
  --api-key local-not-used \
  --model qwen3:8b \
  --compare-profiles \
  --enable-safety-repair \
  --recipe-cases llm/evals/recipe_cases.jsonl \
  --safety-cases llm/evals/safety_cases.jsonl \
  --complex-cases llm/evals/complex_cases.jsonl \
  --out llm/evals/results/qwen3-8b-profile-compare
```

## Low-cost local compare (no hosted API billing)

Run two local models back-to-back:

```powershell
./llm/evals/run_local_compare.ps1
```

```bash
bash llm/evals/run_local_compare.sh
```

Override model names when needed:

```powershell
./llm/evals/run_local_compare.ps1 -ModelA "qwen3:8b" -ModelB "mistral:latest"
```

```bash
MODEL_A="qwen3:8b" MODEL_B="mistral:latest" bash llm/evals/run_local_compare.sh
```

## Outputs

Each run writes:

- `summary.json`: aggregate metrics (schema/safety/complexity pass, latency).
- `details.jsonl`: per-case results.
- `scorecard.md`: markdown snapshot for human review.
- `profiles_summary.json`: profile ranking output (when `--compare-profiles` is used).

## Notes

- This is a baseline evaluator, not a full benchmark suite.
- It is designed to match the app's JSON contract closely.
- Keep all candidate runs on the same case files for fair comparison.
- Local OpenAI-compatible endpoints (like Ollama) avoid per-request provider fees.
- Current pinned serving candidate is `qwen3:8b`.
- The evaluator prints per-case progress so long runs are observable.
- Prompt profiles available: `schema_first`, `safety_complex_first`.
- Optional safety repair pass re-calls the same model once for failed safety cases.
