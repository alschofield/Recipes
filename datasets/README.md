# Datasets

Shared dataset storage used for enrichment priors, calibration inputs, and training experiments.

- `raw/` - immutable upstream/source drops.
- `derived/` - reproducible transforms generated from `raw/` sources.

Current migration target:

- `raw/server-lib/` contains former `server/lib/*` raw sources.
- `derived/server-lib/` contains former `server/lib/derived/*` artifacts.

Do not commit large transient artifacts; keep `.gitignore` aligned at repo root.
