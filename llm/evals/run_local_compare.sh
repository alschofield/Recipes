#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:11434/v1}"
API_KEY="${API_KEY:-local-not-used}"
MODEL_A="${MODEL_A:-qwen3:8b}"
MODEL_B="${MODEL_B:-mistral:latest}"

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
RECIPE_CASES="$ROOT_DIR/llm/evals/recipe_cases.jsonl"
SAFETY_CASES="$ROOT_DIR/llm/evals/safety_cases.jsonl"
COMPLEX_CASES="$ROOT_DIR/llm/evals/complex_cases.jsonl"

run_eval() {
  local model="$1"
  local safe_model
  safe_model="$(echo "$model" | tr '/:' '--')"
  local out_dir="$ROOT_DIR/llm/evals/results/$safe_model"

  python "$ROOT_DIR/llm/evals/run_eval.py" \
    --base-url "$BASE_URL" \
    --api-key "$API_KEY" \
    --model "$model" \
    --recipe-cases "$RECIPE_CASES" \
    --safety-cases "$SAFETY_CASES" \
    --complex-cases "$COMPLEX_CASES" \
    --out "$out_dir"
}

run_eval "$MODEL_A"
run_eval "$MODEL_B"

echo "Finished local compare for: $MODEL_A and $MODEL_B"
echo "Results folder: $ROOT_DIR/llm/evals/results"
