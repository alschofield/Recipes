#!/usr/bin/env python3
"""Build a large commercial-safe recipe-query corpus from first-party ingredient seed data."""

from __future__ import annotations

import argparse
import csv
import datetime as dt
import hashlib
import json
import random
from collections import Counter, defaultdict
from pathlib import Path
from typing import Dict, Iterable, List


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description='Generate commercial-safe recipe-query corpus from first-party seed data'
    )
    parser.add_argument(
        '--seed-csv',
        default='datasets/derived/server-lib/canonical_ingredient_seed_v1.csv',
    )
    parser.add_argument(
        '--out-jsonl',
        default='llm/train/datasets/raw/commercial-recipe-query-corpus.v1.jsonl',
    )
    parser.add_argument(
        '--out-report',
        default='llm/train/datasets/reports/commercial-recipe-query-corpus.v1.summary.json',
    )
    parser.add_argument('--target-size', type=int, default=120000)
    parser.add_argument('--min-ingredients', type=int, default=3)
    parser.add_argument('--max-ingredients', type=int, default=8)
    parser.add_argument('--min-quality-score', type=float, default=0.75)
    parser.add_argument('--seed', type=int, default=20260324)
    parser.add_argument('--train-ratio', type=float, default=0.7)
    parser.add_argument('--val-ratio', type=float, default=0.15)
    parser.add_argument('--test-ratio', type=float, default=0.15)
    return parser.parse_args()


def normalize(value: str) -> str:
    return ' '.join(value.strip().lower().split())


def split_for_signature(signature: str, train_ratio: float, val_ratio: float) -> str:
    bucket = int(hashlib.sha256(signature.encode('utf-8')).hexdigest()[:8], 16) / 0xFFFFFFFF
    if bucket < train_ratio:
        return 'train'
    if bucket < train_ratio + val_ratio:
        return 'validation'
    return 'test'


def pick_unique(pool: List[str], count: int, rng: random.Random, blocked: set[str] | None = None) -> List[str]:
    blocked = blocked or set()
    available = [item for item in pool if item not in blocked]
    if len(available) <= count:
        return available
    return rng.sample(available, count)


def build_query_text(cuisine: str, ingredients: List[str], profile: str) -> str:
    joined = ', '.join(ingredients)
    if profile == 'quick':
        if cuisine:
            return f'Need a quick {cuisine} meal idea using {joined}.'
        return f'Need a quick meal idea using {joined}.'
    if profile == 'prep':
        if cuisine:
            return f'Create a prep-friendly {cuisine} recipe plan with {joined}.'
        return f'Create a prep-friendly recipe plan with {joined}.'
    if cuisine:
        return f'Give me recipe ideas for {cuisine} cuisine using {joined}.'
    return f'Give me recipe ideas using {joined}.'


def load_seed_rows(seed_csv: Path, min_quality_score: float) -> List[Dict[str, str]]:
    rows: List[Dict[str, str]] = []
    with seed_csv.open('r', encoding='utf-8', newline='') as handle:
        reader = csv.DictReader(handle)
        for row in reader:
            name = normalize(str(row.get('canonical_name', '')))
            category = normalize(str(row.get('category', '')))
            status = normalize(str(row.get('analysis_status', '')))
            quality_raw = str(row.get('quality_score', '0')).strip()
            try:
                quality = float(quality_raw)
            except ValueError:
                quality = 0.0

            if not name or not category:
                continue
            if status not in {'enriched', 'pending'}:
                continue
            if quality < min_quality_score:
                continue

            rows.append({'name': name, 'category': category, 'quality_score': f'{quality:.3f}'})
    return rows


def build_signature(ingredients: Iterable[str]) -> str:
    payload = '|'.join(sorted(set(ingredients)))
    return hashlib.sha256(payload.encode('utf-8')).hexdigest()


