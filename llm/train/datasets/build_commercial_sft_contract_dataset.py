#!/usr/bin/env python3
"""Build commercial-safe SFT JSON-contract dataset from commercial query corpus."""

from __future__ import annotations

import argparse
import datetime as dt
import hashlib
import json
from collections import Counter
from pathlib import Path
from typing import Dict, List


SYSTEM_PROMPT = (
    'You are an expert cooking assistant. Return JSON only with schema '
    '{"recipes":[{"name":string,"description":string,"ingredients":[{"name":string,"amount":string,"optional":bool}],'
    '"steps":[string],"prepMinutes":number,"cookMinutes":number,"difficulty":"easy|medium|hard",'
    '"cuisine":string,"dietaryTags":[string],"servings":number,"safetyNotes":[string]}]}.'
)

PROTEINS = {
    'chicken',
    'beef',
    'pork',
    'shrimp',
    'salmon',
    'tuna',
    'lamb',
    'egg',
    'eggs',
}

DAIRY = {
    'milk',
    'cheese',
    'butter',
    'cream',
    'yogurt',
    'feta cheese',
    'parmesan cheese',
    'cheddar cheese',
}

ALLERGEN_MAP = {
    'peanut': 'Contains peanut.',
    'milk': 'Contains milk.',
    'egg': 'Contains egg.',
    'eggs': 'Contains egg.',
    'soy': 'Contains soy.',
    'soy sauce': 'Contains soy.',
    'wheat': 'Contains wheat.',
    'shrimp': 'Contains shellfish.',
    'shellfish': 'Contains shellfish.',
    'tree nut': 'Contains tree nuts.',
}

ALLERGEN_KEYWORDS = ['peanut', 'milk', 'egg', 'soy', 'wheat', 'shrimp', 'shellfish', 'tree nut']

SPICE_KEYWORDS = {
    'salt',
    'pepper',
    'paprika',
    'cinnamon',
    'cumin',
    'turmeric',
    'chili',
    'oregano',
    'thyme',
    'rosemary',
}

HERB_KEYWORDS = {'basil', 'mint', 'coriander', 'cilantro', 'parsley', 'dill'}
LIQUID_KEYWORDS = {
    'water',
    'milk',
    'broth',
    'stock',
    'cream',
    'yogurt',
    'soy sauce',
    'vinegar',
    'buttermilk',
    'coconut milk',
}
FAT_KEYWORDS = {'oil', 'butter', 'ghee'}
PROTEIN_KEYWORDS = {'chicken', 'beef', 'pork', 'lamb', 'shrimp', 'salmon', 'tuna', 'tofu'}
GRAIN_KEYWORDS = {'rice', 'pasta', 'flour', 'oats', 'lentils', 'quinoa', 'couscous', 'beans'}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description='Build commercial-safe SFT contract dataset')
    parser.add_argument(
        '--in-jsonl',
        default='llm/train/datasets/raw/commercial-recipe-query-corpus.v1.jsonl',
    )
    parser.add_argument(
        '--out-train',
        default='llm/train/datasets/processed/commercial-sft-json-contract.v1.train.jsonl',
    )
    parser.add_argument(
        '--out-validation',
        default='llm/train/datasets/processed/commercial-sft-json-contract.v1.validation.jsonl',
    )
    parser.add_argument(
        '--out-test',
        default='llm/train/datasets/processed/commercial-sft-json-contract.v1.test.jsonl',
    )
    parser.add_argument(
        '--out-report',
        default='llm/train/datasets/reports/commercial-sft-json-contract.v1.summary.json',
    )
    parser.add_argument('--max-records', type=int, default=120000)
    return parser.parse_args()


def to_title_case(value: str) -> str:
    return ' '.join(token.capitalize() for token in value.split())


def difficulty_for_count(ingredient_count: int) -> str:
    if ingredient_count <= 4:
        return 'easy'
    if ingredient_count <= 6:
        return 'medium'
    return 'hard'


def amount_for_index(index: int) -> str:
    ladder = ['1 cup', '2 tbsp', '1 tsp', '200 g', '2 cloves', '1/2 cup', '1 piece']
    return ladder[index % len(ladder)]


def infer_amount(ingredient: str, index: int) -> str:
    low = ingredient.lower()

    if 'garlic' in low:
        return '2 cloves, minced'
    if 'onion' in low:
        return '1/2 medium, diced'
    if any(token in low for token in PROTEIN_KEYWORDS):
        return '300 g'
    if 'egg' in low:
        return '2 large'
    if 'cheese' in low:
        return '1/2 cup, grated'
    if 'flour' in low:
        return '1/4 cup'
    if any(token in low for token in GRAIN_KEYWORDS):
        return '1 cup'
    if any(token in low for token in FAT_KEYWORDS):
        return '1 tbsp'
    if any(token in low for token in LIQUID_KEYWORDS):
        return '3/4 cup'
    if any(token in low for token in SPICE_KEYWORDS):
        return '1/2 tsp'
    if any(token in low for token in HERB_KEYWORDS):
        return '1/4 cup, chopped'

    return '1 cup, chopped'


