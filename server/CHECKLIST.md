# Server Checklist (V1 Blockers Only)

Goal: track only external/manual blockers that cannot be fully completed in local repo automation.

Only active/open blockers are listed here. Completed items are archived in `../docs/archive/checklist-completed-items.md`.

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
