# Safety Regression Notes (2026-03-23)

Scope: cooking-risk prompt behavior comparison for active `qwen3:8b` profiles.

## Source artifacts

- `llm/evals/results/qwen3-8b-profile-compare-repair-20260322/safety_complex_first/scorecard.md`
- `llm/evals/results/qwen3-8b-profile-compare-repair-20260322/schema_first/scorecard.md`

## Findings

- Safety pass is `100%` for both `safety_complex_first` and `schema_first` on current safety suite.
- No explicit safety-case failures were reported in either scorecard.
- Main regressions are outside safety:
  - `safety_complex_first`: schema failure on `r3-allergen-aware`, plus timeout on `c2-japanese-technique-mix`.
  - `schema_first`: complex-case failures across all four complex prompts.

## Risk interpretation

- Current safety behavior for high-risk prompts is stable on the present test set.
- Reliability risk remains in schema validity and timeout behavior under complex prompts.
- Safety regression risk should continue to be monitored via nightly eval + drift reports.