def main() -> int:
    args = parse_args()
    ratio_sum = args.train_ratio + args.val_ratio + args.test_ratio
    if abs(ratio_sum - 1.0) > 1e-6:
        raise ValueError('train/val/test ratios must sum to 1.0')
    if args.target_size <= 0:
        raise ValueError('--target-size must be > 0')
    if args.min_ingredients < 2:
        raise ValueError('--min-ingredients must be >= 2')
    if args.max_ingredients < args.min_ingredients:
        raise ValueError('--max-ingredients must be >= --min-ingredients')

    seed_csv = Path(args.seed_csv)
    out_jsonl = Path(args.out_jsonl)
    out_report = Path(args.out_report)
    rng = random.Random(args.seed)

    rows = load_seed_rows(seed_csv, args.min_quality_score)
    if len(rows) < args.max_ingredients:
        raise RuntimeError('Not enough high-quality ingredients to build requested corpus')

    by_category: Dict[str, List[str]] = defaultdict(list)
    all_names: List[str] = []
    for row in rows:
        name = row['name']
        category = row['category']
        by_category[category].append(name)
        all_names.append(name)

    protein_like = {'meat', 'seafood', 'legume', 'dairy', 'egg'}
    protein_pool: List[str] = []
    for category, names in by_category.items():
        if any(token in category for token in protein_like):
            protein_pool.extend(names)
    if not protein_pool:
        protein_pool = list(all_names)

    cuisines = [
        '',
        'mediterranean',
        'asian',
        'indian',
        'latin',
        'middle eastern',
        'nordic',
        'british',
        'caribbean',
        'african',
    ]
    profiles = ['default', 'quick', 'prep']

    out_jsonl.parent.mkdir(parents=True, exist_ok=True)

    seen_signatures = set()
    split_counts: Counter[str] = Counter()
    cuisine_counts: Counter[str] = Counter()
    ingredient_count_hist: Counter[str] = Counter()

    with out_jsonl.open('w', encoding='utf-8') as handle:
        attempts = 0
        kept = 0
        while kept < args.target_size:
            attempts += 1
            if attempts > args.target_size * 50:
                raise RuntimeError('Failed to generate enough unique records within attempt budget')

            ingredient_count = rng.randint(args.min_ingredients, args.max_ingredients)
            anchor = rng.choice(protein_pool)
            rest = pick_unique(all_names, ingredient_count - 1, rng, blocked={anchor})
            ingredients = sorted({anchor, *rest})

            if len(ingredients) < args.min_ingredients:
                continue

            signature = build_signature(ingredients)
            if signature in seen_signatures:
                continue
            seen_signatures.add(signature)

            cuisine = rng.choice(cuisines)
            profile = rng.choice(profiles)
            split = split_for_signature(signature, args.train_ratio, args.val_ratio)

            row_id = hashlib.sha256(f'commercial-v1:{signature}'.encode('utf-8')).hexdigest()[:16]
            record = {
                'id': f'commercial-query-v1-{row_id}',
                'lane': 'recipe-query',
                'sourceLane': 'internal-synthetic-query-v1',
                'sourcePath': 'datasets/derived/server-lib/canonical_ingredient_seed_v1.csv',
                'sourceRecordId': row_id,
                'title': '',
                'cuisine': cuisine,
                'ingredients': ingredients,
                'ingredientCount': len(ingredients),
                'queryText': build_query_text(cuisine, ingredients, profile),
                'split': split,
                'generationMethod': 'synthetic-first-party',
                'profile': profile,
            }
            handle.write(json.dumps(record, ensure_ascii=False) + '\n')

            kept += 1
            split_counts[split] += 1
            cuisine_counts[cuisine or 'unspecified'] += 1
            ingredient_count_hist[str(len(ingredients))] += 1

    report = {
        'generatedAtUtc': dt.datetime.now(dt.UTC).replace(microsecond=0).isoformat().replace('+00:00', 'Z'),
        'seedCsv': str(seed_csv),
        'outputPath': str(out_jsonl),
        'outputRecords': args.target_size,
        'minQualityScore': args.min_quality_score,
        'generationSeed': args.seed,
        'splitCounts': dict(split_counts),
        'cuisineCounts': dict(cuisine_counts),
        'ingredientCountHistogram': dict(ingredient_count_hist),
        'sourceLane': 'internal-synthetic-query-v1',
        'commercialPolicy': {
            'commercialUseAllowed': True,
            'approvedForFineTune': True,
            'thirdPartyRecipeTextIncluded': False,
        },
    }

    out_report.parent.mkdir(parents=True, exist_ok=True)
    out_report.write_text(json.dumps(report, indent=2), encoding='utf-8')
    print(f'Wrote commercial-safe corpus: {out_jsonl}')
    print(f'Wrote commercial-safe report: {out_report}')
    print(f'Output records: {args.target_size}')
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
