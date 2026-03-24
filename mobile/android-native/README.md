# Android Native Lane

Platform stack: Kotlin + Jetpack Compose.

## Bootstrap

1. Open `mobile/android-native/RecipesMobile` in Android Studio.
2. Let Android Studio sync Gradle project.
3. Configure local API endpoint in `local.properties` (copy from `local.properties.example`).
4. Use product flavors for environments:
   - `devDebug`
   - `prodRelease`

## Required modules (initial)

- `core:network` (HTTP client + auth refresh handling)
- `core:storage` (encrypted token storage + cache)
- `feature:search`
- `feature:favorites`
- `feature:account`

## API contract references

- `../../docs/mobile/api-usage-policy-v1.md`
- `../../docs/server/search-contract.md`
- `../../docs/server/mobile-api-baseline.md`
