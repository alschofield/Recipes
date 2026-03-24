#!/usr/bin/env python3
"""Generate a compact V1 gate dashboard from generated reports."""

from __future__ import annotations

import argparse
import datetime as dt
import re
from pathlib import Path


def read_text(path: Path) -> str:
    if not path.exists():
        return ""
    return path.read_text(encoding="utf-8")


def extract_total_open(snapshot_text: str) -> int | None:
    match = re.search(r"Total open gates: \*\*(\d+)\*\*", snapshot_text)
    if not match:
        return None
    return int(match.group(1))


def pass_fail_counts(report_text: str) -> tuple[int, int]:
    passed = len(re.findall(r"\[PASS\]", report_text))
    failed = len(re.findall(r"\[FAIL\]", report_text))
    return passed, failed


def overall_runtime(runtime_text: str) -> str:
    match = re.search(r"Overall web runtime smoke: \*\*(PASS|FAIL)\*\*", runtime_text)
    if not match:
        return "UNKNOWN"
    return match.group(1)


def main() -> int:
    parser = argparse.ArgumentParser(description="Generate V1 gate dashboard")
    parser.add_argument("--repo-root", default=".")
    parser.add_argument("--out", default="docs/ops/v1-gate-dashboard-latest.md")
    args = parser.parse_args()

    root = Path(args.repo_root).resolve()
    preflight = read_text(root / "docs/ops/v1-local-preflight-latest.md")
    runtime = read_text(root / "docs/ops/v1-runtime-smoke-latest.md")
    snapshot = read_text(root / "docs/ops/v1-open-gates-snapshot.md")

    pf_pass, pf_fail = pass_fail_counts(preflight)
    rt_state = overall_runtime(runtime)
    total_open = extract_total_open(snapshot)

    now = dt.datetime.now().isoformat(timespec="seconds")
    lines = [
        "# V1 Gate Dashboard",
        "",
        f"Generated: `{now}`",
        "",
        "## Current State",
        "",
        f"- Local preflight checks: `{pf_pass}` pass / `{pf_fail}` fail",
        f"- Runtime smoke status: `{rt_state}`",
        f"- Total open checklist gates: `{total_open if total_open is not None else 'unknown'}`",
        "",
        "## Source Reports",
        "",
        "- `docs/ops/v1-local-preflight-latest.md`",
        "- `docs/ops/v1-runtime-smoke-latest.md`",
        "- `docs/ops/v1-open-gates-snapshot.md`",
        "- `docs/ops/v1-launch-blocker-evidence.md`",
        "",
        "## Next Execution",
        "",
        "1. Run deployed staging smoke and record evidence.",
        "2. Run manual web/mobile QA sheets and record P0/P1 outcomes.",
        "3. Close store/domain/secrets approvals and capture final go/no-go.",
        "",
    ]

    out_path = (root / args.out).resolve()
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text("\n".join(lines), encoding="utf-8")
    print(f"Wrote dashboard: {out_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
