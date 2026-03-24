# Third-Party Security Audit Brief (V1)

Use this brief when engaging an external security reviewer.

## Objective

Perform an independent security assessment before V1 launch for server/web/mobile surfaces.

## Project Snapshot

- Product: Ingrediential
- Stack:
  - Server: Go, PostgreSQL, Redis
  - Web: Next.js
  - Mobile: Kotlin/Compose + SwiftUI
- Auth/session capabilities include refresh rotation and session revoke/list APIs.

## Required Scope

1. API security assessment
   - auth/session endpoints
   - authorization boundaries
   - input validation and error handling
2. Web app assessment
   - auth/session handling
   - client/server boundary and route protections
   - security header/cookie posture review
3. Mobile app assessment
   - secure token/session storage
   - voice permission and fallback behavior
4. Deployment/ops posture review
   - secret management model
   - CI/workflow controls
   - rollback and incident controls

## Desired Deliverables

- Executive summary with risk rating.
- Technical findings with severity (P0/P1/P2), exploitation notes, and remediation steps.
- Re-test confirmation notes after fixes.
- Letter of assessment or equivalent completion statement.

## Artifacts to Share

- `docs/deliverables/security-pre-audit-report-2026-03-23.md`
- `docs/ops/v1-launch-blocker-evidence.md`
- `docs/ops/v1-local-preflight-latest.md`
- `docs/ops/v1-runtime-smoke-latest.md`
- `docs/ops/v1-open-gates-snapshot.md`

## Engagement Model

- Recommended: fixed-scope assessment + one verification pass.
- Timebox: 1-2 weeks initial review, 2-4 days re-test window.
