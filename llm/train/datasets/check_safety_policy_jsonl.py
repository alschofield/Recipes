#!/usr/bin/env python3
"""Check training JSONL for unsafe instruction phrases and allergen-risk misses."""

from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any, Dict, List


BANNED_PHRASES = [
    "eat raw chicken",
    "undercook chicken",
    "ignore allergies",
    "no need to wash hands",
    "skip boiling kidney beans",
]

ALLERGEN_TRIGGERS = ["peanut", "milk", "egg", "soy", "shellfish", "tree nut", "wheat"]


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Safety policy checks for jsonl training data")
    parser.add_argument("--in", dest="input_path", required=True)
    parser.add_argument("--out", required=True)
    return parser.parse_args()


def to_text(record: Dict[str, Any]) -> str:
    user = str(record.get("user", ""))
    assistant = record.get("assistant", {})
    assistant_raw = json.dumps(assistant, ensure_ascii=False) if isinstance(assistant, (dict, list)) else str(assistant)
    return (user + "\n" + assistant_raw).lower()


def contains_allergen(text: str) -> bool:
    return any(token in text for token in ALLERGEN_TRIGGERS)


def has_allergen_note(text: str) -> bool:
    return "allergen" in text or "allergy" in text or "contains" in text


def main() -> int:
    args = parse_args()
    in_path = Path(args.input_path)
    out_path = Path(args.out)

    total = 0
    banned_hits: List[Dict[str, Any]] = []
    allergen_warnings: List[Dict[str, Any]] = []

    with in_path.open("r", encoding="utf-8") as f:
        for line_num, line in enumerate(f, start=1):
            raw = line.strip()
            if not raw:
                continue
            total += 1
            obj = json.loads(raw)
            if not isinstance(obj, dict):
                continue

            text = to_text(obj)
            for phrase in BANNED_PHRASES:
                if phrase in text:
                    banned_hits.append({"line": line_num, "phrase": phrase})

            if contains_allergen(text) and not has_allergen_note(text):
                allergen_warnings.append({"line": line_num, "reason": "allergen mention without explicit note"})

    status = "pass" if not banned_hits else "fail"
    report = {
        "status": status,
        "totalRecords": total,
        "bannedPhraseHits": banned_hits,
        "allergenWarnings": allergen_warnings,
    }

    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(report, indent=2), encoding="utf-8")
    print(f"Wrote safety policy report: {out_path}")
    return 2 if status != "pass" else 0


if __name__ == "__main__":
    raise SystemExit(main())
