# Web Lighthouse Sanity Attempt (2026-03-23)

Target checklist item: run Lighthouse/perf sanity on recipes index + detail after modernization.

## Attempted commands

- Build and start web app locally.
- Run Lighthouse for:
  - `/recipes`
  - `/recipes/placeholder`

## Outcome

- Lighthouse execution is currently blocked in this environment.
- CLI error indicates no usable Chrome installation or launcher compatibility issue:
  - `No Chrome installations found`
  - Playwright Chromium path also fails for Lighthouse launcher (`Failed to load Chrome DLL ... (0x57)`).

## Next unblock step

- Install a compatible local Chrome/Chromium build detectable by Lighthouse launcher and re-run the same commands.
- Once browser runtime is available, store JSON artifacts under `web/test-results/`.
