# LLM Serving Options (Prod-Oriented)

This document compares serving approaches. For concrete self-host setup details, use `serving-infra-blueprints.md`.

## Recommended Default

- Use Python-first tooling for training/evals.
- Keep Go backend integration via OpenAI-compatible API.
- Current project direction: self-hosted first (`qwen3:8b`) to avoid per-request billing.

## Option A: Hosted OpenAI-Compatible Provider

- Fastest to launch.
- Works with current `LLM_BASE_URL` + `LLM_MODEL` env model.
- Best for low-ops early stages.

## Option B: Self-Hosted vLLM (Python)

- Good production path for GPU throughput and batching.
- Supports OpenAI-compatible serving APIs.
- Best when request volume justifies infra ownership.

## Option C: Self-Hosted llama.cpp (C/C++)

- Strong for local/dev, edge, or smaller-model inference.
- Lower infra overhead for CPU-focused deployments.
- Not a replacement for Python ecosystem in training/fine-tuning.

## Suggested Sequence

1. Local Docker + Ollama baseline for product/contract iteration.
2. Build eval dataset and scorecard (`eval-scorecard-template.md`).
3. Move to vLLM self-host for production throughput/observability.
4. Keep hosted provider as contingency path and llama.cpp for low-footprint edge cases.