def build_safety_notes(ingredients: List[str]) -> List[str]:
    low = {ing.lower() for ing in ingredients}
    ingredient_blob = ' '.join(sorted(low))
    notes: List[str] = []

    if 'chicken' in low:
        notes.append('Cook chicken to 165F/74C before serving.')
    if 'kidney bean' in low or 'kidney beans' in low:
        notes.append('Boil kidney beans thoroughly before simmering.')
    if 'rice' in low:
        notes.append('Cool rice leftovers quickly and refrigerate promptly.')

    for key, note in ALLERGEN_MAP.items():
        if key in low:
            notes.append(note)

    if any(keyword in ingredient_blob for keyword in ALLERGEN_KEYWORDS):
        notes.append('Contains common allergens; check ingredient labels before serving.')

    if not notes:
        notes.append('Use clean utensils and cook to safe temperatures.')

    deduped: List[str] = []
    seen = set()
    for note in notes:
        if note in seen:
            continue
        seen.add(note)
        deduped.append(note)
    return deduped


def build_dietary_tags(ingredients: List[str]) -> List[str]:
    low = {ing.lower() for ing in ingredients}
    tags: List[str] = []
    if not any(item in low for item in PROTEINS):
        tags.append('vegetarian')
    if not any(item in low for item in DAIRY):
        tags.append('dairy-free')
    return tags


