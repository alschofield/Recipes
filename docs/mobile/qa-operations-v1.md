# Mobile QA, Reliability, and Operations v1

## Device/version QA matrix

### iOS

- iPhone SE (small screen), latest iOS
- iPhone 14/15 (standard), latest iOS
- iPhone Pro Max (large screen), latest iOS

### Android

- Pixel mid-tier (Android latest)
- Samsung Galaxy mainstream (latest and n-1)
- One lower-memory Android device profile

### Regression checklist (per release)

- Auth login/refresh/logout paths
- Search strict/inclusive + fallback behaviors
- Favorites add/remove + offline replay
- Session revoke + re-auth behavior
- Crash-free launch and cold-start sanity

## Crash/performance monitoring

- Crash monitor: Firebase Crashlytics (or Sentry mobile) required before public release.
- Performance monitor: startup and network latency traces.

Alert thresholds:

- crash-free users < 99.5%
- p95 cold start regression > 20% vs last stable release
- auth refresh failure rate > 2%

## Staged rollout and triage cadence

- Android production rollout: 5% -> 20% -> 50% -> 100%.
- iOS rollout: phased release enabled after App Store approval.
- Rollback criteria:
  - crash-free users below threshold
  - auth/session failures exceed threshold
  - major data sync regression

Post-launch cadence:

- first 48 hours: twice-daily triage
- week 1: daily triage
- week 2+: weekly reliability review
