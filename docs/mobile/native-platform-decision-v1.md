# Native Platform Decision v1

Date: 2026-03-23

## Decision

Adopt **pure native mobile apps**:

- iOS: Swift + SwiftUI
- Android: Kotlin + Jetpack Compose

## Why

- Highest ceiling for platform polish and performance.
- Better alignment with App Store/Play Store native UX expectations.
- Stronger long-term maintainability for platform-specific capabilities.

## Tradeoff accepted

- Higher implementation cost vs cross-platform frameworks.
- Duplicate feature implementation effort across iOS and Android lanes.

## Scope implication

- This repo now treats mobile as two first-class native tracks.
- Shared artifacts remain at contract/policy level, not shared runtime UI code.
