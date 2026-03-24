# Voice I/O Plan (V1 -> V2)

Status: planning baseline. Implementation tasks are tracked in `../../mobile/CHECKLIST.md`.

## Goals

- Reduce input friction with speech-to-text (STT) in high-frequency entry points.
- Add optional hands-free cooking playback with text-to-speech (TTS) in later phase.
- Keep privacy and permission behavior explicit and conservative.

## V1 (STT Input)

Scope:

- Tap-to-speak for search ingredients input.
- Tap-to-speak for manual ingredient entry fields.
- User-triggered only (no always-on listening).
- No background/lock-screen audio behavior in V1.

Platform approach:

- Android: `SpeechRecognizer` (or Activity Result speech intent fallback).
- iOS: `Speech` framework with explicit authorization checks.

UX requirements:

- Mic button near text inputs with clear "listening" state.
- Live/partial transcript where available.
- Manual edit always available before submit.
- Clear error states: permission denied, unavailable locale, network failure.

Privacy baseline:

- Request microphone/speech permissions only when voice input is invoked.
- Do not persist raw audio.
- Persist only final text user accepts/submits.
- Add policy text in app privacy disclosures and in-store metadata.

Telemetry baseline:

- Track local counters for STT start/success/failure/permission-denied.
- Keep counters aggregate-only and transcript-free.

## V2 (TTS Guided Cooking)

Scope:

- Read recipe steps aloud.
- Controls: play, pause, next, previous, stop.
- Step index persistence during active cooking session.
- Optional lock-screen/background audio behavior (explicitly tested).

Platform approach:

- Android: platform TTS engine + audio focus handling.
- iOS: `AVSpeechSynthesizer` + interruption handling.

UX requirements:

- Voice rate and optional language/voice selection.
- Highlight current step while speaking.
- Immediate resume/replay for current step.
- Safe behavior on interruptions (calls, media, route changes).

## Cross-cutting QA and Accessibility

- Verify TalkBack/VoiceOver labels and announcements for voice controls.
- Verify dynamic text and contrast for mic/playback controls.
- Validate denied-permission fallback path remains fully usable.
- Test interruption flows and app-background transitions.

## Non-goals (for now)

- Wake word/hotword activation.
- Continuous dictation across tabs.
- Audio recording upload or storage.
