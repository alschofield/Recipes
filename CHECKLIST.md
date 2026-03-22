# Recipes Project Checklist (Fresh Backlog)

Only open work is listed here. Completed items were intentionally removed.

## Model Selection Policy (Cost-First)

Use the cheapest model that can reliably complete the task.

- `FREE_FAST` - OpenCode Zen (primary), Big Pickle (fallback)
- `FREE_BALANCED` - MiMo V2 Pro Free (primary), MiniMax M2.5 Free (fallback)
- `CODEX_HIGH` - GPT-5.3 Codex Spark (primary), GPT-5.2 Codex (fallback)
- `ANTHROPIC_STRONG` - Claude Sonnet 4.5 (optional, higher-cost deep reasoning). Caveat: Anthropic API is currently not working in the maintainer's setup, so this tier may be unavailable depending on who is running the checklist.

Escalation rule: if a task fails twice on current tier, move up one tier.

---

## Ongoing Source-of-Truth Maintenance

- [ ] Keep contract docs in sync with behavior changes (`docs/search-contract.md`, `docs/auth-security-baseline.md`, `docs/llm-fallback-contract.md`).
  Model: `FREE_FAST`

---

## Product Direction: Always-Fresh Recipe Mix

Goal: include at least one LLM-generated recipe in normal search responses (when generation succeeds), while preserving relevance and trust.

- [ ] Define policy: minimum generated recipe ratio per page (for example `>= 1`, cap `<= 30%`) with fallback rules when generation is unavailable.
  Model: `CODEX_HIGH`
- [ ] Update search pipeline to fetch DB candidates and generated candidates in the same request flow.
  Model: `CODEX_HIGH`
- [ ] Add deterministic interleaving strategy (for example weighted blend by score + source diversity) so results are not grouped by source.
  Model: `CODEX_HIGH`
- [ ] Add response metadata fields to indicate source and blend rationale (`source`, `blendSlot`, `rankingReason`).
  Model: `FREE_BALANCED`
- [ ] Implement server-side shuffle strategy with stable seed support (for pagination consistency and test reproducibility).
  Model: `CODEX_HIGH`
- [ ] Ensure shuffle/interleave happens before frontend response payload is returned.
  Model: `FREE_BALANCED`
- [ ] Add guardrails so low-quality generated recipes do not displace clearly better DB recipes.
  Model: `CODEX_HIGH`
- [ ] Add tests for blend/shuffle determinism, pagination stability, and source-mix guarantees.
  Model: `CODEX_HIGH`

---

## UI Modernization (State-of-the-Art Pass)

Goal: upgrade visual quality from "functional" to "premium modern product".

- [ ] Define a sharper visual direction board (type scale, elevation, spacing rhythm, surface system, color intent).
  Model: `FREE_BALANCED`
- [ ] Redesign top-level layout shell (header/nav/content framing) with stronger hierarchy and modern spacing.
  Model: `FREE_BALANCED`
- [ ] Replace basic cards/buttons with a cohesive component language (interactive states, subtle depth, refined radii).
  Model: `FREE_BALANCED`
- [ ] Add motion system for page transitions, list reveals, and state changes (tasteful, minimal, performant).
  Model: `FREE_BALANCED`
- [ ] Upgrade recipe results presentation (visual density controls, richer metadata chips, cleaner typography).
  Model: `FREE_BALANCED`
- [ ] Improve pantry input experience (faster add/remove, keyboard-first flow, better suggestion affordances).
  Model: `FREE_BALANCED`
- [ ] Improve recipe detail page design (readability-first layout, print mode polish, better step/ingredient structure).
  Model: `FREE_BALANCED`
- [ ] Add empty/loading/error states with branded visuals and clearer recovery actions.
  Model: `FREE_FAST`
- [ ] Run a full accessibility + contrast pass after redesign and fix any regressions.
  Model: `CODEX_HIGH`

---

## Release Readiness for Option B (Split Providers)

- [ ] Create a filled private copy of `docs/provider-setup-template.md` (do not commit).
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

