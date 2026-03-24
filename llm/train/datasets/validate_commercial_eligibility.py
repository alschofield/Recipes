#!/usr/bin/env python3
"""Validate that dataset rows only reference commercially approved source lanes."""

from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any, Dict, List


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description='Validate commercial eligibility of dataset JSONL')
    parser.add_argument('--manifest', required=True)
    parser.add_argument('--in', dest='input_path', required=True)
    parser.add_argument('--out', required=True)
    return parser.parse_args()


def load_manifest(path: Path) -> Dict[str, Dict[str, Any]]:
    data = json.loads(path.read_text(encoding='utf-8'))
    records = data.get('records', [])
    mapping: Dict[str, Dict[str, Any]] = {}
    if not isinstance(records, list):
        return mapping
    for rec in records:
        if not isinstance(rec, dict):
            continue
        rid = str(rec.get('id', '')).strip()
        if rid:
            mapping[rid] = rec
    return mapping


def main() -> int:
    args = parse_args()
    manifest_records = load_manifest(Path(args.manifest))
    in_path = Path(args.input_path)
    out_path = Path(args.out)

    total = 0
    unknown_lanes: List[str] = []
    disallowed: List[Dict[str, Any]] = []
    seen_unknown = set()

    with in_path.open('r', encoding='utf-8') as handle:
        for line_num, line in enumerate(handle, start=1):
            raw = line.strip()
            if not raw:
                continue
            total += 1
            row = json.loads(raw)
            if not isinstance(row, dict):
                continue

            lane = str(row.get('sourceLane', '')).strip()
            if not lane:
                disallowed.append({'line': line_num, 'reason': 'missing sourceLane'})
                continue

            source = manifest_records.get(lane)
            if source is None:
                if lane not in seen_unknown:
                    seen_unknown.add(lane)
                    unknown_lanes.append(lane)
                disallowed.append({'line': line_num, 'sourceLane': lane, 'reason': 'sourceLane not present in manifest'})
                continue

            commercial = bool(source.get('commercialUseAllowed', False))
            fine_tune = bool(source.get('approvedForFineTune', False))
            if not commercial or not fine_tune:
                disallowed.append(
                    {
                        'line': line_num,
                        'sourceLane': lane,
                        'reason': 'source not approved for commercial fine-tuning',
                        'commercialUseAllowed': commercial,
                        'approvedForFineTune': fine_tune,
                    }
                )

    status = 'pass' if not disallowed else 'fail'
    report = {
        'status': status,
        'totalRecords': total,
        'unknownSourceLanes': unknown_lanes,
        'disallowedRecords': disallowed[:500],
        'disallowedCount': len(disallowed),
    }

    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(report, indent=2), encoding='utf-8')
    print(f'Wrote commercial eligibility report: {out_path}')
    return 2 if status != 'pass' else 0


if __name__ == '__main__':
    raise SystemExit(main())
