#!/usr/bin/env python3
"""Build a markdown trend report from profile eval summaries."""

from __future__ import annotations

import argparse
import datetime as dt
import json
from pathlib import Path
from typing import Any, Dict, List


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Generate eval trend markdown")
    parser.add_argument("--profiles", required=True, help="Path to profiles_summary.json")
    parser.add_argument("--out", required=True, help="Output markdown file path")
    parser.add_argument("--run-id", default="", help="Optional CI run id")
    return parser.parse_args()


def load_profiles(path: Path) -> Dict[str, Any]:
    with path.open("r", encoding="utf-8") as f:
        return json.load(f)


def fmt_pct(value: Any) -> str:
    try:
        return f"{float(value):.2f}%"
    except Exception:
        return "n/a"


def fmt_ms(value: Any) -> str:
    try:
        return f"{float(value):.2f}"
    except Exception:
        return "n/a"


def render(payload: Dict[str, Any], run_id: str) -> str:
    profiles: List[Dict[str, Any]] = list(payload.get("profiles", []))
    winner = payload.get("winner")
    model = payload.get("model", "unknown")
    generated = dt.datetime.utcnow().replace(microsecond=0).isoformat() + "Z"

    lines = [
        "# Nightly LLM Eval Trend Report",
        "",
        f"- Generated at: `{generated}`",
        f"- Model: `{model}`",
        f"- Winner profile: `{winner}`",
    ]
    if run_id:
        lines.append(f"- Run id: `{run_id}`")

    lines.extend(
        [
            "",
            "## Profile Comparison",
            "",
            "| Profile | Weighted | Schema Pass | Safety Pass | Complex Pass | P95 Latency (ms) |",
            "|---|---:|---:|---:|---:|---:|",
        ]
    )

    for profile in profiles:
        lines.append(
            "| {profile} | {weighted:.2f} | {schema} | {safety} | {complexity} | {p95} |".format(
                profile=profile.get("prompt_profile", "unknown"),
                weighted=float(profile.get("weighted_score", 0.0)),
                schema=fmt_pct(profile.get("schema_pass_rate_percent")),
                safety=fmt_pct(profile.get("safety_pass_rate_percent")),
                complexity=fmt_pct(profile.get("complex_pass_rate_percent")),
                p95=fmt_ms(profile.get("p95_latency_ms")),
            )
        )

    return "\n".join(lines) + "\n"


def main() -> int:
    args = parse_args()
    payload = load_profiles(Path(args.profiles))
    markdown = render(payload, args.run_id)

    out_path = Path(args.out)
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(markdown, encoding="utf-8")
    print(f"Wrote trend report to: {out_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
