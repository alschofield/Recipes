# iOS Native Lane

Platform stack: Swift + SwiftUI.

## Bootstrap

1. Install XcodeGen (`brew install xcodegen` recommended).
2. Generate project from template:

```bash
cd mobile/ios-native/RecipesMobile
xcodegen generate
```

3. Open generated `RecipesMobile.xcodeproj` in Xcode.
4. Configure bundle identifiers per environment:
   - `com.recipes.mobile.dev` (debug/dev)
   - `com.recipes.mobile` (production)

## Required modules (initial)

- `Core/API` (HTTP client + auth refresh handling)
- `Core/Storage` (secure token storage + lightweight cache)
- `Features/Search`
- `Features/Favorites`
- `Features/Account`

## API contract references

- `../../docs/mobile/api-usage-policy-v1.md`
- `../../docs/server/search-contract.md`
- `../../docs/server/mobile-api-baseline.md`
