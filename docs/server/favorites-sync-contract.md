# Favorites Sync Contract (Mobile Offline)

Status: v1 baseline contract for offline queue + reconciliation.

## Endpoints in scope

- `GET /favorites/{userid}`
- `POST /favorites/{userid}/{recipeid}`
- `DELETE /favorites/{userid}/{recipeid}`

## Mutation semantics

- `POST` add favorite:
  - first execution: `201`
  - replay/safe retry: `200` with `Idempotency-Status: replayed`
- `DELETE` remove favorite:
  - first execution: `204`
  - replay/safe retry when already absent: `204` with `Idempotency-Status: replayed`

These semantics allow mobile clients to replay queued actions after reconnect without duplicate-state risk.

## Offline queue model (client)

- Queue actions as `(op, recipeId, queuedAt)` where `op` is `add` or `remove`.
- Preserve order by queue time.
- Deduplicate obvious no-op pairs before flush (for example `add` then `remove` same `recipeId` before sync).

## Reconciliation strategy

1. Flush queued actions with retries.
2. After flush completes, call `GET /favorites/{userid}`.
3. Treat server response as source of truth.
4. Replace local favorites cache with server list.

## Conflict behavior

- If the same recipe is changed from multiple devices, last successful mutation on server wins.
- Because POST/DELETE are idempotent, replayed operations should converge after final reconciliation GET.

## Error handling guidance

- `401/403`: stop sync and force re-auth.
- `404` on GET user scope: treat as auth/account issue.
- `5xx` or network timeout: retry with exponential backoff.

## Future improvements

- Add optional server-issued sync cursor/version for larger favorite sets.
- Add batch mutation endpoint for lower mobile round-trip overhead.
