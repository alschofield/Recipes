#!/usr/bin/env python3
"""Judge-output drift check against baseline priors.

Input snapshot may be:
- JSON object with `items` list (for example `/ingredients/catalog` payload)
- JSON object with `records` list
- raw JSON list of records
"""

from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any, Dict, Iterable, List, Tuple


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Check judge drift against baseline bands")
    parser.add_argument("--snapshot", required=True, help="Path to current judge snapshot JSON")
    parser.add_argument("--baseline", required=True, help="Path to baseline priors JSON")
    parser.add_argument("--out-json", required=True, help="Output JSON report path")
    parser.add_argument("--out-md", required=True, help="Output markdown report path")

    parser.add_argument("--low-confidence-threshold", type=float, default=0.65)
    parser.add_argument("--max-category-l1", type=float, default=0.35)
    parser.add_argument("--max-low-confidence-share", type=float, default=0.55)
    parser.add_argument("--max-confidence-p50-delta", type=float, default=0.30)
    return parser.parse_args()


def load_json(path: Path) -> Any:
    with path.open("r", encoding="utf-8") as f:
        return json.load(f)


def as_records(snapshot: Any) -> List[Dict[str, Any]]:
    if isinstance(snapshot, list):
        return [r for r in snapshot if isinstance(r, dict)]
    if isinstance(snapshot, dict):
        if isinstance(snapshot.get("items"), list):
            return [r for r in snapshot["items"] if isinstance(r, dict)]
        if isinstance(snapshot.get("records"), list):
            return [r for r in snapshot["records"] if isinstance(r, dict)]
    return []


def get_path(obj: Dict[str, Any], path: str) -> Any:
    current: Any = obj
    for key in path.split("."):
        if not isinstance(current, dict):
            return None
        current = current.get(key)
    return current


def first_non_empty(obj: Dict[str, Any], paths: Iterable[str]) -> Any:
    for path in paths:
        value = get_path(obj, path)
        if value is None:
            continue
        if isinstance(value, str) and value.strip() == "":
            continue
        return value
    return None


def normalize_category(value: Any) -> str:
    if value is None:
        return "unknown"
    raw = str(value).strip().lower()
    return raw if raw else "unknown"


def parse_confidence(value: Any) -> float | None:
    if value is None:
        return None
    try:
        v = float(value)
    except Exception:
        return None
    if v < 0:
        return 0.0
    if v > 1:
        return 1.0
    return v


def percentile(values: List[float], p: float) -> float:
    if not values:
        return 0.0
    values = sorted(values)
    if p <= 0:
        return values[0]
    if p >= 1:
        return values[-1]
    idx = int(round((len(values) - 1) * p))
    return values[idx]


def dist_from_counts(counts: Dict[str, int]) -> Dict[str, float]:
    total = float(sum(counts.values()))
    if total <= 0:
        return {}
    return {k: v / total for k, v in counts.items()}


def l1_distance(a: Dict[str, float], b: Dict[str, float]) -> float:
    keys = set(a.keys()) | set(b.keys())
    return sum(abs(a.get(k, 0.0) - b.get(k, 0.0)) for k in keys)


def baseline_category_dist(baseline: Dict[str, Any]) -> Dict[str, float]:
    categories = baseline.get("canonical_seed", {}).get("top_categories", [])
    counts: Dict[str, int] = {}
    for pair in categories:
        if not isinstance(pair, list) or len(pair) != 2:
            continue
        name = normalize_category(pair[0])
        try:
            count = int(pair[1])
        except Exception:
            continue
        counts[name] = max(0, count)
    return dist_from_counts(counts)


