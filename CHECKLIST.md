# Recipes Project Checklist (Fresh Backlog)

Only open work is listed here. Completed items were intentionally removed.
Archive location for completed groups/items: `docs/archive/`.

## Model Selection Policy (Cost-First)

Use the cheapest model that can reliably complete the task.

- `FREE_FAST` - OpenCode Zen (primary), Big Pickle (fallback)
- `FREE_BALANCED` - MiMo V2 Pro Free (primary), MiniMax M2.5 Free (fallback)
- `CODEX_HIGH` - GPT-5.3 Codex Spark (primary), GPT-5.2 Codex (fallback)
- `ANTHROPIC_STRONG` - Claude Sonnet 4.5 (optional, higher-cost deep reasoning). Caveat: Anthropic API is currently not working in the maintainer's setup, so this tier may be unavailable depending on who is running the checklist.

Escalation rule: if a task fails twice on current tier, move up one tier.

---

## Ongoing Source-of-Truth Maintenance

- [ ] Keep contract docs in sync with behavior changes (`docs/server/search-contract.md`, `docs/server/auth-security-baseline.md`, `docs/llm/fallback-contract.md`).
  Model: `FREE_FAST`
- [ ] Keep project checklists aligned across root + `server/CHECKLIST.md` + `web/CHECKLIST.md` + `llm/CHECKLIST.md` + `mobile/CHECKLIST.md`.
  Model: `FREE_FAST`
- [ ] Add a lightweight script to generate a curated changelog draft from recent non-merge commits (`CHANGELOG.md`).
  Model: `FREE_BALANCED`

---

## Server and Web Workstreams

- [ ] Server-specific backlog moved to `server/CHECKLIST.md` (search blend, fallback reliability, judge model, ingredient metadata/governance).
  Model: `FREE_FAST`
- [ ] Web-specific backlog moved to `web/CHECKLIST.md` (recipes UX, complex-mode controls, modernization, accessibility).
  Model: `FREE_FAST`

---

## Release Readiness for Option B (Split Providers)

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

## Mobile App Roadmap (Google Play + Apple App Store)

- [ ] Mobile roadmap moved to `mobile/CHECKLIST.md`.
  Model: `FREE_FAST`

---

## LLM Program Direction

- [ ] Keep LLM sourcing/training/serving backlog in `llm/CHECKLIST.md` and treat it as source of truth for model decisions.
  Model: `FREE_FAST`
- [ ] Finish first QLoRA pilot cycle and document promote/iterate/rollback decision.
  Model: `CODEX_HIGH`
- [ ] Add judge-model workflow for ingredient metadata enrichment + secondary quality scoring (non-user-derived).
  Model: `CODEX_HIGH`

---

## Deferred Fixes (Broken but Not Blocking)

Use this section to log known issues that can wait. Keep each item scoped and actionable.

- [ ] `make` command unavailable in current maintainer environment (`/usr/bin/bash: make: command not found`) when trying `make migrate-up`.
  Model: `FREE_FAST`
- [ ] `task` command unavailable in current maintainer environment (`/usr/bin/bash: task: command not found`) when trying `task migrate-up`.
  Model: `FREE_FAST`
- [ ] Add installation/bootstrap guardrails to detect missing `make`/`task` and print fallback commands automatically.
  Model: `FREE_BALANCED`

---

## Monetization Plan (Debt-Reduction Priority)

Goal: ship a pricing and conversion strategy that can fund operations and produce meaningful personal income.

- [ ] Define ICP and willingness-to-pay segments (busy professionals, families, fitness-focused, dietary-restricted users).
  Model: `FREE_BALANCED`
- [ ] Define paid feature boundaries (free vs pro) for recipe generation limits, saved plans, premium filters, and export options.
  Model: `CODEX_HIGH`
- [ ] Create pricing hypothesis (monthly and annual), with a target gross margin after LLM + infra costs.
  Model: `CODEX_HIGH`
- [ ] Add billing infrastructure decision and implementation plan (Stripe subscriptions, trials, failed-payment recovery).
  Model: `FREE_BALANCED`
- [ ] Define activation funnel metrics (visit -> signup -> first recipe -> saved favorite -> paid conversion).
  Model: `FREE_BALANCED`
- [ ] Add paywall experiment plan (timing, messaging, limit thresholds, A/B variants).
  Model: `CODEX_HIGH`
- [ ] Investigate native ad monetization (sponsored ingredients/recipes/placements) with UX guardrails, disclosure rules, and impact on conversion/retention.
  Model: `FREE_BALANCED`
- [ ] Add retention loop plan (weekly meal plan reminders, shopping list utility, streaks, and reactivation prompts).
  Model: `FREE_BALANCED`
- [ ] Build unit economics dashboard (CAC proxy, conversion rate, churn, ARPU, LTV, inference cost per active user).
  Model: `CODEX_HIGH`

---

## Suggested Next Work Order

- [ ] 1) Execute `server/CHECKLIST.md` productionization lane (judge model + ingredient metadata + fallback metrics in `/recipes/metrics`).
- [ ] 2) Execute `web/CHECKLIST.md` UX lane (complex controls + modernization + accessibility).
- [ ] 3) Finish first QLoRA pilot and compare against baseline eval artifacts.
- [ ] 4) Validate staging deployment with Option B single-domain API gateway.
- [ ] 5) Launch production with custom domains and monitoring.
