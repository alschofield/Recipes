# Mobile Architecture and Data Policy v1

## API usage baseline

- Source of truth: `api-usage-policy-v1.md`.
- Keep API clients thin and contract-driven.
- Treat server reconciliation responses as canonical state after offline replay.

## Feature-flag strategy

- Remote flags with local fallback defaults.
- Flag classes:
  - release gate flags (hide incomplete flows)
  - kill switches (disable unstable feature quickly)
  - experiment flags (A/B variants)
- All flags must have owner and expiry/review date.

## Analytics taxonomy (v1)

Core funnel events:

- `mobile_signup_completed`
- `mobile_first_search_submitted`
- `mobile_recipe_detail_viewed`
- `mobile_favorite_added`
- `mobile_paid_wall_viewed` (future paid lane)

Operational events:

- `mobile_refresh_token_success`
- `mobile_refresh_token_failure`
- `mobile_offline_queue_replayed`
- `mobile_offline_queue_replay_failed`

Event requirements:

- event name, timestamp, app version, platform, session id
- avoid PII and raw token payloads
