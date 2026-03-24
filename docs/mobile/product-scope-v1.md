# Mobile Product Scope v1

## Core jobs-to-be-done

1. Turn pantry ingredients into actionable meal options quickly.
2. Save and revisit trusted recipes across sessions/devices.
3. Recover smoothly from poor network conditions.

## Success metrics (initial)

- Activation: signup to first successful recipe detail view.
- 7-day retention: users returning within 7 days.
- Save rate: sessions with at least one favorite mutation.
- Repeat sessions per user per week.

## Offline expectations

- Read cached recipe detail and favorites list when offline.
- Queue favorite add/remove actions offline.
- Replay queued actions on reconnect and reconcile with server truth.

## Notification strategy

- Week-1: no push by default (avoid noisy launch).
- Week-2+ experiments:
  - weekly meal reminder
  - dormant user re-engagement
  - optionally reminder tied to prior saved recipes

Guardrails:

- clear opt-in control
- no more than 2 promotional notifications/week by default
