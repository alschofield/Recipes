# Contributing

Thanks for contributing to Recipes.

## Development workflow

1. Create a branch from `main`:

```bash
git checkout -b feat/<short-topic>
```

2. Run local checks before opening a PR:

```bash
make check
make web-e2e
```

Task equivalents:

```bash
task check
task web-e2e
```

3. Keep PRs scoped (single concern per PR where possible).

## Coding guidelines

- Keep backend changes in `server/` and web changes in `web/`.
- Use explicit error handling; do not silently ignore failures.
- Keep migrations additive and reviewable.
- Preserve API response shapes unless intentionally versioning/changing behavior.
- Prefer SSR-first patterns in the web app unless interactivity requires client components.

## Commit message style

Use conventional, scope-oriented messages:

```text
<type>(<scope>): <why-focused summary>
```

Examples:

- `fix(web): avoid hydration mismatch in auth navigation`
- `feat(server): add admin moderation endpoint for ingredient candidates`

## Pull requests

Include in PR description:

- What changed
- Why it changed
- How it was tested
- Any deployment or migration impact

If your PR includes migrations, call that out explicitly.
