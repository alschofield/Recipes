#!/usr/bin/env python3
"""Run V1 local preflight checks and emit a markdown report.

This script is intentionally lightweight and cross-platform friendly.
"""

from __future__ import annotations

import argparse
import datetime as dt
import os
import subprocess
from dataclasses import dataclass
from pathlib import Path


@dataclass
class CheckResult:
    label: str
    command: str
    cwd: str
    ok: bool
    output: str


def run_check(label: str, command: str, cwd: Path, env: dict[str, str] | None = None) -> CheckResult:
    completed = subprocess.run(
        command,
        cwd=str(cwd),
        shell=True,
        capture_output=True,
        text=True,
        env=env,
    )
    output = (completed.stdout or "") + (completed.stderr or "")
    return CheckResult(
        label=label,
        command=command,
        cwd=str(cwd),
        ok=completed.returncode == 0,
        output=output.strip(),
    )


def build_report(results: list[CheckResult]) -> str:
    now = dt.datetime.now().isoformat(timespec="seconds")
    lines = [
        "# V1 Local Preflight Report",
        "",
        f"Generated: `{now}`",
        "",
        "## Summary",
        "",
    ]

    for result in results:
        status = "PASS" if result.ok else "FAIL"
        lines.append(f"- [{status}] {result.label}")

    lines.extend(["", "## Details", ""])

    for result in results:
        status = "PASS" if result.ok else "FAIL"
        lines.extend(
            [
                f"### {result.label} ({status})",
                "",
                f"- CWD: `{result.cwd}`",
                f"- Command: `{result.command}`",
                "",
                "```text",
                result.output[:12000] if result.output else "(no output)",
                "```",
                "",
            ]
        )

    return "\n".join(lines)


def main() -> int:
    parser = argparse.ArgumentParser(description="Run local V1 preflight checks")
    parser.add_argument(
        "--repo-root",
        default=".",
        help="Path to repository root (default: current directory)",
    )
    parser.add_argument(
        "--out",
        default="docs/ops/v1-local-preflight-latest.md",
        help="Output markdown report path (relative to repo root)",
    )
    parser.add_argument(
        "--skip-android",
        action="store_true",
        help="Skip Android assemble check",
    )
    parser.add_argument(
        "--android-java-home",
        default="",
        help="Optional JAVA_HOME override for Android build",
    )
    args = parser.parse_args()

    repo_root = Path(args.repo_root).resolve()
    server_dir = repo_root / "server"
    web_dir = repo_root / "web"
    android_dir = repo_root / "mobile" / "android-native" / "RecipesMobile"

    checks: list[tuple[str, str, Path, dict[str, str] | None]] = [
        ("Server tests", "go test ./...", server_dir, None),
        ("Server build", "go build ./...", server_dir, None),
        ("Web lint", "npm run lint", web_dir, None),
        ("Web unit tests", "npm run test", web_dir, None),
        ("Web build", "npm run build", web_dir, None),
        ("Web e2e", "npm run test:e2e", web_dir, None),
    ]

    if not args.skip_android:
        android_env = os.environ.copy()
        if args.android_java_home:
            android_env["JAVA_HOME"] = args.android_java_home
            android_env["PATH"] = str(Path(args.android_java_home) / "bin") + os.pathsep + android_env.get("PATH", "")
        android_command = "gradlew.bat :app:assembleDevDebug" if os.name == "nt" else "./gradlew :app:assembleDevDebug"
        checks.append(("Android assembleDevDebug", android_command, android_dir, android_env))

    results = [run_check(label, command, cwd, env) for label, command, cwd, env in checks]

    out_path = (repo_root / args.out).resolve()
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(build_report(results), encoding="utf-8")

    any_fail = any(not result.ok for result in results)
    print(f"Wrote report: {out_path}")
    return 1 if any_fail else 0


if __name__ == "__main__":
    raise SystemExit(main())
