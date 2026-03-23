# LLM Serving Infra Blueprints

This document captures concrete infrastructure blueprints for production-oriented deployment tracks.

## Blueprint: Self-Hosted LLM Inference (No Per-Request Provider Fees)

Goal: run recipe-generation models on your own infrastructure while keeping current app integration stable.

### 1) Architecture

- `recipes-server` keeps calling an OpenAI-compatible endpoint (`LLM_BASE_URL`, `LLM_MODEL`, `LLM_API_KEY`).
- LLM serving runs on a separate inference service (recommended: `vLLM`).
- Fine-tuning runs as a separate pipeline/job (not on the inference node).
- Model artifacts are stored in object storage or a private model registry.

### 2) Environment Layout

- **Local dev**: Ollama in the root `docker-compose.yml` (`docker compose up -d ollama`) for fast iteration.
- **Staging**: 1 GPU inference node + private model artifact storage.
- **Production**: at least 2 inference replicas behind a load balancer for HA.

### 3) Recommended Runtime Stack

- **Training/fine-tuning**: Python (QLoRA first).
- **Inference serving**: vLLM with OpenAI-compatible API.
- **Optional low-footprint path**: `llama.cpp` for small models/edge use.

### 4) Initial Capacity Guidance

For 7B-class instruct models:

- **Staging**: 1x 24GB GPU class (L4/A10G equivalent), 4-8 vCPU, 16-32GB RAM.
- **Production baseline**: 2x staging-equivalent nodes minimum.
- Use quantized variants first for cost control, then increase quality tier if needed.

### 5) vLLM Docker Blueprint

Example container:

```bash
docker run --gpus all --rm -p 8000:8000 \
  -e HUGGING_FACE_HUB_TOKEN=$HUGGING_FACE_HUB_TOKEN \
  -v vllm_cache:/root/.cache/huggingface \
  vllm/vllm-openai:latest \
  --model Qwen/Qwen3-8B \
  --host 0.0.0.0 --port 8000
```

Server wiring in `recipes-server` env:

```env
LLM_API_KEY=internal-token
LLM_BASE_URL=https://llm.yourdomain.com/v1
LLM_MODEL=Qwen/Qwen3-8B
```

### 6) Reliability and Safety Controls

- Keep response schema validation in `recipes-server` (already implemented).
- Add request timeout and retry policy only for idempotent fallback calls.
- Add rate limiting on LLM endpoint.
- Keep at least one rollback model ready (`LLM_MODEL` swap, no code change).
- Run canary rollout for new model versions before full traffic shift.

### 7) Operations Checklist

- [ ] Health endpoints and uptime checks for inference service
- [ ] Metrics: p50/p95 latency, tokens/sec, error rate, GPU memory, queue depth
- [ ] Centralized logs with request IDs
- [ ] Alerting thresholds for latency/error spikes
- [ ] Backup/restore plan for model artifacts and config

### 8) Cost Reality Check

- You avoid per-request vendor billing, but you pay fixed infra costs.
- Self-hosting is usually best when traffic is steady enough to amortize GPU uptime.
- Keep a simple hosted-provider fallback ready for incident recovery.
