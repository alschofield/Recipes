# Native Scaffold Status (2026-03-23)

## Implemented scaffolds

- Android native Compose baseline in `mobile/android-native/RecipesMobile`.
- iOS native SwiftUI baseline in `mobile/ios-native/RecipesMobile` (XcodeGen template).

## Feature slice started

- Search tab now includes live `/recipes/search` request wiring in both native lanes:
  - Android: `SearchTabScreen` + `RecipesApiClient`.
  - iOS: `SearchView` + `RecipesAPIClient`.

- Favorites tab now includes API wiring + offline queue replay baseline in both native lanes:
  - Android: `FavoritesTabScreen` + local SharedPreferences queue.
  - iOS: `FavoritesView` + local UserDefaults queue.

Current behavior:

- ingredients input
- strict/inclusive mode select
- complex toggle
- network request and top-result rendering
- favorites list/load/add/remove and queued replay controls

## Completed implementation slice

- Auth/session bootstrap flow is now wired on both native clients:
  - login with stable client session ID
  - secure local session persistence (Android encrypted preferences, iOS keychain)
  - refresh token rotation
  - logout current session and logout all sessions
  - active session listing

- Favorites queue hardening is now wired on both native clients:
  - persisted queue timestamps
  - conservative dedup/cancel logic for obvious add/remove no-op pairs
  - replay status messaging after queue sync and reconciliation fetch

- V1 voice input baseline is wired for search ingredients on both native clients:
  - Android `RecognizerIntent` path with runtime mic permission and fallback messaging
  - iOS `Speech` framework path with mic/speech permission gating and fallback messaging

- V1 naming/copy and UI polish pass completed across both native clients:
  - app display name aligned to `Ingrediential`
  - bottom navigation labels aligned to `Discover`, `Saved`, `Profile`
  - core screen headings/copy updated for clearer user outcomes
  - Android Material theme + top app bar polish and icon-based navigation
  - iOS grouped list styling + app tint/background alignment

## Next implementation slice

- Accessibility implementation baseline is now in place (headings/sensitive-input handling/status messaging) and tracked in `../mobile/native-accessibility-pass-2026-03-23.md`.
- Execute manual TalkBack/VoiceOver device pass using `native-accessibility-manual-test-sheet-v1.md` and complete contrast/text-scaling validation.
- Execute store-readiness artifacts/testing cycles from `mobile/CHECKLIST.md`.
- Validate and harden Voice I/O V1 on real devices (`voice-io-plan-v1-v2.md`, `voice-privacy-policy-v1.md`).
- Execute STT hardening signoff using `voice-stt-validation-matrix-v1.md`.
