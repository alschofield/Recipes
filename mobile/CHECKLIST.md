# Mobile App Roadmap (Google Play + Apple App Store)

Goal: define and ship production-ready Android and iOS experiences with reliable UX, release flow, and compliance.

Only active/open work is listed here. Completed items should be moved to `../docs/archive/checklist-completed-items.md`.

## Product Scope

- [ ] Define core mobile jobs-to-be-done and success metrics (activation, 7-day retention, recipe saves, repeat sessions).
  Model: `FREE_BALANCED`
- [ ] Define offline expectations (read-only favorites, cached recipe detail, queued actions).
  Model: `FREE_BALANCED`
- [ ] Define notification strategy (reminders, re-engagement, moderation/admin alerts if needed).
  Model: `FREE_FAST`

## Mobile UX and Navigation

- [ ] Define mobile-first IA (top app bar + bottom nav + modal filter sheets).
  Model: `FREE_BALANCED`
- [ ] Define touch target, spacing, and gesture rules for all high-frequency actions.
  Model: `FREE_FAST`
- [ ] Define empty/loading/error states for mobile screens with recovery actions.
  Model: `FREE_FAST`

## Design System Parity

- [ ] Map web tokens to mobile token set (color, typography, spacing, radius, elevation).
  Model: `FREE_BALANCED`
- [ ] Create shared component spec for buttons, chips, cards, inputs, list rows, and status badges.
  Model: `FREE_BALANCED`
- [ ] Run accessibility pass for contrast, dynamic text, and screen reader labels on mobile.
  Model: `CODEX_HIGH`

## Mobile Architecture and Data

- [ ] Define API usage policy for mobile (cache-first reads, retry/backoff, auth refresh behavior).
  Model: `CODEX_HIGH`
- [ ] Define feature-flag strategy for staged rollout and kill switches.
  Model: `FREE_BALANCED`
- [ ] Define analytics event taxonomy and funnel instrumentation for mobile journeys.
  Model: `FREE_BALANCED`

## Android Build and Release Engineering

- [ ] Finalize package ID, signing key strategy, and secure key storage policy.
  Model: `CODEX_HIGH`
- [ ] Define CI build lane for internal, beta, and production tracks.
  Model: `FREE_BALANCED`
- [ ] Define versioning policy (`versionCode`/`versionName`) and release notes workflow.
  Model: `FREE_FAST`

## iOS Build and Release Engineering

- [ ] Finalize iOS bundle ID, signing certificates/profiles, and secure keychain handling.
  Model: `CODEX_HIGH`
- [ ] Define CI build lane for TestFlight internal/external testing and App Store production release.
  Model: `FREE_BALANCED`
- [ ] Define iOS versioning policy (`CFBundleShortVersionString`/`CFBundleVersion`) and release notes workflow.
  Model: `FREE_FAST`

## Play Store Readiness

- [ ] Prepare store assets (icon, feature graphic, screenshots, optional promo video).
  Model: `FREE_FAST`
- [ ] Prepare privacy policy/data safety disclosures and complete content rating questionnaire.
  Model: `CODEX_HIGH`
- [ ] Run closed testing track and collect pre-launch report before production submission.
  Model: `FREE_BALANCED`

## Apple App Store Readiness

- [ ] Prepare App Store Connect assets (app icon, screenshots by device class, preview videos if used).
  Model: `FREE_FAST`
- [ ] Prepare App Privacy details, data collection declarations, and required policy URLs.
  Model: `CODEX_HIGH`
- [ ] Run TestFlight cycle, finalize metadata, and submit with release checklist for review.
  Model: `FREE_BALANCED`

## Quality, Reliability, and Operations

- [ ] Define device/version QA matrix and regression checklist.
  Model: `FREE_BALANCED`
- [ ] Define crash/performance monitoring and alert thresholds.
  Model: `FREE_BALANCED`
- [ ] Define staged rollout plan, rollback criteria, and post-launch triage cadence.
  Model: `CODEX_HIGH`
