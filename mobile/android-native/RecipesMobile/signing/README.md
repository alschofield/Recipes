# Android Signing Assets (Do Not Commit Secrets)

Expected local/CI files (not committed):

- `upload-keystore.jks`
- `signing.properties` (keystore path, alias, passwords)

Use secure secret storage for CI and inject at build time.
