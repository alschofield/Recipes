#!/usr/bin/env python3
"""Build a large, deduped recipe-query corpus from raw dataset lanes."""

from __future__ import annotations

import argparse
import ast
import csv
import datetime as dt
import hashlib
import json
import re
import zipfile
from collections import Counter
from io import StringIO
from pathlib import Path
from typing import Any, Dict, Iterable, List, Tuple


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description='Curate recipe-query corpus from raw lanes')
    parser.add_argument(
        '--raw-root',
        default='datasets/raw/server-lib',
        help='Raw dataset root directory',
    )
    parser.add_argument(
        '--out-jsonl',
        default='llm/train/datasets/raw/recipe-query-corpus.v1.jsonl',
        help='Output JSONL path for curated corpus',
    )
    parser.add_argument(
        '--out-report',
        default='llm/train/datasets/reports/recipe-query-corpus.v1.summary.json',
        help='Output summary report path',
    )
    parser.add_argument('--min-ingredients', type=int, default=3)
    parser.add_argument('--max-ingredients', type=int, default=30)
    parser.add_argument('--max-per-source', type=int, default=0, help='0 means unlimited')
    parser.add_argument('--train-ratio', type=float, default=0.7)
    parser.add_argument('--val-ratio', type=float, default=0.15)
    parser.add_argument('--test-ratio', type=float, default=0.15)
    return parser.parse_args()


def normalize_text(value: str) -> str:
    lowered = value.strip().lower()
    cleaned = re.sub(r'[^a-z0-9\-\s\'\+]', ' ', lowered)
    cleaned = re.sub(r'\s+', ' ', cleaned).strip()
    return cleaned


def normalize_ingredients(values: Iterable[str]) -> List[str]:
    seen = set()
    cleaned: List[str] = []
    for value in values:
        norm = normalize_text(str(value))
        if not norm:
            continue
        if norm in seen:
            continue
        seen.add(norm)
        cleaned.append(norm)
    return cleaned


def row_signature(cuisine: str, ingredients: List[str]) -> str:
    parts = [normalize_text(cuisine), '|'.join(sorted(ingredients))]
    material = '\n'.join(parts)
    return hashlib.sha256(material.encode('utf-8')).hexdigest()


def split_for_key(key: str, train_ratio: float, val_ratio: float) -> str:
    bucket = int(hashlib.sha256(key.encode('utf-8')).hexdigest()[:8], 16) / 0xFFFFFFFF
    if bucket < train_ratio:
        return 'train'
    if bucket < train_ratio + val_ratio:
        return 'validation'
    return 'test'


def build_query_text(cuisine: str, ingredients: List[str]) -> str:
    joined = ', '.join(ingredients)
    if cuisine:
        return f'Give me recipe ideas for {cuisine} cuisine using: {joined}.'
    return f'Give me recipe ideas using: {joined}.'


def load_json_from_zip(path: Path, inner_name: str) -> Any:
    with zipfile.ZipFile(path) as archive:
        with archive.open(inner_name) as handle:
            return json.load(handle)


def load_csv_from_zip(path: Path, inner_name: str) -> List[Dict[str, str]]:
    with zipfile.ZipFile(path) as archive:
        raw = archive.read(inner_name).decode('utf-8', errors='replace')
    reader = csv.DictReader(StringIO(raw))
    return [dict(row) for row in reader]


def parse_world_ingredients(value: str) -> List[str]:
    text = value.strip()
    if not text:
        return []
    try:
        parsed = ast.literal_eval(text)
        if isinstance(parsed, list):
            return [str(item) for item in parsed]
    except Exception:
        pass
    return [token.strip() for token in text.split(',') if token.strip()]


def iterate_kaggle_rows(raw_root: Path) -> Iterable[Dict[str, Any]]:
    path = raw_root / 'kaggle ingredients dataset.zip'
    records = load_json_from_zip(path, 'train.json')
    if not isinstance(records, list):
        return

    for row in records:
        if not isinstance(row, dict):
            continue
        yield {
            'sourceLane': 'kaggle-train-json',
            'sourcePath': 'datasets/raw/server-lib/kaggle ingredients dataset.zip::train.json',
            'sourceRecordId': str(row.get('id', '')),
            'cuisine': str(row.get('cuisine', '')).strip(),
            'ingredients': row.get('ingredients', []),
            'title': '',
        }


def iterate_world_rows(raw_root: Path) -> Iterable[Dict[str, Any]]:
    path = raw_root / 'collection of recipes dataset from kaggle.zip'
    rows = load_csv_from_zip(path, 'Receipes from around the world.csv')
    for idx, row in enumerate(rows, start=1):
        yield {
            'sourceLane': 'world-recipes-csv',
            'sourcePath': 'datasets/raw/server-lib/collection of recipes dataset from kaggle.zip::Receipes from around the world.csv',
            'sourceRecordId': str(idx),
            'cuisine': str(row.get('cuisine', '')).strip(),
            'ingredients': parse_world_ingredients(str(row.get('ingredients', ''))),
            'title': str(row.get('recipe_name', '')).strip(),
        }


