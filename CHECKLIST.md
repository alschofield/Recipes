# Recipes Project Checklist (V1 Blockers Only)

Only external/manual blockers are listed here.
Archive location for completed groups/items: `docs/archive/`.

## Model Selection Policy (Cost-First)

Use the cheapest model that can reliably complete the task.

- `FREE_FAST` - OpenCode Zen (primary), Big Pickle (fallback)
- `FREE_BALANCED` - MiMo V2 Pro Free (primary), MiniMax M2.5 Free (fallback)
- `CODEX_HIGH` - GPT-5.3 Codex Spark (primary), GPT-5.2 Codex (fallback)
- `ANTHROPIC_STRONG` - Claude Sonnet 4.5 (optional, higher-cost deep reasoning). Caveat: Anthropic API is currently not working in the maintainer's setup, so this tier may be unavailable depending on who is running the checklist.

Escalation rule: if a task fails twice on current tier, move up one tier.

---

## Release Readiness for Option B (Split Providers)

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

## LLM Program Direction

- [ ] Execute first QLoRA pilot cycle in provisioned training environment and document promote/iterate/rollback decision.
  Model: `CODEX_HIGH`

---

## Suggested Next Work Order

- 1) Run `python scripts/v1_preflight.py` and attach `docs/ops/v1-local-preflight-latest.md` to release evidence.
- 2) Clear `server/CHECKLIST.md` blocker gates (staging/prod drill + ops approval).
- 3) Clear `web/CHECKLIST.md` and `mobile/CHECKLIST.md` manual/deployed gates using `docs/ops/v1-manual-gate-runbook.md`.
- 4) Clear remaining `llm/CHECKLIST.md` external blocker gates.
