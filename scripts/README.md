# Developer Bootstrap Scripts

These scripts help contributors install or verify required local tooling.

- `bootstrap.sh` - macOS/Linux bootstrap helper
- `bootstrap.ps1` - Windows bootstrap helper

Usage:

```bash
# macOS/Linux
bash scripts/bootstrap.sh
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
