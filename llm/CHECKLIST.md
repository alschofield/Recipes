# LLM Model and Data Sourcing Checklist (Active Backlog)

Only active and partially complete groups are tracked here. Fully complete groups were removed after the latest GitHub-readiness pass.

Completed groups/items are archived in `../docs/archive/checklist-completed-groups.md` and `../docs/archive/checklist-completed-items.md`.

## 1) Model Validation Before Promotion

- [ ] Verify exact model license permits intended commercial usage for `qwen3:8b` and one fallback candidate (ref: `docs/llm/model-stats-matrix.md`).
- [ ] Verify context length and JSON-output reliability under the latest eval suites (ref: `llm/evals/`, `docs/llm/eval-scorecard-template.md`).
- [ ] Verify latency/cost at expected request volume using current traffic assumptions (ref: `docs/llm/local-benchmark-snapshot.md`).
- [ ] Verify safety behavior for cooking-risk prompts and document regressions (ref: `docs/llm/fallback-contract.md`).
- [ ] Decide whether to keep a hosted OpenAI-compatible backup provider for disaster recovery (ref: `docs/llm/serving-options.md`).

## 2) Data Sourcing and Provenance Controls

- [ ] Confirm legal status and attribution obligations for each dataset lane before training use (public-domain, CC, first-party, partner) (ref: `docs/llm/data-provenance-manifest-template.md`).
- [ ] Track source URL/origin, license, and retrieval date per record in provenance manifests (ref: `docs/llm/data-provenance-manifest-template.md`).
- [ ] De-duplicate near-identical recipes before any fine-tune run (ref: `llm/train/qlora/pilot-plan.md`).
- [ ] Maintain denylist for disallowed sources and block unclear-license data from train sets (ref: `docs/llm/data-provenance-manifest-template.md`).
- [ ] Add policy checks for unsafe instructions and allergen-risk phrasing in data prep (ref: `docs/llm/fallback-contract.md`).

## 3) Serving Rollout Controls

- [ ] Add canary rollout toggle before switching default model (ref: `docs/ops/deployment-plan.md`).
- [ ] Add emergency fallback-disable switch and document operator runbook flow (ref: `docs/ops/operations-runbook.md`).
- [ ] Validate rollback path end-to-end (switch model/provider without code changes) in staging (ref: `llm/PRODUCTION-READINESS.md`).

## 4) Judge Model and Metadata Enrichment (V1+)

- [ ] Define lightweight judge model candidate(s) for ingredient metadata inference and quality scoring (ref: `llm/judge/ingredient-metadata-plan.md`, `llm/judge/recipe-quality-plan.md`).
- [ ] Add calibration dataset and acceptance thresholds for judge reliability (ref: `llm/judge/calibration-template.json`).
- [ ] Add calibration tests comparing judge score behavior to deterministic baseline + manual spot checks (ref: `server/CHECKLIST.md`).

## 5) Exit Criteria Before Production Switch

- [ ] Achieve >=95% schema-valid JSON responses on the eval set (ref: `llm/evals/`, `docs/llm/eval-scorecard-template.md`).
- [ ] Pass safety checks for the high-risk prompt suite (ref: `docs/llm/fallback-contract.md`).
- [ ] Meet latency and cost targets against budget ceiling (ref: `docs/llm/local-benchmark-snapshot.md`).
- [ ] Document and approve license/compliance review (ref: `docs/llm/model-stats-matrix.md`, `docs/llm/data-provenance-manifest-template.md`).
- [ ] Sign off rollback path verification in release checklist (ref: `llm/PRODUCTION-READINESS.md`).

## 6) Fresh Workstreams (Next Batch)

- [ ] Complete first QLoRA pilot run record with dataset version, hyperparameters, and eval comparison against `qwen3:8b` base (ref: `llm/train/qlora/pilot-plan.md`).
- [ ] Add automated nightly eval job for recipe/safety/complex suites and publish trend report artifact (ref: `llm/evals/`, `llm/PRODUCTION-READINESS.md`).
- [ ] Add judge-output drift check (confidence and category distribution) and fail alert thresholds when drift exceeds baseline bands (ref: `llm/judge/data-priors.summary.json`).

## Supporting Docs

- `docs/llm/fallback-contract.md`
- `docs/llm/local-docker-setup.md`
- `docs/llm/local-benchmark-snapshot.md`
- `docs/llm/finetune-model-shortlist.md`
- `docs/llm/model-stats-matrix.md`
- `docs/llm/eval-scorecard-template.md`
- `docs/llm/data-provenance-manifest-template.md`
- `docs/llm/serving-options.md`
- `docs/llm/serving-infra-blueprints.md`
- `llm/train/qlora/pilot-plan.md`
- `llm/judge/ingredient-metadata-plan.md`
- `llm/judge/recipe-quality-plan.md`
- `llm/PRODUCTION-READINESS.md`
