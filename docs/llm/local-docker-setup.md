# Local LLM Docker Setup (Threshold Testing)

Use this to run a local LLM endpoint that the recipes server can call while testing fallback thresholds.

## Why this setup

- Keeps integration unchanged: recipes server still calls OpenAI-compatible chat endpoint.
- Keeps cost at zero for local testing.
- Lets you iterate on fallback thresholds without external provider usage.

## 1) Start local model host

From repo root:

```bash
docker compose up -d ollama
```

The root compose config requests GPU access for Ollama (`gpus: all`). If GPU passthrough is unavailable, Ollama falls back to CPU and fallback calls may time out.

## 2) Pull the pinned model

Pinned default for this repo:

```bash
docker exec -it recipes_ollama ollama pull qwen3:8b
```

Optional diagnostic model (not default serving model):

```bash
docker exec -it recipes_ollama ollama pull mistral:latest
```

## 3) Configure server to use local host

In `server/.env` (or runtime env):

```env
LLM_API_KEY=local-not-used
LLM_BASE_URL=http://localhost:11435/v1
LLM_MODEL=qwen3:8b
LLM_PROMPT_PROFILE_DEFAULT=schema_first
LLM_PROMPT_PROFILE_COMPLEX=safety_complex_first
LLM_ENABLE_SAFETY_REPAIR=true
LLM_DISABLE_THINKING_TAG=true
LLM_MAX_TOKENS=1600
LLM_TIMEOUT_SECONDS=180
LLM_REPAIR_TIMEOUT_SECONDS=90
LLM_PROMPT_VERSION=v1-local
LLM_JUDGE_ENABLED=false
LLM_JUDGE_MODEL=mistral:latest
LLM_JUDGE_MIN_CONFIDENCE=0.65
```

Notes:

- `LLM_API_KEY` must be non-empty due to current backend validation.
- When running recipes-server inside Docker, use `http://ollama:11434/v1` instead of localhost.
- Root compose publishes Ollama on `127.0.0.1:11435` by default to avoid host-port collisions.
- On 12GB VRAM, `qwen3:8b` is the current quality/latency sweet spot for this app.
- Runtime strategy: use `schema_first` as default and `safety_complex_first` for complex requests (10+ ingredients / larger servings / high-time filters).
- Client can force complex profile by sending `"complex": true` in `/recipes/search` request body.
- Safety repair path is enabled by `LLM_ENABLE_SAFETY_REPAIR`; backend attempts one repair call before giving up.
- `LLM_MAX_TOKENS` keeps generation bounded for latency stability.
- Optional judge path can enrich ingredient metadata + secondary quality scoring when `LLM_JUDGE_ENABLED=true`.

## 4) Quick health checks

List local models:

```bash
curl http://localhost:11435/api/tags
```

OpenAI-compatible chat check:

```bash
curl http://localhost:11435/v1/chat/completions \
  -H "Authorization: Bearer local-not-used" \
  -H "Content-Type: application/json" \
  -d '{"model":"qwen3:8b","messages":[{"role":"user","content":"Return {\"ok\":true} as JSON only."}]}'
```

Recipes-server fallback health + metrics check:

```bash
curl http://localhost/api/recipes/health/llm
```

LLM/judge metrics only:

```bash
curl http://localhost/api/recipes/metrics/llm
```

If you are bypassing nginx and hitting recipes-server directly:

```bash
curl http://localhost:8081/recipes/health/llm
```

## 5) Threshold testing flow

1. Set threshold values in backend search/fallback code.
2. Run recipe search scenarios (empty DB, low-confidence, dbOnly=true).
3. Confirm fallback trigger behavior and output schema validity.
4. Capture latency and schema-pass observations in `eval-scorecard-template.md`.

## 6) Stop and cleanup

```bash
docker compose stop ollama
```

To remove downloaded model data too:

```bash
docker compose down -v
```
