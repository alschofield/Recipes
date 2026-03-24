# Security Pre-Audit Report (2026-03-23)

Status: internal pre-audit. This is not a substitute for independent security testing.

## Scope

- Server (`server/`)
- Web (`web/`)
- Mobile scaffolds (`mobile/android-native`, `mobile/ios-native`)
- CI/workflow and launch-gate docs

## Methods Used

- Secret-pattern scan across source/docs/config files.
- Web production dependency vulnerability scan (`pnpm audit --prod --json`).
- Go vulnerability scan (`govulncheck ./...`).
- Runtime and preflight checks from generated reports:
  - `docs/ops/v1-local-preflight-latest.md`
  - `docs/ops/v1-runtime-smoke-latest.md`

## Findings Summary

### Green

- No obvious hardcoded credential patterns detected by baseline regex sweep.
- Web production dependency audit reported no known vulnerabilities.
- Local preflight pipeline is repeatable and currently passing.
- Runtime smoke checks for web routes and Android emulator path are passing.

### Yellow

- Go standard library vulnerabilities reported by `govulncheck` for local toolchain level (`go1.25.5`), with fixed versions available in newer patch levels.
  - `GO-2026-4601` (`net/url`)
  - `GO-2026-4341` (`net/url`)
  - `GO-2026-4340` (`crypto/tls`)
  - `GO-2026-4337` (`crypto/tls`)

### Red

- None observed from automated checks in this pass.

## Priority Remediation

1. Upgrade Go runtime/toolchain patch level in local + CI execution environments to at least the fixed patch level reported by `govulncheck`.
2. Re-run `govulncheck ./...` and attach clean or residual report to `docs/ops/v1-launch-blocker-evidence.md`.
3. Execute external security review for auth/session and deployment posture before final launch.

## External Security Audit Required Areas

- Auth/session abuse scenarios (refresh rotation, session family revoke semantics).
- API input validation and authorization boundary checks.
- Mobile session/token storage and voice-permission pathways.
- CI/deploy secret handling and rollback controls.
- Cloud/network posture in deployed environment (TLS, CORS, headers, rate limits, logging).

## Evidence References

- `docs/ops/v1-local-preflight-latest.md`
- `docs/ops/v1-runtime-smoke-latest.md`
- `docs/ops/v1-launch-blocker-evidence.md`
- `docs/ops/v1-open-gates-snapshot.md`
