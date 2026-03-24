# Server Checklist (Deployment + Data Quality)

Goal: keep deployed API reliable while improving seed/training data quality.

Only active/open blockers are listed here. Completed items are archived in `../docs/archive/checklist-completed-items.md`.

## Dataset and Seeding Quality Gates

- [ ] Validate production/staging seed integrity (recipes/ingredients counts, required fields, source distribution, duplicate checks).
  Model: `CODEX_HIGH`
- [x] Add a repeatable remote migration + seed operator flow (with verification queries) for Neon-backed environments.
  Model: `FREE_BALANCED`
- [x] Add API-level data quality smoke checks for `/recipes/catalog` and `/recipes/search` to detect schema/seed regressions early.
  Model: `CODEX_HIGH`

## Environment and Deployment Gates

- [ ] Validate staging deployment smoke with live infra (auth, refresh, sessions, favorites, search) and capture signoff evidence.
  Model: `FREE_BALANCED`
- [ ] Run production deployment readiness/rollback drill in real environment and document results.
  Model: `CODEX_HIGH`

## Ops and Security Approval Gates

- [ ] Confirm production secrets/env wiring and rotation policy in deployment platform (outside repo).
  Model: `FREE_BALANCED`
- [ ] Complete final security/operations approval for launch (session policy, token TTL posture, alerting/monitoring).
  Model: `CODEX_HIGH`
