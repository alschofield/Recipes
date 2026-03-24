# Mobile API Baseline (V1)

Status: draft baseline for native mobile enablement.

## What is ready now

- JWT-protected APIs are already consumable by mobile clients:
  - auth (`signup/login/profile`)
  - recipes (`/recipes/search`, `/recipes/detail/:id`)
  - favorites CRUD
- Search contract is stable and documented in `search-contract.md`.
- Fallback reliability controls and observability are in place (`/recipes/health/llm`, `/recipes/metrics`, `/recipes/metrics/llm`).
- Refresh-token endpoints are now available:
  - `POST /users/refresh` (rotates refresh token + returns new access token)
  - `POST /users/logout` (revokes refresh-token family)
  - `POST /users/logout/session` (revokes current client session only)
  - `GET /users/{userid}/sessions` (lists active session scopes)

## Mobile gaps to close before production mobile rollout

### 1) Token/session lifecycle hardening

- Keep refresh-token flow and rotation as baseline.
- Add revoke-by-device/session identifier (not only family).
- Add explicit session listing/revoke contract for account security UX.

### 2) Retry + idempotency contract

- Favorites mutation endpoints now support idempotent retry semantics:
  - `POST /favorites/{userid}/{recipeid}` returns `201` on first write, `200` on replay with `Idempotency-Status: replayed`.
  - `DELETE /favorites/{userid}/{recipeid}` returns `204` both for first delete and replay; replay includes `Idempotency-Status: replayed`.
- Retry guidance for mobile clients:
  - safe retry: GET requests
  - safe retry: favorites POST/DELETE (idempotent contract)
  - guarded retry with `Idempotency-Key` is available for non-idempotent writes (for example signup) with server TTL window.

### 3) Offline sync safety for favorites

- Offline contract is defined in `favorites-sync-contract.md`:
  - queue/replay model
  - idempotent mutation semantics
  - reconciliation GET as source of truth

### 4) Mobile-oriented error envelope consistency

- Keep a stable machine-readable error envelope across services:

```json
{
  "error": "human-readable message",
  "code": "MACHINE_CODE",
  "details": {}
}
```

### 5) Security hardening for native clients

- Explicitly document token storage expectations (OS secure storage/keychain).
- Add mobile rate-limit policy for auth/search endpoints.
- Ensure CORS/origin setup stays web-specific and does not block mobile-native requests.

## Recommended implementation order

1. Refresh-token hardening (device/session revoke granularity).
2. Extend idempotency strategy beyond favorites to other mutation lanes.
3. Offline sync contract and reconciliation endpoint semantics.
4. Mobile-specific runbook notes (core refresh/retry/conflict tests are in place).

## Related docs

- `auth-security-baseline.md`
- `search-contract.md`
- `favorites-sync-contract.md`
- `../ops/operations-runbook.md`
- `../../mobile/CHECKLIST.md`
