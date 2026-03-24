# Mobile API Usage Policy v1

Status: baseline policy for native clients (Android/iOS).

## Scope

Applies to mobile clients calling:

- `/users/*` (auth/profile/session)
- `/recipes/*` (search/detail)
- `/favorites/*` (list/mutate)

## 1) Auth and session behavior

- Use short-lived access token for normal API calls.
- Use refresh token via `POST /users/refresh` before/after access token expiry.
- Include a stable `X-Client-Session-ID` per installed app session/device profile.
- For account security UX:
  - list sessions: `GET /users/{userid}/sessions`
  - revoke current session: `POST /users/logout/session`
  - revoke session family: `POST /users/logout`

## 2) Cache policy

- Cache-first reads for recipe detail and favorites list with short TTL.
- Recommended local TTLs:
  - recipe detail: 10-30 minutes
  - favorites list: 1-5 minutes
- Always refresh from network when user pulls to refresh.
- On reconnect, reconcile local state using server response as source of truth.

## 3) Retry and backoff policy

- Safe retries:
  - all `GET` requests
  - favorites mutations (`POST/DELETE /favorites/{userid}/{recipeid}`) due to idempotent server semantics
- For non-idempotent writes, send `Idempotency-Key` header.
- Retry windows:
  - `408/429/5xx` and network timeout: exponential backoff with jitter
  - `401/403`: stop retries; refresh/re-auth path
  - `4xx` validation errors: do not retry blindly

Suggested backoff: `1s`, `2s`, `4s`, `8s` + jitter, max 4 attempts.

## 4) Offline behavior (favorites)

- Queue offline favorite add/remove actions.
- Replay in order when online.
- After replay, call `GET /favorites/{userid}` and replace local cache.
- Server semantics support replay:
  - add replay: `200` + `Idempotency-Status: replayed`
  - delete replay: `204` + `Idempotency-Status: replayed`

## 5) Error handling contract

Expect server error envelope:

```json
{
  "error": "human-readable message",
  "code": "MACHINE_CODE",
  "details": {}
}
```

Client handling:

- surface `error` for user-safe messages
- route behavior by `code`
- log `request-id` when present in response headers/log context

## 6) Telemetry expectations

Track at minimum:

- auth refresh success/failure rate
- search request latency + failure buckets
- favorites queue length and replay success rate
- fallback engagement rate (recipes from `source=llm`)
- voice STT start/success/failure/permission-denied counters

## 7) Security requirements

- Store tokens only in platform secure storage (Keychain/Keystore).
- Never log raw access or refresh tokens.
- Use TLS-only endpoints in non-local environments.

## Related contracts

- `../server/mobile-api-baseline.md`
- `../server/favorites-sync-contract.md`
- `../server/auth-security-baseline.md`
- `../server/search-contract.md`
