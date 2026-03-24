# iOS Signing Assets (Do Not Commit Secrets)

Expected CI-managed assets (not committed):

- Apple Distribution certificate (.p12)
- provisioning profiles (.mobileprovision)
- App Store Connect API key

Import signing assets into a temporary keychain during CI and delete afterward.
