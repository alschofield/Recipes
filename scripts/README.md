# Developer Bootstrap Scripts

These scripts help contributors install or verify required local tooling.

- `bootstrap.sh` - macOS/Linux bootstrap helper
- `bootstrap.ps1` - Windows bootstrap helper
- `changelog_draft.py` - generates a curated changelog markdown draft from recent non-merge commits
- `v1_preflight.py` - runs local V1 server/web/mobile preflight checks and writes a markdown report
- `v1_open_gates_snapshot.py` - generates consolidated open-gates snapshot from all checklist files
- `v1_runtime_smoke.py` - runs local runtime smoke checks (web routes, optional Android adb checks)
- `v1_gate_dashboard.py` - generates a compact launch dashboard from latest generated reports

Usage:

```bash
# macOS/Linux
bash scripts/bootstrap.sh

# Changelog draft (prints to stdout)
python scripts/changelog_draft.py --max 30

# Write draft to file
python scripts/changelog_draft.py --max 30 --out docs/changelog-draft.md

# V1 local preflight (writes docs/ops/v1-local-preflight-latest.md)
python scripts/v1_preflight.py

# V1 local preflight with explicit Android JAVA_HOME
python scripts/v1_preflight.py --android-java-home "/c/Program Files/Eclipse Adoptium/jdk-17.0.18.8-hotspot"

# V1 open gates snapshot (writes docs/ops/v1-open-gates-snapshot.md)
python scripts/v1_open_gates_snapshot.py

# V1 runtime smoke (web routes)
python scripts/v1_runtime_smoke.py

# V1 runtime smoke with Android adb checks
python scripts/v1_runtime_smoke.py --with-android

# V1 gate dashboard (writes docs/ops/v1-gate-dashboard-latest.md)
python scripts/v1_gate_dashboard.py
```

```powershell
# Windows PowerShell
powershell -ExecutionPolicy Bypass -File .\scripts\bootstrap.ps1
```

After bootstrap:

```bash
task setup
# or
make setup
```
