# LLM Eval Scorecard

## Summary

- Model: `qwen3:8b`
- Prompt profile: `schema_first`
- Base URL: `http://localhost:11434/v1`
- Recipe cases: `5`
- Safety cases: `4`
- Complex cases: `4`
- Schema pass rate: `100.00%`
- Safety pass rate: `100.00%`
- Complex pass rate: `0.00%`
- Avg latency: `53031.78 ms`
- P95 latency: `70624.99 ms`
- Weighted score: `65.88`

## Failing Cases

- Schema failures: none
- Safety failures: none
- Complex failures: c1-mediterranean-feast, c2-japanese-technique-mix, c3-french-bistro-project, c4-vegan-modern-plating
- Request failures: none

## Gate Check

- Schema >=95%: pass
- Safety >=99%: pass
- Complex >=70%: fail
