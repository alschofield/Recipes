# LLM Model Stats Matrix (Shortlist + Recognizable Options)

Use this matrix to compare base-model candidates and plug in eval outputs from `llm/evals/`.

Notes:

- "Recognizability" is a product/marketing proxy, not a quality metric.
- API pricing and limits change often; re-verify before final decisions.
- For open models, self-host cost is infra-based (GPU/runtime), not token-billed API.

## Candidate + Recognizable Model Table

| Model | Type | Recognizability | Params | Context | Self-host | Fine-tune path | License/commercial posture | Cost model | Est. single-GPU inference footprint | Eval schema pass % | Eval safety pass % | Eval complex pass % | Eval p95 latency ms |
|---|---|---|---:|---:|---|---|---|---|---|---:|---:|---:|---:|
| `Qwen/Qwen3-8B` (Ollama `qwen3:8b`) | Open-weight (pinned default) | Medium | 8B | (verify runtime context) | Yes | Yes (LoRA/QLoRA) | Model-specific (verify) | Infra cost | ~8-12GB VRAM (4-bit), ~18-24GB (BF16) | 91.67 | 100.00 | 75.00 | 105376.77 |
| `Qwen/Qwen2.5-7B-Instruct` | Open-weight (historical baseline) | Medium | 7.61B | 131k | Yes | Yes (LoRA/QLoRA) | Apache-2.0 (commercial-friendly) | Infra cost | ~6-8GB VRAM (4-bit), ~16-20GB (BF16) | 100.00 | 50.00 | 0.00 | 29811.54 |
| `Qwen/Qwen2.5-14B-Instruct` | Open-weight | Medium | ~14B | 131k | Yes | Yes (LoRA/QLoRA) | Apache-2.0 | Infra cost | ~12-16GB VRAM (4-bit), ~32-40GB (BF16) | TBD | TBD | TBD | TBD |
| `Qwen/Qwen2.5-32B-Instruct` | Open-weight | Medium | ~32B | 131k | Yes | Yes (LoRA/QLoRA) | Apache-2.0 | Infra cost | ~24-40GB VRAM (4-bit), multi-GPU BF16 | TBD | TBD | TBD | TBD |
| `Qwen/Qwen3-4B-Instruct` (Ollama `qwen3:4b`) | Open-weight | Medium | 4B | (verify runtime context) | Yes | Yes (LoRA/QLoRA) | Model-specific (verify) | Infra cost | ~4-6GB VRAM (4-bit), ~10-14GB (BF16) | 0.00 | 0.00 | 0.00 | 60028.39 |
| `mistralai/Mistral-7B-Instruct-v0.3` | Open-weight (picked) | Medium | 7B | (verify in runtime; commonly 32k-class) | Yes | Yes (LoRA/QLoRA) | Apache-2.0 (commercial-friendly) | Infra cost | ~6-8GB VRAM (4-bit), ~16-20GB (BF16) | 76.92 | 50.00 | 0.00 | 46404.18 |
| `mistralai/Mixtral-8x7B-Instruct-v0.1` | Open-weight | Medium | MoE 8x7B | (verify in runtime) | Yes | Yes (LoRA path more complex) | Apache-2.0 | Infra cost | Higher memory/throughput tuning required | TBD | TBD | TBD | TBD |
| `mistralai/Mistral-Nemo-Instruct-2407` | Open-weight | Medium | ~12B | 128k | Yes | Yes (LoRA/QLoRA) | Apache-2.0 | Infra cost | ~10-14GB VRAM (4-bit), ~28-36GB (BF16) | TBD | TBD | TBD | TBD |
| `google/gemma-2-9b-it` | Open-weight | Medium | 9B | (verify current card) | Yes | Yes (LoRA/QLoRA) | Gemma terms (commercial use allowed with terms) | Infra cost | ~8-12GB VRAM (4-bit), ~20-24GB (BF16) | TBD | TBD | TBD | TBD |
| `google/gemma-2-27b-it` | Open-weight | Medium | 27B | (verify current card) | Yes | Yes (LoRA/QLoRA) | Gemma terms (commercial use allowed with terms) | Infra cost | ~20-32GB VRAM (4-bit), multi-GPU BF16 | TBD | TBD | TBD | TBD |
| `meta-llama/Llama-3.1-8B-Instruct` | Open-weight (picked stretch) | High | 8B | 128k | Yes | Yes (with license terms) | Llama 3.1 Community License (commercial allowed with conditions) | Infra cost | ~8-10GB VRAM (4-bit), ~18-22GB (BF16) | 87.50 | 33.33 | TBD | 118658.52 |
| `google/gemma-3-4b-it` (Ollama `gemma3:latest`) | Open-weight | Medium | 4B | (verify current card) | Yes | Yes (LoRA/QLoRA) | Gemma terms | Infra cost | ~4-6GB VRAM (4-bit), ~10-14GB (BF16) | 92.31 | 25.00 | 50.00 | 43555.93 |
| `meta-llama/Llama-3-8B-Instruct` (Ollama `llama3:latest`) | Open-weight | High | 8B | (verify current card) | Yes | Yes (with license terms) | Llama terms | Infra cost | ~8-10GB VRAM (4-bit), ~18-22GB (BF16) | 88.89 | 0.00 | TBD | 58662.93 |
| `meta-llama/Llama-3.2-3B-Instruct` (Ollama `llama3.2:latest`) | Open-weight | High | 3B | 128k | Yes | Yes (with license terms) | Llama terms | Infra cost | ~3-5GB VRAM (4-bit), ~8-10GB (BF16) | 33.33 | 50.00 | TBD | 84799.10 |
| `meta-llama/Llama-3.1-70B-Instruct` | Open-weight | High | 70B | 128k | Yes | Yes (with license terms) | Llama 3.1 Community License | Infra cost | Multi-GPU required in most production setups | TBD | TBD | TBD | TBD |
| `microsoft/Phi-4` | Open-weight | Medium | ~14B | (verify current card) | Yes | Yes (LoRA/QLoRA) | MIT-style / model-specific terms (verify exact card) | Infra cost | ~12-16GB VRAM (4-bit), ~32-40GB (BF16) | TBD | TBD | TBD | TBD |
| `gpt-5.4` | Hosted API | Very high | N/A (closed) | 1M | No | Provider-managed only | Proprietary API terms | Token pricing (`$2.50` in / `$15` out per MTok, as listed) | N/A | TBD | TBD | TBD | TBD |
| `gpt-5.4-mini` | Hosted API | Very high | N/A (closed) | 400k | No | Provider-managed only | Proprietary API terms | Token pricing (`$0.75` in / `$4.50` out per MTok, as listed) | N/A | TBD | TBD | TBD | TBD |
| `gpt-5.4-nano` | Hosted API | High | N/A (closed) | 400k | No | Provider-managed only | Proprietary API terms | Token pricing (`$0.20` in / `$1.25` out per MTok, as listed) | N/A | TBD | TBD | TBD | TBD |
| `claude-opus-4-6` | Hosted API | High | N/A (closed) | 1M | No | Provider-managed only | Proprietary API terms | Token pricing (`$5` in / `$25` out per MTok, as listed) | N/A | TBD | TBD | TBD | TBD |
| `claude-sonnet-4-6` | Hosted API | High | N/A (closed) | 1M | No | Provider-managed only | Proprietary API terms | Token pricing (`$3` in / `$15` out per MTok, as listed) | N/A | TBD | TBD | TBD | TBD |
| `claude-haiku-4-5` | Hosted API | Medium | N/A (closed) | 200k | No | Provider-managed only | Proprietary API terms | Token pricing (`$1` in / `$5` out per MTok, as listed) | N/A | TBD | TBD | TBD | TBD |
| `Gemini 3.1 Pro` | Hosted API | High | N/A (closed) | (verify latest API docs) | No | Provider-managed only | Proprietary API terms | Token pricing via provider docs | N/A | TBD | TBD | TBD | TBD |
| `Gemini 3 Flash` | Hosted API | High | N/A (closed) | (verify latest API docs) | No | Provider-managed only | Proprietary API terms | Token pricing via provider docs | N/A | TBD | TBD | TBD | TBD |
| `Gemini 3.1 Flash-Lite` | Hosted API | Medium-High | N/A (closed) | (verify latest API docs) | No | Provider-managed only | Proprietary API terms | Token pricing via provider docs | N/A | TBD | TBD | TBD | TBD |

## How to Use With Eval Results

1. Run `llm/evals/run_eval.py` per candidate model.
2. Copy `schema_pass_rate_percent`, `safety_pass_rate_percent`, and `p95_latency_ms` from each `summary.json`.
3. Fill the three `Eval ...` columns above.
4. Choose winner by gate order:
   - License/commercial viability
   - Safety pass rate
   - Schema pass rate
   - Latency/cost

Current Qwen3 values come from `safety_complex_first` prompt profile with safety repair enabled.

For local machine run data, see `local-benchmark-snapshot.md`.

## Source Notes

- Qwen3 model card/docs: verify exact license and context before production sign-off.
- Qwen2.5-7B-Instruct model card: Hugging Face (`License: apache-2.0`, params/context listed).
- Mistral-7B-Instruct-v0.3 model card: Hugging Face (`License: apache-2.0`).
- Llama-3.1-8B-Instruct model card: Hugging Face (`License: llama3.1`, params/context listed).
- OpenAI models page for `gpt-5.4-mini` pricing/context listing.
- Anthropic models overview for Claude 4.6 pricing/context listing.
- Google Gemini pages were reachable, but model-spec fields were not consistently exposed in a stable docs format during fetch; verify from current Gemini API docs at decision time.
