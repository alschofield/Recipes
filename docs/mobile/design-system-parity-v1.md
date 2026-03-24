# Mobile Design System Parity v1

## Token mapping (web -> mobile)

- Color tokens: keep semantic names (`bg`, `surface`, `text-primary`, `accent`, `success`, `error`).
- Typography: map hierarchy levels (`display`, `title`, `body`, `caption`) to platform-native text styles.
- Spacing/radius/elevation: preserve numeric scale parity where practical.

## Core component spec (shared behavior, native implementation)

- Buttons: primary / secondary / chip variants
- Chips: filter + metadata chips
- Cards: recipe card, status card
- Inputs: ingredient composer input + search input
- List rows: favorites list, account/session list
- Status badges: database vs generated, confidence/reviewability hints

## Accessibility baseline for implementation

- dynamic text enabled by default
- semantic labels for icon-only actions
- contrast checks before release candidate builds
- keyboard/switch control parity on both platforms
