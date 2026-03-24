# Model Validation Report (2026-03-23)

Scope: close current verification tasks for default serving model + one fallback candidate.

## Verified models

| Role | Model | License (model card) | Context | Commercial-use posture |
|---|---|---|---|---|
| Default | `Qwen/Qwen3-8B` (`qwen3:8b`) | Apache-2.0 | 32,768 native; 131,072 with YaRN | Allowed under Apache-2.0 terms |
| Fallback candidate | `Qwen/Qwen2.5-7B-Instruct` | Apache-2.0 | 131,072 | Allowed under Apache-2.0 terms |

Sources:

- https://huggingface.co/Qwen/Qwen3-8B
- https://huggingface.co/Qwen/Qwen2.5-7B-Instruct

## JSON reliability (latest eval suites in repo)

Reference run: `llm/evals/results/qwen3-8b-profile-compare-repair-20260322/profiles_summary.json`

- `safety_complex_first`: schema pass `91.67%`, safety pass `100.00%`, complex pass `75.00%`, p95 `105376.77ms`
- `schema_first`: schema pass `100.00%`, safety pass `100.00%`, complex pass `0.00%`, p95 `70624.99ms`

Conclusion:

- Reliability is verified and measurable, but not yet fully gate-compliant for production switch.
- Current best operational profile remains `safety_complex_first` for safety+complex quality, with schema/latency still requiring optimization.

## Notes

- This report is verification evidence, not final production sign-off.
- Production sign-off continues to follow `llm/PRODUCTION-READINESS.md` hard gates.
