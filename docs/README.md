# Recipes Docs

This folder has contracts, rollout plans, and LLM ops notes. Keep contracts tight and avoid duplicating rules across multiple files.

## Current Priority Themes

- Dataset curation quality for DB seeding and LLM training/eval.
- UI/UX quality uplift across web and native mobile.
- External security/legal/design review readiness packs.

## Start Here (Source of Truth)

- `server/search-contract.md` - request/response behavior for recipe search.
- `server/auth-security-baseline.md` - authn/authz and baseline security controls.
- `llm/fallback-contract.md` - DB-first + LLM fallback rules, schema, health/metrics endpoint.
- Folder indexes: `server/README.md`, `ops/README.md`, `product/README.md`, `llm/README.md`, `mobile/README.md`, `deliverables/README.md`, `archive/README.md`.

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
- `ops/v1-launch-blocker-evidence.md` - final evidence capture for launch blocker gates.
- `ops/v1-manual-gate-runbook.md` - manual execution guide for remaining launch gates.
- `ops/v1-server-web-smoke-command-pack.md` - deployed smoke command templates for server/web gates.
- `ops/v1-local-preflight-latest.md` - latest generated local preflight report.
- `ops/v1-open-gates-snapshot.md` - latest generated consolidated open-gates snapshot.
- `ops/v1-external-inputs-checklist.md` - required external inputs and approvals.
- `ops/v1-runtime-smoke-latest.md` - latest generated local runtime smoke report.
- `ops/v1-gate-dashboard-latest.md` - compact generated gate dashboard for quick visibility.
- `ops/v1-human-execution-session.md` - live operator session flow for manual/external gates.
- `ops/api-gateway-cutover-ingrediential.md` - Cloudflare Worker gateway cutover checklist for `api.ingrediential.uk`.

## Product and Domain Reference

- `server/architecture.md` - service map and repository structure.
- `server/ingredient-governance.md` - ingredient normalization/governance process.
- `product/domain-language.md` - shared domain terminology.
- `product/v1-brand-copy-system.md` - Ingrediential V1 naming/tone/copy baseline.
- `product/v1-web-usability-sanity-sheet.md` - web UX sanity checklist and defect log template.
- `product/ad-monetization-policy.md` - native ad guardrails and disclosure policy.
- `product/design-modernization-v1.md` - UI modernization direction notes.
- `mobile/api-usage-policy-v1.md` - native API usage, retry/cache/offline/session policy.

## Shareable Deliverables

- `archive/deliverables/security-pre-audit-report-2026-03-23.md`
- `archive/deliverables/legal-compliance-pre-audit-report-2026-03-23.md`
- `deliverables/third-party-security-audit-brief-v1.md`
- `deliverables/third-party-legal-review-brief-v1.md`
- `deliverables/third-party-audit-options-v1.md`
- `deliverables/third-party-design-ux-brief-v1.md`
- `deliverables/third-party-data-curation-brief-v1.md`

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
