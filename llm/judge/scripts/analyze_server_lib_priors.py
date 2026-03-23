#!/usr/bin/env python3

import csv
import json
import statistics
import zipfile
from collections import Counter, defaultdict
from datetime import UTC, datetime
from pathlib import Path


ROOT = Path(__file__).resolve().parents[3]
DATASETS_RAW = ROOT / "datasets" / "raw" / "server-lib"
DATASETS_DERIVED = ROOT / "datasets" / "derived" / "server-lib"


def percentile(sorted_vals, p):
    if not sorted_vals:
        return 0
    idx = max(0, min(len(sorted_vals) - 1, int(p * len(sorted_vals)) - 1))
    return sorted_vals[idx]


def canonical_seed_stats():
    path = DATASETS_DERIVED / "canonical_ingredient_seed_v1.csv"
    rows = list(csv.DictReader(path.open(encoding="utf-8")))

    quality = []
    coverage_buckets = defaultdict(list)
    molecule_buckets = defaultdict(list)
    status = Counter()
    category = Counter()

    for row in rows:
        status[row.get("analysis_status", "").strip() or "unknown"] += 1
        category[row.get("category", "").strip().lower() or "unknown"] += 1

        try:
            q = float(row.get("quality_score") or 0)
            c = int(float(row.get("source_coverage") or 0))
            m = int(float(row.get("flavour_molecule_count") or 0))
        except ValueError:
            continue

        quality.append(q)
        coverage_buckets[c].append(q)
        molecule_buckets[c].append(m)

    quality_sorted = sorted(quality)

    by_coverage = {}
    for cov in sorted(coverage_buckets):
        vals = sorted(coverage_buckets[cov])
        mols = sorted(molecule_buckets[cov])
        by_coverage[str(cov)] = {
            "count": len(vals),
            "quality_mean": round(sum(vals) / len(vals), 4),
            "quality_p90": round(percentile(vals, 0.9), 4),
            "molecule_p50": percentile(mols, 0.5),
        }

    return {
        "rows": len(rows),
        "analysis_status_counts": dict(status),
        "quality": {
            "min": round(min(quality_sorted), 4) if quality_sorted else 0,
            "p50": round(percentile(quality_sorted, 0.5), 4),
            "p90": round(percentile(quality_sorted, 0.9), 4),
            "max": round(max(quality_sorted), 4) if quality_sorted else 0,
        },
        "top_categories": category.most_common(20),
        "quality_by_source_coverage": by_coverage,
    }


def kaggle_recipe_stats():
    path = DATASETS_RAW / "kaggle ingredients dataset.zip"
    with zipfile.ZipFile(path) as zf:
        name = next(n for n in zf.namelist() if n.endswith("train.json"))
        data = json.loads(zf.read(name))

    cuisine = Counter(item.get("cuisine", "unknown") for item in data)
    ingredient_counts = sorted(len(item.get("ingredients", [])) for item in data)

    return {
        "recipes": len(data),
        "top_cuisines": cuisine.most_common(15),
        "ingredients_per_recipe": {
            "p50": percentile(ingredient_counts, 0.5),
            "p90": percentile(ingredient_counts, 0.9),
            "max": max(ingredient_counts) if ingredient_counts else 0,
            "mean": round(statistics.mean(ingredient_counts), 4) if ingredient_counts else 0,
        },
    }


def main():
    out = {
        "generated_at": datetime.now(UTC).isoformat(),
        "sources": {
            "canonical_seed": "datasets/derived/server-lib/canonical_ingredient_seed_v1.csv",
            "kaggle_train": "datasets/raw/server-lib/kaggle ingredients dataset.zip::train.json",
        },
        "canonical_seed": canonical_seed_stats(),
        "kaggle": kaggle_recipe_stats(),
    }

    out_path = ROOT / "llm" / "judge" / "data-priors.summary.json"
    out_path.write_text(json.dumps(out, indent=2), encoding="utf-8")
    print(f"Wrote {out_path}")


if __name__ == "__main__":
    main()
