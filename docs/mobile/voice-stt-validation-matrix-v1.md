# Voice STT Validation Matrix v1

Use this matrix to validate STT behavior before enabling voice input broadly.

## Run metadata

- RC/build:
- Tester:
- Date:
- Device/OS:
- Locale:

## Android scenarios

- [ ] First use -> allow microphone -> dictation returns transcript -> input merged correctly.
- [ ] First use -> deny microphone -> recovery message shown -> manual typing path remains usable.
- [ ] Microphone denied permanently -> repeat tap path remains non-blocking.
- [ ] Speech recognizer unavailable on device -> fallback message shown.
- [ ] User cancels recognizer sheet -> non-blocking cancel message shown.
- [ ] Network degraded/offline path (if recognizer requires network) -> failure message shown.
- [ ] Accessibility announcement fired for success and error states.

## iOS scenarios

- [ ] First use -> allow speech + microphone -> transcript captured and merged.
- [ ] Deny speech permission -> fallback message shown -> manual typing path remains usable.
- [ ] Deny microphone permission -> fallback message shown -> manual typing path remains usable.
- [ ] Speech recognizer unavailable/unsupported locale -> fallback message shown.
- [ ] User stops dictation manually -> clean stop with no stale listening state.
- [ ] Interruption during dictation (call/audio route change) -> recognition stops cleanly.
- [ ] Accessibility announcement fired for voice errors.

## Data/privacy checks

- [ ] No raw audio files are persisted by app logic.
- [ ] No transcript text is written to telemetry counters.
- [ ] Voice counters increment only aggregate keys (start/success/failure/permission-denied).

## Result

- [ ] Android pass
- [ ] iOS pass
- [ ] Blockers filed and linked
