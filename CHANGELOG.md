# Changelog

This project currently uses a curated, commit-based changelog (no tagged release history yet).

## Unreleased

### Fixed

- CI Docker build now uses the correct server build context to prevent pipeline failures (`470170b`).

### Changed

- Deployment workflows now skip deploy jobs when required deploy hooks are not configured, reducing false-fail CI runs (`add3238`).
- Project licensing policy updated to all-rights-reserved usage (`eb6d407`).

### Dependencies

- GitHub Actions updated: `actions/checkout` v6 (`2d72a18`), `actions/setup-node` v6 (`87c44a3`), `pnpm/action-setup` v5 (`9dd4f00`), `actions/setup-go` v6 (`b0b0a22`).
- Server dependency updated: `github.com/jackc/pgx/v5` to `v5.9.0` (`1ac0108`).

### Notes

- Some historical commits use non-standard messages (for example link-only/day-log entries: `4365088`, `21edb0b`, `f4a0595`) and are not expanded into product-level changelog bullets.

To refresh this section quickly:

```bash
git log --no-merges --date=short --pretty=format:"- %ad `%h` %s" -n 25
```

Or generate a grouped draft:

```bash
python scripts/changelog_draft.py --max 30 --out docs/changelog-draft.md
```
