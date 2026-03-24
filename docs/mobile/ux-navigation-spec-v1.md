# Mobile UX and Navigation Spec v1

## Information architecture

Primary tabs:

1. Search
2. Favorites
3. Account

Navigation pattern:

- top app bar for page context + quick actions
- bottom nav for primary sections
- modal bottom sheet for advanced filters

## Touch and spacing rules

- minimum touch target: 44x44 pt (iOS), 48x48 dp (Android)
- standard spacing scale: 4/8/12/16/24
- primary actions placed in thumb-friendly zone where possible

## Empty/loading/error states

- Empty search results: offer mode switch and ingredient broadening suggestion.
- Loading: skeleton cards for search and detail.
- Error: inline retry action + short reason; do not block app-level navigation.
- Offline: explicit banner and stale-data label when using cache.
