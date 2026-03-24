# V1 Manual Gate Runbook

Use this runbook to clear remaining external/manual launch blockers and keep evidence consistent.

For a single-session operator flow, see `v1-human-execution-session.md`.

## 0) Local Preflight (before deployed/manual gates)

- Run `python scripts/v1_preflight.py` from repo root.
- Confirm `docs/ops/v1-local-preflight-latest.md` reports all local checks as `PASS`.
- If any check fails, fix locally before consuming manual QA or deployment time.
- Run `python scripts/v1_runtime_smoke.py --with-android` and attach `docs/ops/v1-runtime-smoke-latest.md`.

Recommended execution order for one release session:

1. Local preflight (`scripts/v1_preflight.py`)
2. Server staging smoke
3. Web usability/accessibility sanity + perf/domain gates
4. Mobile accessibility/voice/UX sanity
5. Store/compliance gates
6. Final go/no-go decision capture

## 1) Server Staging Smoke

Scope:

- auth login/refresh/logout/session-revoke/session-list
- recipes search/detail
- favorites add/remove/list

Record in `v1-launch-blocker-evidence.md`:

- timestamp
- scope pass/fail
- blocker defects with owners

Execution aid:

- use `v1-server-web-smoke-command-pack.md` command templates

## 2) Web Usability + Accessibility Sanity

Validate on deployed environment:

- home/discover/detail/favorites/account flow clarity
- keyboard focus order and visible focus ring
- error/retry paths and empty states
- responsive behavior on mobile and desktop widths

Execution aid:

- use `../product/v1-web-usability-sanity-sheet.md`

Record:

- pass/fail
- P0/P1 defect list and fix status

## 3) Web Perf and Domain Gates

- run Lighthouse in compatible runtime on discover/detail
- confirm production DNS/TLS/domain wiring
- capture owner signoff

Record key metrics and signoff in evidence pack.

## 4) Mobile Accessibility and Voice Gates

- run Android TalkBack pass (`docs/mobile/native-accessibility-manual-test-sheet-v1.md`)
- run iOS VoiceOver pass (`docs/mobile/native-accessibility-manual-test-sheet-v1.md`)
- run Android/iOS STT matrix (`docs/mobile/voice-stt-validation-matrix-v1.md`)
- run native UX sanity pass (`docs/mobile/v1-mobile-ux-sanity-sheet.md`)

Record:

- completion status per platform
- blocker defects and triage outcome

## 5) Store and Compliance Gates

- fill and finalize release copy from `docs/mobile/store-submission-pack-template-v1.md` (keep dated working copy in archive if needed)
- complete Play/App Store privacy declarations
- run Play closed track + TestFlight cycles

Record outcomes in evidence pack.

## 6) Final Go/No-Go

Before decision:

- confirm no unresolved P0 blockers
- confirm waiver notes for any P1 exceptions
- capture final approver and decision timestamp
