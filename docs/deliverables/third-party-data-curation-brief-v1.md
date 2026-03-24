# Third-Party Data Curation Brief (V1)

Use this brief when engaging external support for dataset curation and quality controls.

## Objective

Improve recipe/ingredient dataset quality for both production DB seeding and LLM training/evaluation pipelines.

## Scope

1. Seed dataset quality review
   - schema completeness
   - duplicate detection and merge policy
   - source quality classification
2. LLM dataset curation review
   - provenance checks
   - dedup/noise filtering
   - evaluation set representativeness
3. Operator workflow recommendations
   - repeatable refresh cadence
   - quality report format and thresholds

## Required Deliverables

- Data quality report with blocking/non-blocking issues.
- Curated dataset acceptance criteria proposal.
- Recommended curation pipeline and runbook updates.
- Verification checklist for each refresh cycle.

## Artifacts to Share

- `llm/CHECKLIST.md`
- `docs/llm/data-provenance-manifest-template.md`
- `docs/server/search-load-baseline.md`
- `docs/server/mobile-api-baseline.md`
- `docs/ops/v1-launch-blocker-evidence.md`

## Engagement Model

- Recommended: fixed-scope review sprint with one remediation follow-up pass.
