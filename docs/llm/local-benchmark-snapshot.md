# Local LLM Benchmark Snapshot (This Machine)

This captures real local runs against the eval harness in `llm/evals/`.

## Environment

- Host GPU: `NVIDIA GeForce RTX 3080 Ti` (12 GB VRAM)
- Current decision: pin serving model to `qwen3:8b`
- Practical VRAM guidance on this machine: `<=12B` class models are realistic; `8B` is the most reliable quality/latency tradeoff so far.

## Eval Results (Baseline Harness)

Using:

- `llm/evals/recipe_cases.jsonl` (5 cases)
- `llm/evals/safety_cases.jsonl` (4 cases)
- `llm/evals/run_eval.py`

| Model | Schema pass % | Safety pass % | Avg latency (ms) | P95 latency (ms) | Notes |
|---|---:|---:|---:|---:|---|
| `qwen3:8b` (`safety_complex_first` + safety repair, profile compare run) | 91.67 | 100.00 | 56,210.46 | 105,376.77 | Best weighted score in current profile compare; one complex timeout still present |
| `qwen3:8b` (`schema_first` + safety repair, profile compare run) | 100.00 | 100.00 | 53,031.78 | 70,624.99 | Best schema/safety reliability, but complex pass remains 0% |
| `qwen3:8b` (`--disable-thinking-tag`, timeout 120s, tuned prompt v2) | 92.31 | 75.00 | 63,398.58 | 102,077.59 | Best current safety + complex balance, but schema/p95 regress from baseline |
| `qwen3:8b` (`--disable-thinking-tag`, timeout 120s) | 100.00 | 33.33 | 51,635.69 | 92,317.42 | Pinned model; schema strong, safety/complexity need prompt/data tuning |
| `qwen3:4b` (`--disable-thinking-tag`) | 0.00 | 0.00 | 60,013.99 | 60,028.39 | Every case timed out at the evaluator's 60s timeout |
| `qwen2.5:latest` | 100.00 | 50.00 | 14,646.73 | 27,109.94 | Historical strong baseline before Qwen3 decision |
| `mistral:latest` | 66.67 | 50.00 | 23,788.65 | 31,993.04 | Diagnostic comparison model |

## Direct Prompt Smoke Tests (OpenAI-compatible endpoint)

- `qwen3:4b` on full recipe prompt: timed out at `180s`.
- `qwen3:8b` on same prompt: completed in about `42.8s`.
- `mistral:latest` on same prompt: completed in about `22.3s`.

## Artifact Paths (Current Focus)

- `llm/evals/results/qwen3-8b-with-complex-timeout120/summary.json`
- `llm/evals/results/qwen3-8b-with-complex-timeout120/details.jsonl`
- `llm/evals/results/qwen3-8b-with-complex-timeout120-tuned-prompt-v2/summary.json`
- `llm/evals/results/qwen3-8b-with-complex-timeout120-tuned-prompt-v2/details.jsonl`
- `llm/evals/results/qwen3-8b-profile-compare-repair-20260322/profiles_summary.json`
- `llm/evals/results/qwen3-8b-profile-compare-repair-20260322/schema_first/summary.json`
- `llm/evals/results/qwen3-8b-profile-compare-repair-20260322/safety_complex_first/summary.json`
- `llm/evals/results/qwen3-4b-no-think-20260322/summary.json`
- `llm/evals/results/qwen3-4b-no-think-20260322/details.jsonl`
- `llm/evals/results/qwen2.5-latest-timeout120/summary.json`
- `llm/evals/results/mistral-latest-timeout120/summary.json`

## Interpretation

- `qwen3:4b` is currently unreliable for full recipe prompts in this stack due to timeout behavior.
- `qwen3:8b` is the pinned serving direction despite being larger, because it returns full prompts where `qwen3:4b` stalls.
- New dual-profile run: `safety_complex_first` wins weighted score, while `schema_first` remains the most stable schema profile.
- Safety repair pass is effective for recoverable safety-case failures in `schema_first` profile.
- Baseline `qwen3:8b` meets schema but misses safety/complexity gates.
- Tuned-prompt v2 improves safety (`75%`) and complexity (`75%`) but drops schema (`92.31%`) and increases p95 latency (`102.1s`).
- `mistral:latest` remains useful for benchmarking, but it is not the selected default.

## Immediate Next Runs

1. Add complex-case retry policy for timed-out cases and measure timeout-rate change.
2. Build initial QLoRA training set from schema/safety/complex hard examples.
3. Run first QLoRA pilot and compare against `qwen3-8b-profile-compare-repair-20260322` baseline.
