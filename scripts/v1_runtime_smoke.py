#!/usr/bin/env python3
"""Run local runtime smoke checks for web and optional Android emulator.

Outputs a markdown report under docs/ops by default.
"""

from __future__ import annotations

import argparse
import datetime as dt
import json
import os
import subprocess
import time
import urllib.error
import urllib.request
from pathlib import Path


def http_status(url: str, timeout: float = 5.0) -> tuple[bool, int | None, str]:
    try:
        req = urllib.request.Request(url, method="GET")
        with urllib.request.urlopen(req, timeout=timeout) as response:
            body = response.read().decode("utf-8", errors="replace")
            return True, response.status, body
    except urllib.error.HTTPError as exc:
        return False, exc.code, str(exc)
    except Exception as exc:  # noqa: BLE001
        return False, None, str(exc)


def run_adb(adb_path: str, args: list[str]) -> tuple[bool, str]:
    try:
        completed = subprocess.run(
            [adb_path, *args],
            capture_output=True,
            text=True,
            timeout=90,
            check=False,
        )
    except Exception as exc:  # noqa: BLE001
        return False, str(exc)

    output = (completed.stdout or "") + (completed.stderr or "")
    return completed.returncode == 0, output.strip()


def kill_port_listener(port: int) -> None:
    if os.name != "nt":
        return
    cmd = f'for /f "tokens=5" %a in (\'netstat -ano ^| findstr :{port} ^| findstr LISTENING\') do taskkill /PID %a /F'
    subprocess.run(["cmd.exe", "/c", cmd], check=False, capture_output=True, text=True)


def main() -> int:
    parser = argparse.ArgumentParser(description="Run V1 runtime smoke checks")
    parser.add_argument("--repo-root", default=".")
    parser.add_argument("--port", type=int, default=3010)
    parser.add_argument("--with-android", action="store_true")
    parser.add_argument(
        "--adb-path",
        default="C:/Users/alexs/AppData/Local/Android/Sdk/platform-tools/adb.exe",
    )
    parser.add_argument(
        "--out",
        default="docs/ops/v1-runtime-smoke-latest.md",
        help="Output markdown path relative to repo root",
    )
    args = parser.parse_args()

    repo_root = Path(args.repo_root).resolve()
    web_dir = repo_root / "web"
    now = dt.datetime.now().isoformat(timespec="seconds")

    if os.name == "nt":
        start_cmd = f"cmd.exe /c npm run start -- -p {args.port}"
    else:
        start_cmd = f"npm run start -- -p {args.port}"

    server = subprocess.Popen(  # noqa: S603
        start_cmd,
        cwd=str(web_dir),
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
        shell=True,
    )

    base = f"http://127.0.0.1:{args.port}"
    routes = ["/", "/recipes", "/login", "/signup", "/api/health"]
    web_results: list[tuple[str, bool, int | None, str]] = []

    try:
        # Give Next a moment to boot.
        time.sleep(8)
        for route in routes:
            ok, status, body = http_status(base + route)
            web_results.append((route, ok, status, body[:500]))
    finally:
        server.terminate()
        try:
            server.wait(timeout=20)
        except Exception:  # noqa: BLE001
            server.kill()
        kill_port_listener(args.port)

    android_lines: list[str] = []
    if args.with_android:
        adb = args.adb_path
        ok_devices, out_devices = run_adb(adb, ["devices"])
        android_lines.append(f"- adb devices: {'PASS' if ok_devices else 'FAIL'}")
        android_lines.append("```text")
        android_lines.append(out_devices or "(no output)")
        android_lines.append("```")

        ok_boot, out_boot = run_adb(adb, ["shell", "getprop", "sys.boot_completed"])
        boot_ok = ok_boot and out_boot.strip().endswith("1")
        android_lines.append(f"- emulator boot completed: {'PASS' if boot_ok else 'FAIL'}")
        android_lines.append("```text")
        android_lines.append(out_boot or "(no output)")
        android_lines.append("```")

        ok_pid, out_pid = run_adb(adb, ["shell", "pidof", "com.recipes.mobile.dev"])
        pid_ok = ok_pid and bool(out_pid.strip())
        android_lines.append(f"- app process check (com.recipes.mobile.dev): {'PASS' if pid_ok else 'FAIL'}")
        android_lines.append("```text")
        android_lines.append(out_pid or "(no output)")
        android_lines.append("```")

    lines = [
        "# V1 Runtime Smoke Report",
        "",
        f"Generated: `{now}`",
        "",
        "## Web Routes",
        "",
    ]
    all_web_ok = True
    for route, ok, status, body in web_results:
        good = ok and status == 200
        all_web_ok = all_web_ok and good
        lines.append(f"- {route}: {'PASS' if good else 'FAIL'} (status={status})")
        if not good:
            lines.append("```text")
            lines.append(body or "(no output)")
            lines.append("```")

    if args.with_android:
        lines.extend(["", "## Android Runtime", "", *android_lines])

    lines.extend(["", f"Overall web runtime smoke: **{'PASS' if all_web_ok else 'FAIL'}**", ""])

    out_path = (repo_root / args.out).resolve()
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text("\n".join(lines), encoding="utf-8")
    print(f"Wrote runtime smoke report: {out_path}")

    if not all_web_ok:
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
