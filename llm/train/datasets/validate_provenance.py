#!/usr/bin/env python3
"""Validate provenance manifest records and denylist constraints."""

from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any, Dict, List


REQUIRED_FIELDS = [
    "id",
    "lane",
    "sourceName",
    "sourceUrl",
    "originType",
    "retrievedDateUtc",
    "licenseType",
    "commercialUseAllowed",
    "derivativeWorksAllowed",
    "attributionRequired",
    "shareAlikeRequired",
    "legalReviewReference",
    "approvedForEval",
    "approvedForFineTune",
    "approvedForProductionInference",
    "approvedBy",
]


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Validate dataset provenance manifest")
    parser.add_argument("--manifest", required=True)
    parser.add_argument("--denylist", required=True)
    parser.add_argument("--out", required=True)
    return parser.parse_args()


def load_json(path: Path) -> Dict[str, Any]:
    with path.open("r", encoding="utf-8") as f:
        data = json.load(f)
    if not isinstance(data, dict):
        raise ValueError(f"manifest must be object: {path}")
    return data


def load_denylist(path: Path) -> List[str]:
    lines: List[str] = []
    with path.open("r", encoding="utf-8") as f:
        for raw in f:
            line = raw.strip().lower()
            if not line or line.startswith("#"):
                continue
            lines.append(line)
    return lines


def contains_denylisted(url: str, denylist: List[str]) -> str:
    low = url.lower()
    for entry in denylist:
        if entry in low:
            return entry
    return ""


def validate(manifest: Dict[str, Any], denylist: List[str]) -> Dict[str, Any]:
    records = manifest.get("records", [])
    errors: List[str] = []
    warnings: List[str] = []
    blocked: List[str] = []

    if not isinstance(records, list) or not records:
        errors.append("manifest.records must be a non-empty list")
        return {
            "status": "fail",
            "errors": errors,
            "warnings": warnings,
            "blockedRecords": blocked,
            "recordCount": 0,
        }

    seen_ids = set()
    for idx, record in enumerate(records):
        if not isinstance(record, dict):
            errors.append(f"records[{idx}] must be an object")
            continue

        rid = str(record.get("id", "")).strip()
        if rid == "":
            errors.append(f"records[{idx}] missing id")
        elif rid in seen_ids:
            errors.append(f"duplicate record id: {rid}")
        else:
            seen_ids.add(rid)

        for key in REQUIRED_FIELDS:
            if key not in record:
                errors.append(f"records[{idx}] missing field: {key}")

        src = str(record.get("sourceUrl", "")).strip()
        if src == "":
            errors.append(f"records[{idx}] sourceUrl is empty")
        else:
            hit = contains_denylisted(src, denylist)
            if hit:
                errors.append(f"records[{idx}] denylisted source matched '{hit}'")
                if rid:
                    blocked.append(rid)

        license_type = str(record.get("licenseType", "")).strip().lower()
        if license_type in {"", "unknown", "review-required"}:
            warnings.append(f"records[{idx}] license requires explicit legal decision ({license_type or 'empty'})")

        fine_tune = bool(record.get("approvedForFineTune", False))
        commercial = bool(record.get("commercialUseAllowed", False))
        if fine_tune and not commercial:
            errors.append(f"records[{idx}] fine-tune approved but commercialUseAllowed is false")

    status = "fail" if errors else "pass"
    return {
        "status": status,
        "recordCount": len(records),
        "errors": errors,
        "warnings": warnings,
        "blockedRecords": blocked,
    }


def main() -> int:
    args = parse_args()
    manifest = load_json(Path(args.manifest))
    denylist = load_denylist(Path(args.denylist))
    result = validate(manifest, denylist)

    out_path = Path(args.out)
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(result, indent=2), encoding="utf-8")
    print(f"Wrote provenance validation report: {out_path}")

    if result["warnings"]:
        for warning in result["warnings"]:
            print(f"warning: {warning}")

    return 2 if result["status"] != "pass" else 0


if __name__ == "__main__":
    raise SystemExit(main())
