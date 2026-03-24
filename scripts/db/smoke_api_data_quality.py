#!/usr/bin/env python3
"""Run API-level data quality smoke checks for catalog/search endpoints."""

from __future__ import annotations

import argparse
import datetime as dt
import json
import sys
import urllib.error
import urllib.request
from pathlib import Path


def get_json(url: str, timeout: float = 10.0) -> tuple[int, object]:
    req = urllib.request.Request(
        url,
        method='GET',
        headers={
            'User-Agent': 'IngredientialOpsSmoke/1.0',
            'Accept': 'application/json',
        },
    )
    with urllib.request.urlopen(req, timeout=timeout) as response:
        body = response.read().decode('utf-8', errors='replace')
        return response.status, json.loads(body)


def post_json(url: str, payload: dict[str, object], timeout: float = 10.0) -> tuple[int, object]:
    raw = json.dumps(payload).encode('utf-8')
    req = urllib.request.Request(
        url,
        data=raw,
        method='POST',
        headers={
            'Content-Type': 'application/json',
            'Accept': 'application/json',
            'User-Agent': 'IngredientialOpsSmoke/1.0',
        },
    )
    with urllib.request.urlopen(req, timeout=timeout) as response:
        body = response.read().decode('utf-8', errors='replace')
        return response.status, json.loads(body)


def as_dict(value: object) -> dict[str, object]:
    if not isinstance(value, dict):
        raise ValueError('Expected JSON object response')
    return value


def as_list(value: object) -> list[object]:
    if not isinstance(value, list):
        raise ValueError('Expected JSON array field')
    return value


def validate_catalog_response(data: dict[str, object], min_total: int, min_items: int) -> list[str]:
    failures: list[str] = []

    total = data.get('total')
    items = data.get('items')
    page = data.get('page')
    page_size = data.get('pageSize')

    if not isinstance(total, int):
        failures.append('catalog.total is missing or not an int')
    elif total < min_total:
        failures.append(f'catalog.total below threshold: {total} < {min_total}')

    if not isinstance(page, int) or page <= 0:
        failures.append('catalog.page is missing or invalid')
    if not isinstance(page_size, int) or page_size <= 0:
        failures.append('catalog.pageSize is missing or invalid')

    if not isinstance(items, list):
        failures.append('catalog.items is missing or not a list')
        return failures

    if len(items) < min_items:
        failures.append(f'catalog.items below threshold: {len(items)} < {min_items}')

    if items:
        first = items[0]
        if not isinstance(first, dict):
            failures.append('catalog.items[0] is not an object')
        else:
            required_str = ['id', 'name', 'source', 'difficulty', 'updatedAt']
            for key in required_str:
                if not isinstance(first.get(key), str) or not str(first.get(key)).strip():
                    failures.append(f'catalog.items[0].{key} missing/invalid')

            required_num = ['totalMinutes', 'servings', 'qualityScore']
            for key in required_num:
                value = first.get(key)
                if not isinstance(value, (int, float)):
                    failures.append(f'catalog.items[0].{key} missing/invalid')

    return failures


def validate_search_response(data: dict[str, object], min_results: int) -> list[str]:
    failures: list[str] = []

    mode = data.get('mode')
    query = data.get('query')
    pagination = data.get('pagination')
    results = data.get('results')

    if mode not in {'strict', 'inclusive'}:
        failures.append('search.mode is missing or invalid')

    if not isinstance(query, dict):
        failures.append('search.query is missing or invalid')
    else:
        ingredients = query.get('ingredients')
        if not isinstance(ingredients, list) or not ingredients:
            failures.append('search.query.ingredients missing/empty')

    if not isinstance(pagination, dict):
        failures.append('search.pagination is missing or invalid')
    else:
        if not isinstance(pagination.get('page'), int):
            failures.append('search.pagination.page missing/invalid')
        if not isinstance(pagination.get('pageSize'), int):
            failures.append('search.pagination.pageSize missing/invalid')
        if not isinstance(pagination.get('total'), int):
            failures.append('search.pagination.total missing/invalid')

    if not isinstance(results, list):
        failures.append('search.results is missing or not a list')
        return failures

    if len(results) < min_results:
        failures.append(f'search.results below threshold: {len(results)} < {min_results}')

    if results:
        first = results[0]
        if not isinstance(first, dict):
            failures.append('search.results[0] is not an object')
        else:
            for key in ['id', 'name', 'source', 'difficulty']:
                if not isinstance(first.get(key), str) or not str(first.get(key)).strip():
                    failures.append(f'search.results[0].{key} missing/invalid')
            if not isinstance(first.get('matchPercent'), (int, float)):
                failures.append('search.results[0].matchPercent missing/invalid')

    return failures


