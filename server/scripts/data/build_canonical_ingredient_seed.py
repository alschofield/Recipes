#!/usr/bin/env python3

import csv
import json
import os
import zipfile


ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
DERIVED = os.path.join(ROOT, "lib", "derived")
INPUT_CSV = os.path.join(DERIVED, "unified_ingredient_index.csv")
OUTPUT_CSV = os.path.join(DERIVED, "canonical_ingredient_seed_v1.csv")
OUTPUT_JSON = os.path.join(DERIVED, "canonical_ingredient_seed_v1.summary.json")
OUTPUT_ZIP = os.path.join(DERIVED, "canonical_ingredient_seed_v1.zip")


def compute_quality(row):
    coverage = int(row["source_coverage"])
    usda_hits = int(row["usda_exact_name_matches"])
    kaggle_count = int(row["kaggle_recipe_count"])
    culinary_count = int(row["culinary_alias_count"])
    molecules = int(row["flavour_molecule_count"] or 0)

    score = 0.0
    score += min(coverage / 5.0, 1.0) * 0.45
    score += min(usda_hits / 50.0, 1.0) * 0.20
    score += min(kaggle_count / 200.0, 1.0) * 0.15
    score += min(culinary_count / 2000.0, 1.0) * 0.10
    score += min(molecules / 120.0, 1.0) * 0.10

    return round(min(score, 1.0), 3)


def resolve_category(row):
    flavour = row["flavourdb_category"].strip()
    fallback = row["food_info_category"].strip()
    if flavour:
        return flavour.lower()
    if fallback:
        return fallback.lower()
    return "unknown"


def resolve_status(row):
    if int(row["source_coverage"]) >= 3:
        return "enriched"
    if int(row["source_coverage"]) == 2:
        return "pending"
    return "review_required"


def main():
    os.makedirs(DERIVED, exist_ok=True)

    with open(INPUT_CSV, "r", encoding="utf-8", newline="") as handle:
        reader = csv.DictReader(handle)
        selected = []
        for row in reader:
            source_coverage = (
                int(row["in_flavourdb"])
                + int(row["in_food_info"])
                + int(row["in_culinarydb"])
                + int(row["in_kaggle_train"])
                + int(row["in_world_recipes"])
            )

            if source_coverage < 2:
                continue

            canonical_name = row["ingredient"].strip()
            if not canonical_name:
                continue

            normalized = {
                "canonical_name": canonical_name,
                "category": resolve_category(row),
                "natural_source": row["food_info_natural_source"].strip().lower(),
                "flavour_molecule_count": row["flavour_molecule_count"].strip() or "",
                "source_coverage": str(source_coverage),
                "quality_score": "",
                "analysis_status": "",
                "analysis_notes": "auto-seeded from unified ingredient index",
                "metadata": "",
            }

            quality = compute_quality(
                {
                    **row,
                    "source_coverage": source_coverage,
                }
            )
            normalized["quality_score"] = f"{quality:.3f}"
            normalized["analysis_status"] = resolve_status({"source_coverage": source_coverage})
            normalized["metadata"] = json.dumps(
                {
                    "sources": {
                        "flavourdb": int(row["in_flavourdb"]),
                        "food_info": int(row["in_food_info"]),
                        "culinarydb": int(row["in_culinarydb"]),
                        "kaggle_train": int(row["in_kaggle_train"]),
                        "world_recipes": int(row["in_world_recipes"]),
                    },
                    "signals": {
                        "culinary_alias_count": int(row["culinary_alias_count"]),
                        "kaggle_recipe_count": int(row["kaggle_recipe_count"]),
                        "world_recipe_count": int(row["world_recipe_count"]),
                        "usda_exact_name_matches": int(row["usda_exact_name_matches"]),
                    },
                    "seed_version": "v1",
                },
                separators=(",", ":"),
            )
            selected.append(normalized)

    selected.sort(key=lambda item: (-float(item["quality_score"]), item["canonical_name"]))

    columns = [
        "canonical_name",
        "category",
        "natural_source",
        "flavour_molecule_count",
        "source_coverage",
        "quality_score",
        "analysis_status",
        "analysis_notes",
        "metadata",
    ]

    with open(OUTPUT_CSV, "w", encoding="utf-8", newline="") as handle:
        writer = csv.DictWriter(handle, fieldnames=columns)
        writer.writeheader()
        writer.writerows(selected)

    summary = {
        "input": os.path.relpath(INPUT_CSV, ROOT),
        "output": os.path.relpath(OUTPUT_CSV, ROOT),
        "rows": len(selected),
        "quality": {
            "min": min(float(item["quality_score"]) for item in selected) if selected else 0,
            "max": max(float(item["quality_score"]) for item in selected) if selected else 0,
        },
        "analysis_status_counts": {
            "enriched": sum(1 for item in selected if item["analysis_status"] == "enriched"),
            "pending": sum(1 for item in selected if item["analysis_status"] == "pending"),
            "review_required": sum(1 for item in selected if item["analysis_status"] == "review_required"),
        },
    }

    with open(OUTPUT_JSON, "w", encoding="utf-8") as handle:
        json.dump(summary, handle, indent=2)

    with zipfile.ZipFile(OUTPUT_ZIP, "w", compression=zipfile.ZIP_DEFLATED, compresslevel=9) as archive:
        archive.write(OUTPUT_CSV, arcname=os.path.basename(OUTPUT_CSV))
        archive.write(OUTPUT_JSON, arcname=os.path.basename(OUTPUT_JSON))

    print(f"Wrote {OUTPUT_CSV}")
    print(f"Wrote {OUTPUT_JSON}")
    print(f"Wrote {OUTPUT_ZIP}")
    print(f"Rows: {len(selected)}")


if __name__ == "__main__":
    main()
