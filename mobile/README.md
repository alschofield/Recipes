# Mobile Workspace (Native)

This workspace is native-first:

- `ios-native/` - Swift + SwiftUI app lane
- `android-native/` - Kotlin + Jetpack Compose app lane
- `shared/` - shared API/domain contract assets (non-runtime shared source)

Current scaffold status:

- Android baseline project scaffold in `android-native/RecipesMobile/`.
- iOS XcodeGen template scaffold in `ios-native/RecipesMobile/`.

Planning and policy docs live in `../docs/mobile/`.

Live endpoints used by mobile apps:

- API base URL: `https://api.ingrediential.uk`
- Public web URL: `https://www.ingrediential.uk`

Use this folder for native project files, platform-specific CI lanes, and mobile release artifacts.
