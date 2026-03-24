# Legal and Compliance Pre-Audit Report (2026-03-23)

Status: internal pre-audit. Not legal advice.

## Scope

- OSS license exposure for runtime web dependencies.
- Dataset provenance/legal blocker posture in LLM checklist.
- Privacy and store-declaration readiness for mobile voice input.

## Methods Used

- Web production license inventory (`npx --yes license-checker --production --json`).
- Review of active legal/compliance blockers in `llm/CHECKLIST.md`.
- Review of store/privacy policy docs:
  - `docs/mobile/voice-privacy-policy-v1.md`
  - `docs/mobile/store-submission-pack-rc0-2026-03-23.md`
  - `docs/ops/v1-external-inputs-checklist.md`

## Findings Summary

### Green

- Web runtime dependencies currently inventory as MIT for core framework packages (Next/React stack).
- Mobile voice feature policy baseline exists and explicitly states no raw-audio persistence in app logic.

### Yellow

- LLM lane still has explicit legal blocker items (dataset lane legal status + license/compliance approval pending).
- Store/legal inputs remain incomplete (policy URLs, support/marketing URLs, declaration finalization).
- Full transitive OSS attribution package (all services/languages) is not yet assembled in a single release artifact.

### Red

- None proven yet by automation, but launch legal signoff is blocked until domain/store/privacy fields and LLM legal gates are cleared.

## Priority Remediation

1. Finalize policy URLs and owner-provided legal/store fields in `docs/mobile/store-submission-pack-rc0-2026-03-23.md`.
2. Obtain legal review for LLM dataset lanes and close blocked items in `llm/CHECKLIST.md`.
3. Produce a release-ready attribution bundle for all distributed components before public launch.

## External Legal Review Required Areas

- Privacy policy and disclosure accuracy vs actual data handling (web/mobile/voice telemetry).
- App Store and Play declaration completeness.
- LLM dataset rights, attribution obligations, and commercial-use constraints.
- OSS attribution obligations and notice packaging for release artifacts.

## Evidence References

- `llm/CHECKLIST.md`
- `docs/mobile/store-submission-pack-rc0-2026-03-23.md`
- `docs/mobile/voice-privacy-policy-v1.md`
- `docs/ops/v1-external-inputs-checklist.md`