def build_reports(
    records: List[Dict[str, Any]],
    baseline: Dict[str, Any],
    low_conf_threshold: float,
    max_category_l1: float,
    max_low_conf_share: float,
    max_conf_p50_delta: float,
) -> Tuple[Dict[str, Any], str, bool]:
    category_counts: Dict[str, int] = {}
    confidences: List[float] = []

    for record in records:
        category = normalize_category(
            first_non_empty(record, ["category", "judgeCategory", "metadata.category", "judge.category"])
        )
        category_counts[category] = category_counts.get(category, 0) + 1

        conf_raw = first_non_empty(
            record,
            [
                "confidence",
                "judgeConfidence",
                "metadata.judge.confidence",
                "judge.confidence",
                "qualityScore",
            ],
        )
        conf = parse_confidence(conf_raw)
        if conf is not None:
            confidences.append(conf)

    current_cat_dist = dist_from_counts(category_counts)
    baseline_cat_dist = baseline_category_dist(baseline)

    category_l1 = l1_distance(current_cat_dist, baseline_cat_dist)
    low_conf_share = 0.0
    confidence_p50 = 0.0
    if confidences:
        low_conf_share = sum(1 for c in confidences if c < low_conf_threshold) / float(len(confidences))
        confidence_p50 = percentile(confidences, 0.5)

    baseline_conf_p50 = float(
        baseline.get("canonical_seed", {}).get("quality", {}).get("p50", 0.0)
    )
    confidence_p50_delta = abs(confidence_p50 - baseline_conf_p50)

    breaches = {
        "categoryL1": category_l1 > max_category_l1,
        "lowConfidenceShare": low_conf_share > max_low_conf_share,
        "confidenceP50Delta": confidence_p50_delta > max_conf_p50_delta,
    }
    failed = any(breaches.values())

    report = {
        "recordCount": len(records),
        "confidenceCount": len(confidences),
        "thresholds": {
            "lowConfidenceThreshold": low_conf_threshold,
            "maxCategoryL1": max_category_l1,
            "maxLowConfidenceShare": max_low_conf_share,
            "maxConfidenceP50Delta": max_conf_p50_delta,
        },
        "metrics": {
            "categoryL1": category_l1,
            "lowConfidenceShare": low_conf_share,
            "confidenceP50": confidence_p50,
            "baselineConfidenceP50": baseline_conf_p50,
            "confidenceP50Delta": confidence_p50_delta,
        },
        "breaches": breaches,
        "status": "fail" if failed else "pass",
        "topCategories": sorted(category_counts.items(), key=lambda x: x[1], reverse=True)[:15],
    }

    lines = [
        "# Judge Drift Report",
        "",
        f"- Status: `{report['status']}`",
        f"- Records: `{len(records)}`",
        f"- Confidence values: `{len(confidences)}`",
        "",
        "## Metrics",
        "",
        f"- Category L1 distance: `{category_l1:.4f}` (max `{max_category_l1:.4f}`)",
        f"- Low-confidence share: `{low_conf_share:.4f}` (max `{max_low_conf_share:.4f}`; threshold `< {low_conf_threshold:.2f}`)",
        f"- Confidence p50 delta vs baseline: `{confidence_p50_delta:.4f}` (max `{max_conf_p50_delta:.4f}`)",
        "",
        "## Breaches",
        "",
        f"- categoryL1: `{'breach' if breaches['categoryL1'] else 'ok'}`",
        f"- lowConfidenceShare: `{'breach' if breaches['lowConfidenceShare'] else 'ok'}`",
        f"- confidenceP50Delta: `{'breach' if breaches['confidenceP50Delta'] else 'ok'}`",
    ]

    markdown = "\n".join(lines) + "\n"
    return report, markdown, failed


def main() -> int:
    args = parse_args()
    records = as_records(load_json(Path(args.snapshot)))
    baseline = load_json(Path(args.baseline))

    report, markdown, failed = build_reports(
        records=records,
        baseline=baseline,
        low_conf_threshold=args.low_confidence_threshold,
        max_category_l1=args.max_category_l1,
        max_low_conf_share=args.max_low_confidence_share,
        max_conf_p50_delta=args.max_confidence_p50_delta,
    )

    out_json = Path(args.out_json)
    out_md = Path(args.out_md)
    out_json.parent.mkdir(parents=True, exist_ok=True)
    out_md.parent.mkdir(parents=True, exist_ok=True)

    out_json.write_text(json.dumps(report, indent=2), encoding="utf-8")
    out_md.write_text(markdown, encoding="utf-8")

    print(f"Wrote drift JSON: {out_json}")
    print(f"Wrote drift markdown: {out_md}")

    return 2 if failed else 0


if __name__ == "__main__":
    raise SystemExit(main())
