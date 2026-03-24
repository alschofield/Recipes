# Recipes Project Checklist (Current Priorities)

This checklist now tracks the active next-phase priorities after first public deployment.
Archive location for completed groups/items: `docs/archive/`.

## Model Selection Policy (Cost-First)

Use the cheapest model that can reliably complete the task.

- `FREE_FAST` - OpenCode Zen (primary), Big Pickle (fallback)
- `FREE_BALANCED` - MiMo V2 Pro Free (primary), MiniMax M2.5 Free (fallback)
- `CODEX_HIGH` - GPT-5.3 Codex Spark (primary), GPT-5.2 Codex (fallback)
- `ANTHROPIC_STRONG` - Claude Sonnet 4.5 (optional, higher-cost deep reasoning). Caveat: Anthropic API is currently not working in the maintainer's setup, so this tier may be unavailable depending on who is running the checklist.

Escalation rule: if a task fails twice on current tier, move up one tier.

---

## Priority A: Dataset Curation (DB Seeding + LLM Training)

- [ ] Define V1 seed dataset acceptance criteria (minimum record counts, required fields, quality thresholds, duplicate policy) and publish in docs.
  Model: `FREE_BALANCED`
- [ ] Run staging/prod data profiling and fix schema/data issues affecting `/recipes/catalog` and search quality.
  Model: `CODEX_HIGH`
- [ ] Build repeatable seed refresh workflow with verification report (counts, null/duplicate checks, source distribution) and operator runbook.
  Model: `CODEX_HIGH`
- [ ] Curate LLM training/eval datasets from approved lanes with provenance and dedup checks.
  Model: `CODEX_HIGH`
- [ ] Define and implement schema V2 for ingredient measurements (`amount` compatibility + structured `quantity`/`unit`/`prep`) across contract, eval harness, and dataset generators.
  Model: `CODEX_HIGH`
- [ ] Execute first QLoRA pilot cycle in provisioned training environment and document promote/iterate/rollback decision.
  Model: `CODEX_HIGH`

---

## Priority B: Product UI/UX Quality (Web + Mobile)

- [ ] Close all known "weird" UX issues via structured triage log and P0/P1 fix pass on web and mobile.
  Model: `CODEX_HIGH`
- [ ] Ship V1.1 visual polish pass for web discover/detail/auth flows (hierarchy, motion restraint, readability, empty/error clarity).
  Model: `CODEX_HIGH`
- [ ] Ship V1.1 visual polish pass for mobile Discover/Saved/Profile flows (spacing, typography, status clarity, action ergonomics).
  Model: `CODEX_HIGH`
- [ ] Prepare third-party design/UX audit packet and decide whether to contract specialist support before broader launch.
  Model: `FREE_BALANCED`

---

## Priority C: Release Readiness for Option B (Split Providers)

- [ ] Fill `docs/ops/v1-launch-blocker-evidence.md` with latest server/web/mobile gate evidence before final go/no-go.
  Model: `FREE_FAST`
- [ ] Create a filled private copy of `docs/ops/provider-setup-template.md` (do not commit) and verify against `docs/ops/provider-onboarding-checklist.md`.
  Model: `FREE_FAST`
- [ ] Provision staging infrastructure and deploy staging web + API gateway + services.
  Model: `FREE_BALANCED`
- [ ] Wire GitHub deploy hook secrets for staging and validate `.github/workflows/deploy-staging.yml`.
  Model: `FREE_BALANCED`
- [ ] Run staging smoke suite (auth, search, blend quality, favorites, detail).
  Model: `CODEX_HIGH`
- [ ] Provision production infra and custom domains (`www`, `api`), TLS, DNS, and CORS lockdown.
  Model: `FREE_BALANCED`
- [ ] Enable and validate `.github/workflows/deploy-prod.yml` with rollback drill.
  Model: `CODEX_HIGH`

---

## Suggested Next Work Order

- 1) Stabilize deployed data layer (migrations + seed quality + catalog/search integrity).
- 2) Run structured web/mobile UX triage and close P0/P1 polish issues.
- 3) Continue LLM data curation + first QLoRA pilot evidence.
- 4) Clear remaining external launch gates in server/web/mobile/llm checklists.
