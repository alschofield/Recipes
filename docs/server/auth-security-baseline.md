# Auth and Security Baseline

Status: v1 locked baseline (update via checklist preflight item)

## Authentication Strategy

- Access token: short-lived JWT (default 15 minutes, configurable via env).
- Refresh token: rotating refresh-token flow enabled (`/users/refresh`) with session revoke (`/users/logout/session`) and family revoke support (`/users/logout`).
- Password hashing: bcrypt with strong cost.
- Password minimums: length >= 12, block common weak patterns.

## Authorization Rules

- Roles: `user`, `admin`.
- `user` can read/update/delete only self resources.
- `admin` can perform all user operations.
- Favorites endpoints enforce ownership unless role is `admin`.

## Route Guarding

- Public: signup/login/health.
- Auth required: profile and favorites.
- Admin required: any cross-user management route.

## Minimum Security Controls

- Rate limiting: login, signup, password reset, search.
- CORS: explicit origin allowlist per environment.
- Secure headers: HSTS, X-Content-Type-Options, X-Frame-Options, Referrer-Policy.
- Request size limits for JSON and upload endpoints.
- Structured error responses without leaking internals.

## Session and Token Handling

- Revoke refresh token on logout.
- Rotate refresh token each use; store token family ID server-side.
- Invalidate all sessions on password change (remaining hardening task).
- Mobile clients should provide stable `X-Client-Session-ID` to scope device/session revocation.

## API Error Contract

```json
{
  "error": "human-readable message",
  "code": "MACHINE_CODE",
  "details": {}
}
```

## Audit and Observability

- Log auth attempts (success/failure) with request IDs.
- Track rate-limit and forbidden access events.
- Never log passwords, raw tokens, or secrets.

## Testing Expectations

- Unit tests for password hashing/verification and token validation.
- Integration tests for login/signup/profile/favorites authorization matrix.
- Negative tests: expired token, invalid token, forbidden role access.
