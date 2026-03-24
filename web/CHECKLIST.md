# Web Checklist (V1 Blockers Only)

Goal: track only external/manual blockers that cannot be fully completed in-repo.

Only active/open blockers are listed here. Completed items are archived in `../docs/archive/checklist-completed-items.md`.

## V1 Product Quality Gates (Web)

- [ ] Run post-implementation usability/accessibility sanity pass and fix all P0/P1 web UX defects.
  Model: `CODEX_HIGH`

## External Validation and Launch Gates

- [ ] Run Lighthouse/perf sanity check on recipes index + detail in a compatible browser/runtime environment.
  Model: `FREE_BALANCED`
- [ ] Validate production web smoke in deployed environment (auth, search, detail, favorites) with owner signoff.
  Model: `FREE_BALANCED`

## Account/Platform Dependencies

- [ ] Observe first passing run of `.github/workflows/web-perf-audit.yml` in GitHub Actions and attach run URL to launch evidence.
  Model: `FREE_BALANCED`
