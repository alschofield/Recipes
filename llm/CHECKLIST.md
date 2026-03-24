# LLM Model and Data Sourcing Checklist (Active Backlog)

Only active and partially complete groups are tracked here. Fully complete groups were removed after the latest GitHub-readiness pass.

Completed groups/items are archived in `../docs/archive/checklist-completed-groups.md` and `../docs/archive/checklist-completed-items.md`.

## 2) Data Sourcing and Provenance Controls

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

- [ ] Complete first QLoRA pilot run record with dataset version, hyperparameters, and eval comparison against `qwen3:8b` base (ref: `llm/train/qlora/pilot-plan.md`) (blocked: first pilot training/eval run not yet executed).

## Supporting Docs

- `docs/llm/fallback-contract.md`
- `docs/llm/local-docker-setup.md`
- `docs/llm/local-benchmark-snapshot.md`
- `docs/llm/latency-cost-validation-2026-03-23.md`
- `docs/llm/finetune-model-shortlist.md`
- `docs/llm/model-stats-matrix.md`
- `docs/llm/model-validation-report-2026-03-23.md`
- `docs/llm/safety-regression-notes-2026-03-23.md`
- `docs/llm/provider-dr-decision-2026-03-23.md`
- `docs/llm/open-items-execution-plan-2026-03-23.md`
- `docs/llm/eval-scorecard-template.md`
- `docs/llm/data-provenance-manifest-template.md`
- `docs/llm/serving-options.md`
- `docs/llm/serving-infra-blueprints.md`
- `llm/train/qlora/pilot-plan.md`
- `llm/judge/ingredient-metadata-plan.md`
- `llm/judge/recipe-quality-plan.md`
- `llm/PRODUCTION-READINESS.md`
