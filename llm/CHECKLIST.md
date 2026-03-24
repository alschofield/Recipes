# LLM Model and Data Sourcing Checklist (Active Backlog)

Only active and partially complete groups are tracked here. Fully complete groups were removed after the latest GitHub-readiness pass.

Completed groups/items are archived in `../docs/archive/checklist-completed-groups.md` and `../docs/archive/checklist-completed-items.md`.

## 2) Data Sourcing and Provenance Controls

- [x] Define V1 training corpus contract (required schema fields, quality thresholds, source mix targets, dedup policy) and publish alongside provenance manifest.
- [ ] Produce a curated seed/training export from deployed DB data and score for quality/noise before training inclusion.
- [ ] Build a repeatable eval set from real-world query patterns and edge cases observed in production.

- [ ] Confirm legal status and attribution obligations for each dataset lane before training use (public-domain, CC, first-party, partner) (ref: `docs/llm/data-provenance-manifest-template.md`) (blocked: pending legal approval for unresolved third-party lanes).

## 3) Serving Rollout Controls

- [ ] Validate rollback path end-to-end (switch model/provider without code changes) in staging (ref: `llm/PRODUCTION-READINESS.md`) (blocked: requires staging drill execution).

## 5) Exit Criteria Before Production Switch

- [ ] Achieve >=95% schema-valid JSON responses on the eval set (ref: `llm/evals/`, `docs/llm/eval-scorecard-template.md`) (blocked: current baseline schema gate failing).
- [ ] Pass safety checks for the high-risk prompt suite (ref: `docs/llm/fallback-contract.md`) (pending: keep nightly evidence current).
- [ ] Meet latency and cost targets against budget ceiling (ref: `docs/llm/local-benchmark-snapshot.md`) (blocked: current p95/capacity posture above target assumptions).
- [ ] Document and approve license/compliance review (ref: `docs/llm/model-stats-matrix.md`, `docs/llm/data-provenance-manifest-template.md`) (blocked: legal approval pending).
- [ ] Sign off rollback path verification in release checklist (ref: `llm/PRODUCTION-READINESS.md`) (blocked: staging rollback drill pending).

## 6) Fresh Workstreams (Next Batch)

- [ ] Draft schema V2 proposal for recipe ingredient measurements (keep `amount`; add optional `quantity`, `unit`, `prep`) and publish migration notes.
- [ ] Update `docs/llm/fallback-contract.md`, `llm/evals/run_eval.py`, and SFT dataset generators for schema V2 with backward-compatible validation.
- [ ] Complete first QLoRA pilot run record with dataset version, hyperparameters, and eval comparison against `qwen3:8b` base (ref: `llm/train/qlora/pilot-plan.md`) (blocked: first pilot training/eval run not yet executed).

## Supporting Docs

- `docs/llm/fallback-contract.md`
- `docs/llm/local-docker-setup.md`
- `docs/llm/local-benchmark-snapshot.md`
- `docs/archive/llm/latency-cost-validation-2026-03-23.md`
- `docs/llm/finetune-model-shortlist.md`
- `docs/llm/model-stats-matrix.md`
- `docs/archive/llm/model-validation-report-2026-03-23.md`
- `docs/archive/llm/safety-regression-notes-2026-03-23.md`
- `docs/archive/llm/provider-dr-decision-2026-03-23.md`
- `docs/archive/llm/open-items-execution-plan-2026-03-23.md`
- `docs/llm/eval-scorecard-template.md`
- `docs/llm/data-provenance-manifest-template.md`
- `docs/llm/serving-options.md`
- `docs/llm/serving-infra-blueprints.md`
- `llm/train/qlora/pilot-plan.md`
- `llm/judge/ingredient-metadata-plan.md`
- `llm/judge/recipe-quality-plan.md`
- `llm/PRODUCTION-READINESS.md`
