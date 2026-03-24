#!/usr/bin/env python3
"""Deduplicate near-identical recipe JSONL rows by normalized signature."""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path
from typing import Any, Dict, Iterable, Tuple


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Deduplicate recipe jsonl")
    parser.add_argument("--in", dest="input_path", required=True)
    parser.add_argument("--out", required=True)
    parser.add_argument("--report", required=True)
    return parser.parse_args()


def normalize_tokens(values: Iterable[str]) -> str:
    tokens = []
    for value in values:
        low = str(value).strip().lower()
        if low:
            tokens.append(low)
    tokens.sort()
    return "|".join(tokens)


def signature(record: Dict[str, Any]) -> Tuple[str, str]:
    prompt = str(record.get("user", "")).strip().lower()
    assistant = record.get("assistant", {})
    if not isinstance(assistant, dict):
        assistant = {}

    recipes = assistant.get("recipes", [])
    names = []
    ingredients = []
    if isinstance(recipes, list):
        for recipe in recipes:
            if not isinstance(recipe, dict):
                continue
            names.append(str(recipe.get("name", "")))
            rec_ingredients = recipe.get("ingredients", [])
            if isinstance(rec_ingredients, list):
                for ing in rec_ingredients:
                    if isinstance(ing, dict):
                        ingredients.append(str(ing.get("name", "")))

    material = "\n".join([prompt, normalize_tokens(names), normalize_tokens(ingredients)])
    return hashlib.sha256(material.encode("utf-8")).hexdigest(), material


def main() -> int:
    args = parse_args()
    in_path = Path(args.input_path)
    out_path = Path(args.out)
    report_path = Path(args.report)

    seen = set()
    total = 0
    kept = 0
    dropped = 0

    out_path.parent.mkdir(parents=True, exist_ok=True)
    with in_path.open("r", encoding="utf-8") as src, out_path.open("w", encoding="utf-8") as dst:
        for line in src:
            raw = line.strip()
            if not raw:
                continue
            total += 1
            obj = json.loads(raw)
            if not isinstance(obj, dict):
                dropped += 1
                continue

            sig, _ = signature(obj)
            if sig in seen:
                dropped += 1
                continue

            seen.add(sig)
            dst.write(json.dumps(obj, ensure_ascii=False) + "\n")
            kept += 1

    report = {
        "input": str(in_path),
        "output": str(out_path),
        "total": total,
        "kept": kept,
        "dropped": dropped,
        "dedupRatePercent": (dropped / total * 100.0) if total else 0.0,
    }
    report_path.parent.mkdir(parents=True, exist_ok=True)
    report_path.write_text(json.dumps(report, indent=2), encoding="utf-8")
    print(f"Wrote deduped dataset: {out_path}")
    print(f"Wrote dedup report: {report_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
