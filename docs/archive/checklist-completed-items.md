# Checklist Completed Items Archive

Archived on 2026-03-23.

## Source: `server/CHECKLIST.md`

### Fallback Runtime and Reliability

- Add runtime prompt profile routing (`schema_first` / `safety_complex_first`).
- Add client complex hint support and auto-complex threshold (>=10 ingredients).
- Add one-pass safety repair flow for recoverable schema issues.
- Add runtime `/no_think` control and bounded token output (`LLM_DISABLE_THINKING_TAG`, `LLM_MAX_TOKENS`).

### Observability and Operations

- Add `GET /recipes/health/llm` endpoint exposing runtime config and fallback counters.
- Add machine-friendly LLM metrics endpoint (`GET /recipes/metrics/llm`).

### Ingredient Governance and Metadata

- Resolve or create missing ingredients during LLM recipe persistence.
- Add background judge-triggered enrichment for newly created ingredients (category/coverage/quality/metadata when enabled).

### Judge Model (Non-User-Derived Quality/Metadata)

- Add lightweight judge-model pass for ingredient metadata inference on newly created ingredients.
- Add lightweight judge-model pass for secondary recipe quality score (alongside deterministic score).
- Persist judge outputs with confidence + trace fields and keep deterministic score as fallback.
- Derive initial metadata/quality calibration priors from `datasets/raw/server-lib` and `datasets/derived/server-lib`.

## Source: `web/CHECKLIST.md`

### Search UX and Controls

- Send `complex` hint from recipes search page payload.
- Auto-enable complex mode in UI when 10+ ingredients are entered.

## Source: `llm/CHECKLIST.md`

### Judge Model and Metadata Enrichment (V1+)

- Add judge pass for secondary recipe quality scoring (non-user-derived), persisted alongside deterministic score.
- Keep deterministic non-LLM score path as mandatory fallback when judge model is unavailable.

### Data Sourcing and Provenance Controls

- Track source URL/origin, license, and retrieval date per record in provenance manifests.
- De-duplicate near-identical recipes before fine-tune runs.
- Maintain source denylist and block unclear-license entries from train sets.
- Add policy checks for unsafe instructions and allergen-risk phrasing in data prep.

### Serving Rollout Controls

- Add canary rollout toggle before switching default model.
- Add emergency fallback-disable switch and document operator runbook flow.

### Fresh Workstreams (Next Batch)

- Add automated nightly eval job for recipe/safety/complex suites and publish trend report artifact.
- Add judge-output drift check (confidence and category distribution) with fail thresholds.

## Source: `server/CHECKLIST.md`

### Mobile API Readiness

- Add refresh-token flow for mobile session continuity (`/users/refresh`, `/users/logout`, `/users/logout/session`, `/users/{userid}/sessions`).
- Add idempotency strategy for mobile write retries (favorites first, `Idempotency-Key` middleware for non-idempotent writes).
- Define offline favorites sync conflict and reconciliation contract.
- Add mobile-focused refresh/retry/conflict tests.

## Source: `web/CHECKLIST.md`

### Search UX and Trust

- Add strict/inclusive/complex help text and mixed-source legend.
- Add blend explanation UI and generated-recipe reviewability badge.
- Add strict empty-state guidance and robust retry/error states.

### UI Modernization and Reliability

- Define visual direction board and modernized shell/component language.
- Add motion, density controls, pantry composer ergonomics, and detail print polish.
- Add telemetry and production env wiring verification helper.
- Run accessibility pass for recipes surfaces.

## Source: `mobile/CHECKLIST.md`

### Product/UX/Architecture Definitions

- Define mobile scope, success metrics, offline expectations, and notification strategy.
- Define mobile IA, touch/spacing, and error/loading/empty behavior.
- Define design-token/component parity policy.
- Define API usage, feature flags, and analytics taxonomy.

### Release and Ops Definitions

- Define Android/iOS package IDs, signing, CI lanes, and versioning policies.
- Define QA matrix, crash/perf alert thresholds, and staged rollout cadence.

## Source: `CHECKLIST.md`

### Ongoing Maintenance/Scaffolding

- Add curated changelog-draft script.
- Move server/web/mobile/llm backlogs into their dedicated checklists.
- Add bootstrap guardrails for missing `make`/`task`.

### Monetization Plan

- Define ICP, paid boundaries, pricing, billing, activation funnel, paywall experiments, retention loops, native-ad guardrails, and unit economics dashboard spec.

### Suggested Next Work Order

- Execute server productionization lane as first suggested milestone.

## Source: `mobile/CHECKLIST.md`

### Native Implementation Track

- Implemented auth/session bootstrap flow on both native clients (login, refresh, logout current session, logout all sessions, session listing).
- Hardened favorites offline queue policy on both native clients (dedup/cancel logic, replay order, reconciliation status UX).

### Voice I/O Enhancements (STT + TTS)

- Defined V1 voice UX scope (tap-to-speak search/ingredient entry) and V2 TTS expansion plan.
- Defined V1 voice consent/privacy and data handling policy (no raw audio persistence, explicit permission gating).
- Implemented V1 STT baseline for search ingredient input on Android and iOS with permission gating and fallback messaging.
- Improved V1 STT fallback UX copy/recovery messaging for denied/unavailable/canceled flows on Android and iOS.
- Added lightweight local telemetry counters for V1 STT start/success/failure/permission-denied events.

### V1 Product Quality Gates (Mobile)

- Finalized V1 app naming and in-app/store-facing naming strings across Android and iOS (`Ingrediential`).
- Shipped V1 copy pass for Search/Saved/Profile screens, auth/session statuses, and voice fallback messages.

## Source: `web/CHECKLIST.md`

### V1 Product Quality Gates (Web)

- Finalized V1 app/site naming and core web-facing brand references (`Ingrediential`) across metadata and shell.
- Shipped V1 copy pass for homepage, discover/detail flows, auth/account/favorites, and recovery states.
- Shipped V1 modern UI pass for recipes index/detail (hierarchy, spacing rhythm, typography polish, mobile-first behavior).

### Account/Platform Dependencies

- Configured CI/browser perf audit lane via `.github/workflows/web-perf-audit.yml` and `web/.lighthouserc.json`.
- Finalized production web domain routing (`ingrediential.uk` -> `www.ingrediential.uk`) and validated DNS/TLS baseline for web host.

## Source: `docs/ops/v1-external-inputs-checklist.md`

### Identity and Domains

- Finalized production API domain routing at `api.ingrediential.uk` with Cloudflare Worker gateway upstreams active.

## Source: `CHECKLIST.md`

### Deferred Fixes (Broken but Not Blocking)

- Replaced direct `make`/`task` dependency in release workflow with script-driven automation (`scripts/v1_preflight.py`, `scripts/v1_runtime_smoke.py`, `scripts/v1_open_gates_snapshot.py`), removing shell-command blocker for V1 progression.

### LLM Program Direction

- Removed duplicate root-level judge-model workflow item because the implementation lane was already completed and archived under prior server/LLM completed groups.

### Doc Hygiene

- Moved dated point-in-time docs from active folders into `docs/archive/{mobile,product,llm,deliverables}/` to keep active indexes focused on evergreen guidance.

## Source: `mobile/CHECKLIST.md`

### V1 Product Quality Gates (Mobile)

- Shipped V1 modern native UI pass for Search/Saved/Profile (hierarchy, spacing, typography, interaction clarity) across Android Compose and iOS SwiftUI.
