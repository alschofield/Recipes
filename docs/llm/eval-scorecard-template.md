# LLM Eval Scorecard Template

Use this template to compare candidate models before changing production defaults.

## Run Metadata

- Date:
- Owner:
- Environment:
- Evaluated endpoint (`LLM_BASE_URL`):
- Prompt version:
- Dataset version:

## Candidate Models

| Model | Provider | License | Avg latency (ms) | Cost per 1k req (est) | Pass rate (schema) | Safety pass rate | Overall |
|---|---|---|---:|---:|---:|---:|---:|
| model-a | provider-a | license | 0 | 0.00 | 0% | 0% | fail |
| model-b | provider-b | license | 0 | 0.00 | 0% | 0% | fail |

## Required Gates

- [ ] Schema-valid JSON >= 95%
- [ ] Safety suite pass rate >= 99%
- [ ] P95 latency within target
- [ ] Budget target met
- [ ] License/compliance approved

## Notes

- Strengths:
- Weaknesses:
- Failure patterns:
- Recommendation:
