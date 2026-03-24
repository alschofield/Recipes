#!/usr/bin/env python3
"""Generate a consolidated snapshot of open checklist gates."""

from __future__ import annotations

import argparse
import datetime as dt
from pathlib import Path


CHECKLISTS = [
    ("Root", "CHECKLIST.md"),
    ("Server", "server/CHECKLIST.md"),
    ("Web", "web/CHECKLIST.md"),
    ("Mobile", "mobile/CHECKLIST.md"),
    ("LLM", "llm/CHECKLIST.md"),
]


def extract_open_items(path: Path) -> list[str]:
    items: list[str] = []
    for raw_line in path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if line.startswith("- [ ] "):
            items.append(line[6:].strip())
    return items


def build_snapshot(repo_root: Path) -> str:
    timestamp = dt.datetime.now().isoformat(timespec="seconds")
    out: list[str] = [
        "# V1 Open Gates Snapshot",
        "",
        f"Generated: `{timestamp}`",
        "",
        "This file is generated from checklist sources. Do not edit by hand.",
        "",
    ]

    total = 0
    for label, rel_path in CHECKLISTS:
        path = repo_root / rel_path
        if not path.exists():
            continue
        items = extract_open_items(path)
        total += len(items)
        out.extend([f"## {label} ({len(items)})", ""])
        if not items:
            out.append("- No open gates")
            out.append("")
            continue
        for item in items:
            out.append(f"- {item}")
        out.append("")

    out.extend([f"Total open gates: **{total}**", ""])
    return "\n".join(out)


def main() -> int:
    parser = argparse.ArgumentParser(description="Generate V1 open gates snapshot")
    parser.add_argument("--repo-root", default=".", help="Repository root path")
    parser.add_argument(
        "--out",
        default="docs/ops/v1-open-gates-snapshot.md",
        help="Output markdown path relative to repo root",
    )
    args = parser.parse_args()

    repo_root = Path(args.repo_root).resolve()
    out_path = (repo_root / args.out).resolve()
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(build_snapshot(repo_root), encoding="utf-8")
    print(f"Wrote snapshot: {out_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