def iterate_culinary_rows(raw_root: Path) -> Iterable[Dict[str, Any]]:
    path = raw_root / 'CulinaryDB.zip'
    recipe_rows = load_csv_from_zip(path, '01_Recipe_Details.csv')
    alias_rows = load_csv_from_zip(path, '04_Recipe-Ingredients_Aliases.csv')

    recipe_by_id: Dict[str, Dict[str, str]] = {}
    for row in recipe_rows:
        rid = str(row.get('Recipe ID', '')).strip()
        if rid:
            recipe_by_id[rid] = row

    ingredients_by_id: Dict[str, List[str]] = {}
    for row in alias_rows:
        rid = str(row.get('Recipe ID', '')).strip()
        if not rid:
            continue
        alias = str(row.get('Aliased Ingredient Name', '')).strip()
        original = str(row.get('Original Ingredient Name', '')).strip()
        candidate = alias or original
        if not candidate:
            continue
        ingredients_by_id.setdefault(rid, []).append(candidate)

    for rid, ingredients in ingredients_by_id.items():
        detail = recipe_by_id.get(rid, {})
        yield {
            'sourceLane': 'culinarydb-aliases',
            'sourcePath': 'datasets/raw/server-lib/CulinaryDB.zip::01_Recipe_Details.csv+04_Recipe-Ingredients_Aliases.csv',
            'sourceRecordId': rid,
            'cuisine': str(detail.get('Cuisine', '')).strip(),
            'ingredients': ingredients,
            'title': str(detail.get('Title', '')).strip(),
        }


def build_curated_record(
    source_rank: int,
    row: Dict[str, Any],
    train_ratio: float,
    val_ratio: float,
) -> Dict[str, Any]:
    cuisine = normalize_text(str(row.get('cuisine', '')))
    ingredients = normalize_ingredients(row.get('ingredients', []))
    signature = row_signature(cuisine, ingredients)
    split = split_for_key(signature, train_ratio=train_ratio, val_ratio=val_ratio)

    source_lane = str(row['sourceLane'])
    source_record_id = str(row['sourceRecordId'])
    stable_id = hashlib.sha256(f'{source_lane}:{source_record_id}'.encode('utf-8')).hexdigest()[:16]

    return {
        'id': f'query-v1-{source_rank}-{stable_id}',
        'lane': 'recipe-query',
        'sourceLane': source_lane,
        'sourcePath': str(row['sourcePath']),
        'sourceRecordId': source_record_id,
        'title': str(row.get('title', '')).strip(),
        'cuisine': cuisine,
        'ingredients': ingredients,
        'ingredientCount': len(ingredients),
        'queryText': build_query_text(cuisine, ingredients),
        'split': split,
        'signature': signature,
    }


def main() -> int:
    args = parse_args()
    ratio_sum = args.train_ratio + args.val_ratio + args.test_ratio
    if abs(ratio_sum - 1.0) > 1e-6:
        raise ValueError('train/val/test ratios must sum to 1.0')
    if args.min_ingredients < 1:
        raise ValueError('--min-ingredients must be >= 1')
    if args.max_ingredients < args.min_ingredients:
        raise ValueError('--max-ingredients must be >= --min-ingredients')

    raw_root = Path(args.raw_root)
    out_jsonl = Path(args.out_jsonl)
    out_report = Path(args.out_report)

    source_iterators = [
        iterate_kaggle_rows(raw_root),
        iterate_culinary_rows(raw_root),
        iterate_world_rows(raw_root),
    ]

    seen_signatures = set()
    split_counts: Counter[str] = Counter()
    source_counts_in: Counter[str] = Counter()
    source_counts_kept: Counter[str] = Counter()
    dropped: Counter[str] = Counter()
    cuisine_counts: Counter[str] = Counter()

    out_jsonl.parent.mkdir(parents=True, exist_ok=True)
    kept_total = 0

    with out_jsonl.open('w', encoding='utf-8') as handle:
        for source_rank, iterator in enumerate(source_iterators, start=1):
            per_source = 0
            for row in iterator:
                source_lane = str(row['sourceLane'])
                source_counts_in[source_lane] += 1

                curated = build_curated_record(
                    source_rank=source_rank,
                    row=row,
                    train_ratio=args.train_ratio,
                    val_ratio=args.val_ratio,
                )

                ingredient_count = int(curated['ingredientCount'])
                if ingredient_count < args.min_ingredients:
                    dropped['too_few_ingredients'] += 1
                    continue
                if ingredient_count > args.max_ingredients:
                    dropped['too_many_ingredients'] += 1
                    continue

                signature = str(curated['signature'])
                if signature in seen_signatures:
                    dropped['dedup_signature'] += 1
                    continue

                if args.max_per_source > 0 and per_source >= args.max_per_source:
                    dropped['max_per_source_cap'] += 1
                    continue

                seen_signatures.add(signature)
                per_source += 1
                source_counts_kept[source_lane] += 1
                split_counts[str(curated['split'])] += 1

                cuisine = str(curated['cuisine'])
                if cuisine:
                    cuisine_counts[cuisine] += 1

                del curated['signature']
                handle.write(json.dumps(curated, ensure_ascii=False) + '\n')
                kept_total += 1

    report = {
        'generatedAtUtc': dt.datetime.now(dt.UTC).replace(microsecond=0).isoformat().replace('+00:00', 'Z'),
        'inputs': dict(source_counts_in),
        'keptBySource': dict(source_counts_kept),
        'dropped': dict(dropped),
        'outputPath': str(out_jsonl),
        'outputRecords': kept_total,
        'splitCounts': dict(split_counts),
        'topCuisines': cuisine_counts.most_common(25),
        'filters': {
            'minIngredients': args.min_ingredients,
            'maxIngredients': args.max_ingredients,
            'maxPerSource': args.max_per_source,
        },
    }

    out_report.parent.mkdir(parents=True, exist_ok=True)
    out_report.write_text(json.dumps(report, indent=2), encoding='utf-8')
    print(f'Wrote curated corpus: {out_jsonl}')
    print(f'Wrote curation report: {out_report}')
    print(f'Output records: {kept_total}')
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
