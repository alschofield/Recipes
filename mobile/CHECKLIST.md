# Mobile App Roadmap (UX + Launch Quality)

Goal: ship polished native UX and clear remaining store/manual launch gates.

Only active/open blockers are listed here. Completed items should be moved to `../docs/archive/checklist-completed-items.md`.

## V1 Product Quality Gates (Mobile)

- [ ] Run post-implementation UX sanity pass and fix all P0/P1 mobile UX defects before store submission.
  Model: `CODEX_HIGH`
- [ ] Build and maintain a mobile "weird issues" triage board (navigation, state sync, queue edge cases, auth/session friction) and close P0/P1 first.
  Model: `FREE_BALANCED`
- [ ] Improve native interaction polish (visual hierarchy, spacing consistency, loading/empty/error transitions, voice state feedback).
  Model: `CODEX_HIGH`
- [ ] Decide and execute third-party mobile UI/UX design audit if quality bar is not consistently met in manual QA.
  Model: `FREE_BALANCED`

## Manual Device Validation Gates

- [ ] Run Android TalkBack manual pass using `docs/mobile/native-accessibility-manual-test-sheet-v1.md` and file defects.
  Model: `CODEX_HIGH`
- [ ] Execute Android STT matrix on real devices using `docs/mobile/voice-stt-validation-matrix-v1.md` (permission states, no recognizer devices, cancellation, degraded network).
  Model: `CODEX_HIGH`
- [ ] Run iOS VoiceOver manual pass using `docs/mobile/native-accessibility-manual-test-sheet-v1.md` and file defects.
  Model: `CODEX_HIGH`
- [ ] Execute iOS STT matrix on real devices using `docs/mobile/voice-stt-validation-matrix-v1.md` (speech+mic permission combinations, interruption, locale mismatch).
  Model: `CODEX_HIGH`

## Store and Compliance Gates

- [ ] Finalize Play Store and App Store asset packs/screenshots and fill remaining owner-provided fields in `docs/mobile/store-submission-pack-template-v1.md`.
  Model: `FREE_FAST`
- [ ] Complete privacy/data safety declarations for mobile voice usage and validate policy URLs in both store consoles.
  Model: `CODEX_HIGH`
- [ ] Run Play closed testing + TestFlight cycles and capture blocking findings for go/no-go decision.
  Model: `FREE_BALANCED`

## Account and Release Credentials

- [ ] Provide release signing/material secrets and verify CI lanes for Android and iOS release artifacts.
  Model: `FREE_BALANCED`
- [ ] Capture final launch approval with owner signoff after blocker triage.
  Model: `FREE_BALANCED`