def write_report(
    out_path: Path,
    base_url: str,
    catalog_url: str,
    search_url: str,
    ingredient: str,
    catalog_total: int,
    catalog_items: int,
    search_total: int,
    search_results: int,
    failures: list[str],
) -> None:
    now = dt.datetime.now().isoformat(timespec='seconds')
    lines = [
        '# V1 API Data Quality Smoke Report',
        '',
        f'Generated: `{now}`',
        f'Base URL: `{base_url}`',
        '',
        '## Endpoint checks',
        '',
        f'- `GET {catalog_url}` -> total={catalog_total}, items={catalog_items}',
        f"- `POST {search_url}` ingredient=`{ingredient}` -> total={search_total}, results={search_results}",
        '',
        f"Overall: **{'PASS' if not failures else 'FAIL'}**",
    ]

    if failures:
        lines.extend(['', '## Failures', ''])
        for failure in failures:
            lines.append(f'- {failure}')

    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text('\n'.join(lines) + '\n', encoding='utf-8')


def main() -> int:
    parser = argparse.ArgumentParser(description='Run API data quality smoke checks')
    parser.add_argument('--base-url', default='https://api.ingrediential.uk')
    parser.add_argument('--ingredient', default='chicken')
    parser.add_argument('--page-size', type=int, default=5)
    parser.add_argument('--search-mode', default='inclusive', choices=['strict', 'inclusive'])
    parser.add_argument('--min-catalog-total', type=int, default=1)
    parser.add_argument('--min-catalog-items', type=int, default=1)
    parser.add_argument('--min-search-results', type=int, default=1)
    parser.add_argument(
        '--out',
        default='docs/ops/v1-api-data-quality-smoke-latest.md',
        help='Output markdown path relative to repo root',
    )
    args = parser.parse_args()

    repo_root = Path(__file__).resolve().parents[2]
    out_path = (repo_root / args.out).resolve()
    base_url = args.base_url.rstrip('/')

    catalog_url = f'{base_url}/recipes/catalog?page=1&pageSize={args.page_size}'
    search_url = f'{base_url}/recipes/search'

    failures: list[str] = []
    catalog_total = 0
    catalog_items_count = 0
    search_total = 0
    search_results_count = 0

    try:
        catalog_status, catalog_raw = get_json(catalog_url)
        if catalog_status != 200:
            failures.append(f'catalog endpoint returned status {catalog_status}')
        catalog = as_dict(catalog_raw)
        failures.extend(
            validate_catalog_response(catalog, args.min_catalog_total, args.min_catalog_items)
        )
        catalog_total_value = catalog.get('total')
        catalog_total = catalog_total_value if isinstance(catalog_total_value, int) else 0
        catalog_items = as_list(catalog.get('items', []))
        catalog_items_count = len(catalog_items)
    except urllib.error.HTTPError as exc:
        failures.append(f'catalog endpoint HTTP error: {exc.code}')
    except Exception as exc:  # noqa: BLE001
        failures.append(f'catalog endpoint error: {exc}')

    payload = {
        'ingredients': [args.ingredient],
        'mode': args.search_mode,
        'dbOnly': True,
        'debugNoCache': True,
        'pagination': {'page': 1, 'pageSize': args.page_size},
    }

    try:
        search_status, search_raw = post_json(search_url, payload)
        if search_status != 200:
            failures.append(f'search endpoint returned status {search_status}')
        search = as_dict(search_raw)
        failures.extend(validate_search_response(search, args.min_search_results))
        pagination = as_dict(search.get('pagination', {}))
        search_total_value = pagination.get('total')
        search_total = search_total_value if isinstance(search_total_value, int) else 0
        search_results = as_list(search.get('results', []))
        search_results_count = len(search_results)
    except urllib.error.HTTPError as exc:
        failures.append(f'search endpoint HTTP error: {exc.code}')
    except Exception as exc:  # noqa: BLE001
        failures.append(f'search endpoint error: {exc}')

    write_report(
        out_path=out_path,
        base_url=base_url,
        catalog_url=catalog_url,
        search_url=search_url,
        ingredient=args.ingredient,
        catalog_total=catalog_total,
        catalog_items=catalog_items_count,
        search_total=search_total,
        search_results=search_results_count,
        failures=failures,
    )

    print(f'Wrote API data quality smoke report: {out_path}')
    if failures:
        for failure in failures:
            print(f'FAIL: {failure}', file=sys.stderr)
        return 1

    return 0


if __name__ == '__main__':
    raise SystemExit(main())
