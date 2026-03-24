# Voice Privacy Policy (Mobile V1)

Status: policy baseline for STT-only phase.

## Scope

Applies to voice input features in native mobile clients:

- search ingredient entry
- ingredient text-field dictation entry points

## Consent and permissions

- Ask for microphone/speech permissions only when the user taps a voice-input control.
- Provide a clear rationale before system prompts ("speak ingredients to fill input fields").
- If denied, keep full manual text-entry path available.

## Data handling

- Do not persist raw audio recordings.
- Do not upload/store raw audio in backend services owned by this project.
- Persist only final text accepted/submitted by the user.
- Do not log transcript text with auth/session identifiers in the same event payload.

## Retention and logging

- Voice session telemetry is aggregate-only (success/failure counters, permission-state counters).
- Error logs must redact transcript content and token/session fields.
- Client crash logs must avoid microphone/transcript payloads.

## User controls

- Voice input is opt-in per interaction (tap-to-speak).
- User can stop dictation immediately.
- User can edit transcript before submit.

## Security constraints

- TLS-only network environments.
- Existing token/keychain/keystore policies remain unchanged.
- No background listening in V1.

## Store disclosure alignment

- Reflect microphone/speech usage in Play/App Store privacy declarations.
- Keep privacy-policy URL and store disclosures consistent with this policy.
