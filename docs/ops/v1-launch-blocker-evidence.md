# V1 Launch Blocker Evidence Pack

Use this file to capture final evidence for web/mobile/server blocker gates before go/no-go.

Companion tracking files:

- `v1-open-gates-snapshot.md` (generated open-gates state)
- `v1-external-inputs-checklist.md` (non-code required inputs)

## Server Gates

- Staging smoke run timestamp:
- Smoke scope completed (auth, refresh, sessions, favorites, search):
- Production rollback drill timestamp:
- Rollback drill outcome:
- Ops/security approval owner + date: `Alexander Schofield (alex.schofield816@gmail.com)` (date pending final approval)
- Local automated preflight: `go test ./...` in `server/` passed on 2026-03-23.
- Local build preflight: `go build ./...` in `server/` passed on 2026-03-23.
- Consolidated local preflight report: `docs/ops/v1-local-preflight-latest.md` (all local checks passing as of 2026-03-23).
- Consolidated runtime smoke report: `docs/ops/v1-runtime-smoke-latest.md` (web route and Android adb smoke passing as of 2026-03-23).
- API gateway strategy: Cloudflare Worker selected for `api.ingrediential.uk` (cutover checklist: `docs/ops/api-gateway-cutover-ingrediential.md`).
- Gateway upstreams identified:
  - `RECIPES_ORIGIN=https://recipes-production-b30c.up.railway.app`
  - `USERS_ORIGIN=https://users-production-8fab.up.railway.app`
  - `FAVORITES_ORIGIN=https://favorites-production.up.railway.app`

## Web Gates

- Lighthouse/perf run environment:
- Lighthouse key results (recipes index/detail):
- Deployed smoke validation timestamp:
- Domain/TLS/DNS validation completed: `ingrediential.uk` + `www.ingrediential.uk` connected; apex redirects to `https://www.ingrediential.uk`.
- Web owner signoff: `Alexander Schofield (alex.schofield816@gmail.com)` (domain routing confirmed; deployed functional smoke pending full manual sheet).
- Local build evidence: `web` production build successful (`npm run build`, 2026-03-23).
- Local lint/test evidence: `npm run lint` passed and `npm run test` passed (`--passWithNoTests`) on 2026-03-23.
- Local e2e preflight: Playwright smoke suite passed (`web/tests/e2e/smoke.spec.js`, 4 tests) on 2026-03-23.
- Local route smoke preflight: local `next start` smoke returned `HTTP 200` for `/`, `/recipes`, `/login`, `/signup`, `/api/health` on 2026-03-23.
- CI perf lane configured: `.github/workflows/web-perf-audit.yml` with `web/.lighthouserc.json`.
- Local Lighthouse CLI attempt status: blocked by missing local Chrome install (`@lhci/cli` healthcheck) on 2026-03-23; pending GitHub runner execution evidence.

## Mobile Gates

- Android TalkBack manual pass completed:
- iOS VoiceOver manual pass completed:
- Android STT matrix completed:
- iOS STT matrix completed:
- Store metadata/assets complete:
- Play closed test + TestFlight outcomes:
- Signing/CI release artifact verification:
- Mobile owner signoff: `Alexander Schofield (alex.schofield816@gmail.com)` (pending manual/device/store gates)
- Local Android build evidence: `:app:assembleDevDebug` successful (2026-03-23).
- Emulator install/launch evidence: APK installed and launched on AVD `Medium_Phone_API_36.1`.
- Android runtime smoke evidence: `adb install -r` success, launcher monkey invocation success, process PID confirmed (`pidof com.recipes.mobile.dev`) on 2026-03-23.

## Cross-Cutting Product Quality

- V1 naming/copy/UI pass complete for web: yes (Ingrediential rename + copy refresh + discover/detail UX modernization).
- V1 naming/copy/UI pass complete for mobile: yes (Ingrediential rename + Discover/Saved/Profile copy + tab/icon/theme polish).
- Remaining P0/P1 UX defects (if any): pending manual QA verification in real usage sessions.
- Mitigation/waiver notes:

## Final Decision

- Go / No-Go:
- Decision date:
- Approved by: `Alexander Schofield (alex.schofield816@gmail.com)` (single owner mode)
- Notes:

## Snapshot

- Open gates snapshot: `docs/ops/v1-open-gates-snapshot.md`.
- External required inputs: `docs/ops/v1-external-inputs-checklist.md`.
- Quick status dashboard: `docs/ops/v1-gate-dashboard-latest.md`.

## Third-Party Review Inputs

- Security pre-audit: `docs/deliverables/security-pre-audit-report-2026-03-23.md`.
- Legal/compliance pre-audit: `docs/deliverables/legal-compliance-pre-audit-report-2026-03-23.md`.
- Security audit brief: `docs/deliverables/third-party-security-audit-brief-v1.md`.
- Legal review brief: `docs/deliverables/third-party-legal-review-brief-v1.md`.