Goal: define and ship production-ready Android and iOS experiences with reliable UX, release flow, and compliance.

### Product Scope

- [ ] Define core mobile jobs-to-be-done and success metrics (activation, 7-day retention, recipe saves, repeat sessions).
  Model: `FREE_BALANCED`
- [ ] Define offline expectations (read-only favorites, cached recipe detail, queued actions).
  Model: `FREE_BALANCED`
- [ ] Define notification strategy (reminders, re-engagement, moderation/admin alerts if needed).
  Model: `FREE_FAST`

### Mobile UX and Navigation

- [ ] Define mobile-first IA (top app bar + bottom nav + modal filter sheets).
  Model: `FREE_BALANCED`
- [ ] Define touch target, spacing, and gesture rules for all high-frequency actions.
  Model: `FREE_FAST`
- [ ] Define empty/loading/error states for mobile screens with recovery actions.
  Model: `FREE_FAST`

### Design System Parity

- [ ] Map web tokens to mobile token set (color, typography, spacing, radius, elevation).
  Model: `FREE_BALANCED`
- [ ] Create shared component spec for buttons, chips, cards, inputs, list rows, and status badges.
  Model: `FREE_BALANCED`
- [ ] Run accessibility pass for contrast, dynamic text, and screen reader labels on mobile.
  Model: `CODEX_HIGH`

### Mobile Architecture and Data

- [ ] Define API usage policy for mobile (cache-first reads, retry/backoff, auth refresh behavior).
  Model: `CODEX_HIGH`
- [ ] Define feature-flag strategy for staged rollout and kill switches.
  Model: `FREE_BALANCED`
- [ ] Define analytics event taxonomy and funnel instrumentation for mobile journeys.
  Model: `FREE_BALANCED`

### Android Build and Release Engineering

- [ ] Finalize package ID, signing key strategy, and secure key storage policy.
  Model: `CODEX_HIGH`
- [ ] Define CI build lane for internal, beta, and production tracks.
  Model: `FREE_BALANCED`
- [ ] Define versioning policy (`versionCode`/`versionName`) and release notes workflow.
  Model: `FREE_FAST`

### iOS Build and Release Engineering

- [ ] Finalize iOS bundle ID, signing certificates/profiles, and secure keychain handling.
  Model: `CODEX_HIGH`
- [ ] Define CI build lane for TestFlight internal/external testing and App Store production release.
  Model: `FREE_BALANCED`
- [ ] Define iOS versioning policy (`CFBundleShortVersionString`/`CFBundleVersion`) and release notes workflow.
  Model: `FREE_FAST`

### Play Store Readiness

- [ ] Prepare store assets (icon, feature graphic, screenshots, optional promo video).
  Model: `FREE_FAST`
- [ ] Prepare privacy policy/data safety disclosures and complete content rating questionnaire.
  Model: `CODEX_HIGH`
- [ ] Run closed testing track and collect pre-launch report before production submission.
  Model: `FREE_BALANCED`

### Apple App Store Readiness

- [ ] Prepare App Store Connect assets (app icon, screenshots by device class, preview videos if used).
  Model: `FREE_FAST`
- [ ] Prepare App Privacy details, data collection declarations, and required policy URLs.
  Model: `CODEX_HIGH`
- [ ] Run TestFlight cycle, finalize metadata, and submit with release checklist for review.
  Model: `FREE_BALANCED`

### Quality, Reliability, and Operations

- [ ] Define device/version QA matrix and regression checklist.
  Model: `FREE_BALANCED`
- [ ] Define crash/performance monitoring and alert thresholds.
  Model: `FREE_BALANCED`
- [ ] Define staged rollout plan, rollback criteria, and post-launch triage cadence.
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

## Suggested Next Work Order

- [ ] 1) Finalize LLM+DB blend policy and API response contract.
- [ ] 2) Implement backend interleave + shuffle with deterministic tests.
- [ ] 3) Execute UI modernization pass on results, shell, and detail pages.
- [ ] 4) Validate staging deployment with Option B single-domain API gateway.
- [ ] 5) Launch production with custom domains and monitoring.
