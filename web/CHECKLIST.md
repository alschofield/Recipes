# Web Checklist

Goal: deliver a premium, fast, trustworthy recipe UX that clearly communicates DB + LLM results and keeps complex mode understandable.

Only active/open work is listed here. Completed items are archived in `../docs/archive/checklist-completed-items.md`.

## Search UX and Controls

- [ ] Add explicit help text/tooltips explaining strict vs inclusive vs complex behavior (ref: `docs/server/search-contract.md`).
  Model: `FREE_BALANCED`
- [ ] Add clear source legend and confidence hints for mixed DB + LLM result sets.
  Model: `FREE_BALANCED`

## Result Presentation and Trust Signals

- [ ] Add blend explanation UI (`why this result`, source rationale, matched/missing clarity) (ref: `docs/server/search-contract.md`, `docs/llm/fallback-contract.md`).
  Model: `FREE_BALANCED`
- [ ] Add visual handling for generated-recipe reviewability (badge/status) without clutter.
  Model: `FREE_BALANCED`
- [ ] Add empty-state guidance when strict mode + fallback yields no results.
  Model: `FREE_FAST`

## UI Modernization (Moved from Root)

- [ ] Define sharper visual direction board (typography, spacing rhythm, surface system, color intent) (ref: `docs/product/design-modernization-v1.md`).
  Model: `FREE_BALANCED`
- [ ] Redesign top-level layout shell with stronger hierarchy and modern spacing.
  Model: `FREE_BALANCED`
- [ ] Replace baseline cards/buttons with cohesive component language and interaction states.
  Model: `FREE_BALANCED`
- [ ] Add motion system for reveals/state changes (minimal, meaningful, performant).
  Model: `FREE_BALANCED`
- [ ] Upgrade recipe results density controls and metadata chips.
  Model: `FREE_BALANCED`
- [ ] Improve pantry input flow (keyboard-first, fast add/remove, suggestion ergonomics).
  Model: `FREE_BALANCED`
- [ ] Improve recipe detail readability and print layout polish.
  Model: `FREE_BALANCED`

## Accessibility and Reliability

- [ ] Run full accessibility pass (contrast, focus order, labels, keyboard-only paths).
  Model: `CODEX_HIGH`
- [ ] Add robust loading/error/retry states for recipe search and detail views.
  Model: `FREE_FAST`
- [ ] Add frontend telemetry for search intent, complex toggle usage, and fallback result engagement.
  Model: `FREE_BALANCED`

## Release Readiness (Web)

- [ ] Verify production env wiring for API routes and auth flows in nginx-proxied deployment (ref: `docs/ops/deployment-plan.md`).
  Model: `FREE_BALANCED`
- [ ] Run Lighthouse/perf sanity check on recipes index + detail after UI modernization.
  Model: `FREE_BALANCED`
