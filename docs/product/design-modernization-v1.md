# Design Modernization v1

This document defines the macro UI direction for the Ingrediential web app so implementation decisions stay consistent across pages.

## Goals

1. Improve scan speed and information hierarchy.
2. Make admin and data-heavy pages easier to operate.
3. Preserve current brand warmth while modernizing structure and typography.
4. Create reusable tokens/primitives for consistent future UI work.

## Product-level design principles

- Clarity over novelty: each page should make the next action obvious.
- Progressive disclosure: show critical info first, details second.
- Consistent shells and patterns: same nav model, spacing rhythm, and component behavior across pages.
- Data readability first: optimize for list/table/card scanning and status recognition.
- Honest state design: loading, empty, partial, success, and failure states all get intentional UI.

## Layout architecture

### Desktop

- Left sidebar for primary nav and section grouping.
- Top utility bar for global search, quick actions, account controls.
- Main content in a 12-column fluid grid with max readable widths per section type.

### Mobile

- Compact top bar + collapsible nav drawer.
- Optional bottom quick nav for top destinations (`Recipes`, `Ingredients`, `Favorites`).
- Filters and side panels become modal sheets.

### Page macro template

1. Page header (title, short intent text, primary action)
2. KPI/status strip (when relevant)
3. Primary content modules (cards/table/list)
4. Secondary utilities (queue, logs, diagnostics)

## Typography system

Use two primary families plus a mono utility family.

- Heading/display: `Space Grotesk`
- Body/UI: `Plus Jakarta Sans`
- Mono/metrics: `IBM Plex Mono`

### Type scale (base 16)

- Display: 40/48
- H1: 32/40
- H2: 24/32
- H3: 20/28
- Body: 16/24
- Small: 14/20
- Micro/meta: 12/16

Use heavier weight only for semantic anchors (title, key metric, CTA).

## Color and token direction

Keep the existing warm accent family but formalize token usage.

### Core tokens (semantic)

- `--bg-app`
- `--bg-surface-1`
- `--bg-surface-2`
- `--text-primary`
- `--text-secondary`
- `--text-muted`
- `--border-subtle`
- `--border-strong`
- `--accent-primary`
- `--accent-primary-contrast`
- `--accent-soft`
- `--success`
- `--warning`
- `--danger`
- `--info`
- `--focus-ring`

### Usage rules

- Status colors indicate state, not decoration.
- Keep accent usage to action and emphasis points.
- Avoid flat single-tone pages; use subtle layered surfaces for depth.

## Spacing, radius, elevation

- 8-point spacing system.
- Radii: 8 (small), 12 (default), 16 (large).
- Elevation: subtle shadows only on interactive or overlay layers.

## Motion language

- Page enter: short fade/translate (120-180ms).
- List/card stagger only for first render in large lists.
- Interactive transitions: 100-150ms; no spring-heavy motion for data views.

## Component primitives

- App shell (sidebar/topbar/content)
- Page header block
- KPI card
- Data card
- Filter bar
- Table/list row with status badges
- Empty state panel
- Error state panel
- Loading skeleton set

All components must define hover, focus, active, disabled, and error states.

## Accessibility baseline

- WCAG AA contrast for text and controls.
- Keyboard navigable nav/filter/table interactions.
- Visible focus ring on all actionable controls.
- Semantic heading structure (`h1` -> `h2` -> `h3`).
- Proper labels/aria for filters and action controls.

## Performance guardrails

- Self-host and subset web fonts where possible.
- Limit initial font families/weights loaded above the fold.
- Avoid large client-only component bundles for static views.

## Page-specific modernization plan

### Recipes search

- Split header/filter/results clearly.
- Improve card readability and action grouping.
- Show rationale metadata (match, quality, time) in consistent order.

### Recipe detail

- Two-column desktop layout: recipe content + analysis/meta panel.
- Strong visual separation for ingredients, steps, and quality diagnostics.

### Ingredients

- Catalog-first UX (search/filter/paginate), suggestions as secondary panel.
- Better status and quality signal rendering for scanability.

### Admin pages

- Metrics strip first.
- Trend modules second.
- Operational queue (pending/stale items) last with fast action affordances.

## Implementation phases

1. Foundations: tokens, typography, shell, primitives.
2. Core surfaces: recipes search/detail and ingredients catalog.
3. Admin surfaces: moderation and analysis dashboards.
4. Polish pass: motion, responsive QA, accessibility/perf hardening.

## Working agreement

When implementing UI updates, use this document as the source of truth and avoid introducing one-off patterns that bypass these tokens and primitives.