def build_recipe(cuisine: str, ingredients: List[str], ingredient_count: int) -> Dict[str, object]:
    anchor = to_title_case(ingredients[0])
    cuisine_title = to_title_case(cuisine) if cuisine else 'Everyday'
    difficulty = difficulty_for_count(ingredient_count)
    prep_minutes = 10 + ingredient_count * 2
    cook_minutes = 12 + ingredient_count * 3

    recipe_ingredients = []
    for idx, ingredient in enumerate(ingredients):
        recipe_ingredients.append(
            {
                'name': ingredient,
                'amount': infer_amount(ingredient, idx),
                'optional': idx >= max(3, ingredient_count - 2),
            }
        )

    required_ingredients = [ing for ing in recipe_ingredients if not ing['optional']]
    highlighted = required_ingredients[:4] if required_ingredients else recipe_ingredients[:4]
    highlighted_names = [str(ing['name']) for ing in highlighted]

    aromatic = next((str(ing['name']) for ing in recipe_ingredients if 'onion' in str(ing['name']) or 'garlic' in str(ing['name'])), '')
    protein = next((str(ing['name']) for ing in recipe_ingredients if any(token in str(ing['name']) for token in PROTEIN_KEYWORDS)), '')
    liquid = next((str(ing['name']) for ing in recipe_ingredients if any(token in str(ing['name']) for token in LIQUID_KEYWORDS)), '')
    herb = next((str(ing['name']) for ing in recipe_ingredients if any(token in str(ing['name']) for token in HERB_KEYWORDS)), '')
    sturdy = next(
        (
            str(ing['name'])
            for ing in recipe_ingredients
            if not any(token in str(ing['name']) for token in LIQUID_KEYWORDS | SPICE_KEYWORDS | HERB_KEYWORDS | FAT_KEYWORDS)
        ),
        highlighted_names[0],
    )

    cook_window = max(10, cook_minutes - 6)
    simmer_minutes = max(6, min(14, cook_window // 2))

    step_templates = [
        f"Prep {', '.join(highlighted_names)}: wash and cut into even bite-size pieces so they cook at the same pace.",
        (f"Heat 1 tbsp oil over medium heat for 1 minute, then cook {aromatic} for 3-4 minutes until softened and aromatic." if aromatic else "Heat 1 tbsp oil over medium heat for 1 minute, then add sturdy vegetables and cook 3-4 minutes until lightly golden."),
        (
            f"Add {protein} and cook for 5-7 minutes, stirring occasionally, until lightly browned and nearly cooked through."
            if protein
            else f"Add {sturdy} and cook for 4-6 minutes, stirring occasionally, until tender with light caramelization."
        ),
        (
            f"Stir in remaining ingredients and toast spices for 60-90 seconds so the dish develops deeper flavor."
        ),
        (
            f"Pour in {liquid or '3/4 cup water'}, bring to a gentle simmer, then cook uncovered for {simmer_minutes} minutes until the sauce thickens and ingredients are tender."
        ),
        (
            f"Finish with {herb or 'fresh herbs'} and adjust salt/acidity to taste before serving warm."
        ),
    ]

    if protein == 'chicken':
        step_templates[2] = (
            'Add chicken and cook for 6-8 minutes until opaque; finish cooking until the thickest piece reaches 165F/74C.'
        )

    variation = int(hashlib.sha256('|'.join(ingredients).encode('utf-8')).hexdigest()[:2], 16) % 3
    if variation == 1:
        step_templates[3] = 'Add dense vegetables first, then quick-cooking items last, stirring between additions to keep texture balanced.'
    elif variation == 2:
        step_templates[5] = f"Rest off heat for 2 minutes, then fold in {herb or 'fresh garnish'} and serve while hot."

    steps = step_templates

    return {
        'name': f'{cuisine_title} {anchor} Bowl',
        'description': f'Balanced {cuisine_title.lower()} style dish built from pantry-safe ingredients.',
        'ingredients': recipe_ingredients,
        'steps': steps,
        'prepMinutes': prep_minutes,
        'cookMinutes': cook_minutes,
        'difficulty': difficulty,
        'cuisine': cuisine if cuisine else 'global',
        'dietaryTags': build_dietary_tags(ingredients),
        'servings': 2 if ingredient_count <= 5 else 4,
        'safetyNotes': build_safety_notes(ingredients),
    }


def build_user_prompt(ingredients: List[str], mode: str, cuisine: str) -> str:
    filters = {'cuisine': [cuisine]} if cuisine else {}
    filters_json = json.dumps(filters, separators=(',', ':')) if filters else 'none'
    return (
        f'Generate 1-3 recipes for these normalized ingredients: {", ".join(ingredients)}.\n'
        f'Search mode: {mode}.\n'
        f'Filters: {filters_json}.\n'
        'Return JSON ONLY with the required schema.'
    )


def main() -> int:
    args = parse_args()
    in_path = Path(args.in_jsonl)
    out_train = Path(args.out_train)
    out_validation = Path(args.out_validation)
    out_test = Path(args.out_test)
    out_report = Path(args.out_report)

    out_train.parent.mkdir(parents=True, exist_ok=True)
    out_validation.parent.mkdir(parents=True, exist_ok=True)
    out_test.parent.mkdir(parents=True, exist_ok=True)
    out_report.parent.mkdir(parents=True, exist_ok=True)

    split_paths = {
        'train': out_train,
        'validation': out_validation,
        'test': out_test,
    }
    split_handles = {
        split: path.open('w', encoding='utf-8') for split, path in split_paths.items()
    }

    split_counts: Counter[str] = Counter()
    difficulty_counts: Counter[str] = Counter()
    source_lane_counts: Counter[str] = Counter()
    total = 0

    try:
        with in_path.open('r', encoding='utf-8') as handle:
            for line in handle:
                if total >= args.max_records:
                    break
                raw = line.strip()
                if not raw:
                    continue
                row = json.loads(raw)
                if not isinstance(row, dict):
                    continue

                split = str(row.get('split', 'train'))
                if split not in split_handles:
                    continue

                ingredients = row.get('ingredients', [])
                if not isinstance(ingredients, list) or not ingredients:
                    continue
                ingredients = [str(item).strip().lower() for item in ingredients if str(item).strip()]
                if not ingredients:
                    continue

                cuisine = str(row.get('cuisine', '')).strip().lower()
                mode = 'inclusive'
                ingredient_count = len(ingredients)
                recipe = build_recipe(cuisine, ingredients, ingredient_count)
                difficulty_counts[str(recipe['difficulty'])] += 1

                source_lane = str(row.get('sourceLane', ''))
                source_lane_counts[source_lane] += 1

                out_row = {
                    'id': f"sft-{row.get('id', total)}",
                    'lane': 'sft-json-contract',
                    'sourceLane': source_lane,
                    'sourcePath': row.get('sourcePath', ''),
                    'sourceRecordId': row.get('sourceRecordId', ''),
                    'split': split,
                    'system': SYSTEM_PROMPT,
                    'user': build_user_prompt(ingredients, mode, cuisine),
                    'assistant': {'recipes': [recipe]},
                    'metadata': {
                        'generationMethod': 'synthetic-first-party',
                        'commercialSafe': True,
                        'schemaVersion': 'fallback-contract-v1',
                    },
                }

                split_handles[split].write(json.dumps(out_row, ensure_ascii=False) + '\n')
                split_counts[split] += 1
                total += 1
    finally:
        for handle in split_handles.values():
            handle.close()

    report = {
        'generatedAtUtc': dt.datetime.now(dt.UTC).replace(microsecond=0).isoformat().replace('+00:00', 'Z'),
        'inputPath': str(in_path),
        'maxRecordsRequested': args.max_records,
        'totalRecordsWritten': total,
        'outputs': {
            'train': str(out_train),
            'validation': str(out_validation),
            'test': str(out_test),
        },
        'splitCounts': dict(split_counts),
        'difficultyCounts': dict(difficulty_counts),
        'sourceLaneCounts': dict(source_lane_counts),
        'commercialPolicy': {
            'sourceLaneRequired': 'internal-synthetic-query-v1',
            'thirdPartyRecipeTextIncluded': False,
        },
    }

    out_report.write_text(json.dumps(report, indent=2), encoding='utf-8')
    print(f'Wrote SFT summary report: {out_report}')
    print(f'Total records written: {total}')
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
