# Recipes Docs

This folder has contracts, rollout plans, and LLM ops notes. Keep contracts tight and avoid duplicating rules across multiple files.

## Start Here (Source of Truth)

- `server/search-contract.md` - request/response behavior for recipe search.
- `server/auth-security-baseline.md` - authn/authz and baseline security controls.
- `llm/fallback-contract.md` - DB-first + LLM fallback rules, schema, health/metrics endpoint.
- Folder indexes: `server/README.md`, `ops/README.md`, `product/README.md`, `llm/README.md`, `archive/README.md`.

## LLM Runbook Set

- `llm/local-docker-setup.md` - local Docker/Ollama setup and smoke checks.
- `llm/local-benchmark-snapshot.md` - latest measured local benchmark outcomes.
- `llm/model-stats-matrix.md` - model/license/cost/eval matrix.
- `llm/serving-options.md` - provider/serving architecture options tradeoffs.
- `llm/serving-infra-blueprints.md` - concrete self-hosted serving blueprint.
- `llm/finetune-model-shortlist.md` - QLoRA candidate plan and bakeoff guidance.
- LLM workspace lanes live under `../llm/` (`evals/`, `train/`, `judge/`).
- LLM release gate doc: `../llm/PRODUCTION-READINESS.md`.

## Deployment and Operations

- `ops/deployment-plan.md` - staged rollout, env variables, and release gates.
- `ops/operations-runbook.md` - incident/ops checklists and runtime procedures.
- `ops/provider-setup-template.md` - private env/provider setup template.
- `ops/provider-onboarding-checklist.md` - provider setup validation checklist.
- `ops/hosting-strategy.md` - deployment and host tradeoff notes.
- `ops/session-handoff-2026-03-23.md` - preserved progress snapshot and next actions.

## Product and Domain Reference

- `server/architecture.md` - service map and repository structure.
- `server/ingredient-governance.md` - ingredient normalization/governance process.
- `product/domain-language.md` - shared domain terminology.
- `product/ad-monetization-policy.md` - native ad guardrails and disclosure policy.
- `product/design-modernization-v1.md` - UI modernization direction notes.

## Templates

- `llm/eval-scorecard-template.md`
- `llm/data-provenance-manifest-template.md`

## Doc Hygiene Rules

1. Keep normative behavior in contract docs only.
2. Keep rollout steps in `deployment-plan.md` instead of scattering checklists.
3. Keep historical benchmark details in benchmark docs; link from contracts instead of duplicating values.

## Checklist Index

- Root: `../CHECKLIST.md`
- Server: `../server/CHECKLIST.md`
- Web: `../web/CHECKLIST.md`
- LLM: `../llm/CHECKLIST.md`
- Mobile: `../mobile/CHECKLIST.md`

## Dataset Index

- `../datasets/README.md`
- `../datasets/raw/README.md`
- `../datasets/derived/README.md`
