# Remote DB Operator Scripts

Scripts in this folder apply remote migrations and verify seed/data integrity against hosted Postgres environments (for example Neon).

## Scripts

- `apply_remote_migrations.ps1` - apply all `server/migrations/*.up.sql` files in lexical order using `psql`.
- `apply_remote_migrations.sh` - same flow for macOS/Linux shells.
- `verify_remote_seed.ps1` - run seed/data quality checks and fail fast on regressions.
- `smoke_api_data_quality.py` - call `/recipes/catalog` and `/recipes/search` and validate response/data thresholds.

## Requirements

- `psql` in PATH
- `DATABASE_URL` set to the target database connection string

## Usage

```powershell
$env:DATABASE_URL = "postgresql://user:pass@host/db?sslmode=require&channel_binding=require"

# Optional: list migrations only
powershell -ExecutionPolicy Bypass -File .\scripts\db\apply_remote_migrations.ps1 -DryRun

# Apply migrations
powershell -ExecutionPolicy Bypass -File .\scripts\db\apply_remote_migrations.ps1

# Verify seed integrity and quality gates
powershell -ExecutionPolicy Bypass -File .\scripts\db\verify_remote_seed.ps1

# API-level data quality smoke report
python .\scripts\db\smoke_api_data_quality.py --base-url "https://api.ingrediential.uk"
```

```bash
export DATABASE_URL="postgresql://user:pass@host/db?sslmode=require&channel_binding=require"

# Optional: list migrations only
bash scripts/db/apply_remote_migrations.sh --dry-run

# Apply migrations
bash scripts/db/apply_remote_migrations.sh

# API-level data quality smoke report
python scripts/db/smoke_api_data_quality.py --base-url "https://api.ingrediential.uk"
```

## Verification thresholds

`verify_remote_seed.ps1` uses minimum thresholds and hard-fail checks:

- minimum rows: ingredients, recipes, aliases
- required fields: recipe name/steps, canonical ingredient names, alias values
- duplicate checks: canonical ingredient names and aliases (case-insensitive)
- relationship checks: recipes without ingredients
- source distribution: requires at least one `database` recipe

Override minimum thresholds when needed:

```powershell
$env:SEED_MIN_INGREDIENTS = "80"
$env:SEED_MIN_RECIPES = "20"
$env:SEED_MIN_ALIASES = "10"
powershell -ExecutionPolicy Bypass -File .\scripts\db\verify_remote_seed.ps1
```
