# Security Policy

## Reporting a vulnerability

Please do not open public GitHub issues for security vulnerabilities.

Report privately to the maintainers with:

- Affected component (`server`, `web`, infra, etc.)
- Reproduction steps or proof of concept
- Impact assessment
- Suggested mitigation (if available)

We will acknowledge valid reports promptly and coordinate a fix/release.

## Secrets handling

- Never commit real `.env` files, tokens, credentials, or private keys.
- Use environment-specific secret stores for staging/production.
- Rotate any credential immediately if it is accidentally exposed.

## Baseline hardening expectations

- Keep dependencies updated and review CI alerts.
- Enforce auth/authorization checks at service boundaries.
- Validate and sanitize all external input.
- Use least-privilege credentials for DB and deployment systems.
