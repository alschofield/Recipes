# Mobile Release Engineering v1

## Android

Workflow scaffold: `.github/workflows/mobile-android.yml`

### Package IDs

- Dev: `com.recipes.mobile.dev`
- Prod: `com.recipes.mobile`

### Signing strategy

- One Play App Signing key for production lane.
- Upload key used in CI only via secure secret storage.
- No keystore files committed to repo.

### Key storage policy

- Store upload keystore in secret manager (encrypted blob).
- Store passwords/aliases as separate secrets.
- CI decrypts into ephemeral workspace and deletes post-build.

### CI lanes

- Internal lane: `devDebug` artifact for internal QA.
- Beta lane: `prodRelease` uploaded to closed testing.
- Production lane: `prodRelease` uploaded to production track after gate approval.

Release secret for Play upload:

- `ANDROID_PLAY_SERVICE_ACCOUNT_JSON`

### Versioning policy

- `versionCode`: monotonically increasing integer per release.
- `versionName`: semantic version (`MAJOR.MINOR.PATCH`).
- Release notes required for beta and production tracks.

## iOS

Workflow scaffold: `.github/workflows/mobile-ios.yml`

### Bundle identifiers

- Dev: `com.recipes.mobile.dev`
- Prod: `com.recipes.mobile`

### Signing strategy

- Apple Distribution certificate for release builds.
- Provisioning profiles per environment/bundle id.
- Fastlane/App Store Connect API key used for upload automation.

### Keychain/security policy

- Store signing certs/profiles in secure CI secret store.
- Import into temporary keychain during CI job only.
- Delete temporary keychain at job end.

### CI lanes

- Internal TestFlight lane.
- External TestFlight beta lane.
- App Store production submission lane.

Signing secret placeholders for release lane:

- `IOS_BUILD_CERT_BASE64`
- `IOS_BUILD_CERT_PASSWORD`
- `IOS_BUILD_PROVISION_BASE64`

### Versioning policy

- `CFBundleShortVersionString`: semantic marketing version (`MAJOR.MINOR.PATCH`).
- `CFBundleVersion`: build number incremented every CI release build.
- Release notes required for TestFlight and App Store submission.
