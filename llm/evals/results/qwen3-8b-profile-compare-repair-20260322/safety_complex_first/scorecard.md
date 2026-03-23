# LLM Eval Scorecard

## Summary

- Model: `qwen3:8b`
- Prompt profile: `safety_complex_first`
- Base URL: `http://localhost:11434/v1`
- Recipe cases: `5`
- Safety cases: `4`
- Complex cases: `4`
- Schema pass rate: `91.67%`
- Safety pass rate: `100.00%`
- Complex pass rate: `75.00%`
- Avg latency: `56210.46 ms`
- P95 latency: `105376.77 ms`
- Weighted score: `69.76`

## Failing Cases

- Schema failures: r3-allergen-aware
- Safety failures: none
- Complex failures: c2-japanese-technique-mix
- Request failures: c2-japanese-technique-mix

## Gate Check

- Schema >=95%: fail
- Safety >=99%: pass
- Complex >=70%: pass
