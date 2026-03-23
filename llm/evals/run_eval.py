#!/usr/bin/env python3
"""Lightweight LLM evaluator for recipe-generation contract compliance.

Targets OpenAI-compatible chat endpoints (including local Ollama `/v1`).
"""

from __future__ import annotations

import argparse
import json
import math
import os
import re
import statistics
import time
import unicodedata
import urllib.error
import urllib.request
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Dict, Iterable, List, Optional, Tuple


RECIPE_KEYS = {
    "name",
    "description",
    "ingredients",
    "steps",
    "prepMinutes",
    "cookMinutes",
    "difficulty",
    "cuisine",
    "dietaryTags",
    "servings",
    "safetyNotes",
}

INGREDIENT_KEYS = {"name", "amount", "optional"}
DIFFICULTIES = {"easy", "medium", "hard"}
COMMON_ALLERGENS = {
    "peanut",
    "milk",
    "egg",
    "soy",
    "wheat",
    "shrimp",
    "shellfish",
    "tree nut",
}

TECHNIQUE_PATTERNS = {
    "sear": r"\bsear(?:ed|ing)?\b",
    "roast": r"\broast(?:ed|ing)?\b",
    "braise": r"\bbrais(?:e|ed|ing)\b",
    "grill": r"\bgrill(?:ed|ing)?\b",
    "smoke": r"\bsmok(?:e|ed|ing)\b",
    "steam": r"\bsteam(?:ed|ing)?\b",
    "poach": r"\bpoach(?:ed|ing)?\b",
    "bake": r"\bbak(?:e|ed|ing)\b",
    "saute": r"\bsaute(?:ed|ing)?\b",
    "deglaze": r"\bdeglaz(?:e|ed|ing)\b",
    "reduce": r"\breduc(?:e|ed|ing|tion)\b",
    "emulsify": r"\bemuls(?:ify|ified|ifying|ion)\b",
    "ferment": r"\bferment(?:ed|ing)?\b",
    "blanch": r"\bblanch(?:ed|ing)?\b",
    "caramelize": r"\bcarameli(?:ze|zed|zing|sation|zation)\b",
    "marinate": r"\bmarinat(?:e|ed|ing|ion)\b",
    "pickle": r"\bpickl(?:e|ed|ing)\b",
}

SAFE_CHICKEN_TEMP_PATTERN = re.compile(r"(165\s*\u00b0?\s*f|74\s*\u00b0?\s*c)", re.IGNORECASE)

BANNED_SAFETY_PHRASES = {
    "eat raw chicken",
    "undercook chicken",
    "ignore allergies",
    "no need to wash hands",
}


@dataclass
class EvalCase:
    case_id: str
    normalized_ingredients: List[str]
    mode: str
    filters: Dict[str, Any]
    extra_instruction: str
    case_type: str
    min_recipes: int
    min_ingredients_per_recipe: int
    min_steps_per_recipe: int
    min_technique_keywords: int


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Run recipe LLM eval suite")
    parser.add_argument("--base-url", required=True, help="OpenAI-compatible base URL")
    parser.add_argument("--api-key", required=True, help="Auth token (can be dummy for local)")
    parser.add_argument("--model", required=True, help="Model name")
    parser.add_argument("--recipe-cases", required=True, help="Path to recipe JSONL")
    parser.add_argument("--safety-cases", required=True, help="Path to safety JSONL")
    parser.add_argument(
        "--complex-cases",
        required=False,
        default="",
        help="Optional path to complex recipe JSONL",
    )
    parser.add_argument("--out", required=True, help="Output directory")
    parser.add_argument("--temperature", type=float, default=0.2)
    parser.add_argument("--timeout-seconds", type=int, default=60)
    parser.add_argument(
        "--prompt-profile",
        choices=["schema_first", "safety_complex_first"],
        default="schema_first",
        help="Prompt style to evaluate",
    )
    parser.add_argument(
        "--compare-profiles",
        action="store_true",
        help="Run both prompt profiles and rank by weighted score",
    )
    parser.add_argument(
        "--enable-safety-repair",
        action="store_true",
        help="Attempt one same-model repair call for safety-case failures",
    )
    parser.add_argument(
        "--repair-timeout-seconds",
        type=int,
        default=45,
        help="Timeout for safety repair pass",
    )
    parser.add_argument(
        "--disable-thinking-tag",
        action="store_true",
        help="Prefix prompts with /no_think for models that support it (e.g., Qwen3)",
    )
    return parser.parse_args()


def read_jsonl(path: str, case_type: str) -> List[EvalCase]:
    cases: List[EvalCase] = []
    with open(path, "r", encoding="utf-8") as f:
        for line_num, line in enumerate(f, start=1):
            raw = line.strip()
            if not raw:
                continue
            obj = json.loads(raw)
            case_id = str(obj.get("id", f"{case_type}-{line_num}"))
            cases.append(
                EvalCase(
                    case_id=case_id,
                    normalized_ingredients=list(obj.get("normalizedIngredients", [])),
                    mode=str(obj.get("mode", "inclusive")),
                    filters=dict(obj.get("filters", {})),
                    extra_instruction=str(obj.get("extraInstruction", "")).strip(),
                    case_type=case_type,
                    min_recipes=int(obj.get("minRecipes", 1)),
                    min_ingredients_per_recipe=int(obj.get("minIngredientsPerRecipe", 0)),
                    min_steps_per_recipe=int(obj.get("minStepsPerRecipe", 0)),
                    min_technique_keywords=int(obj.get("minTechniqueKeywords", 0)),
                )
            )
    return cases


def evaluate_complexity(case: EvalCase, parsed: Dict[str, Any]) -> Tuple[bool, List[str]]:
    errors: List[str] = []
    recipes = parsed.get("recipes", [])
    if not isinstance(recipes, list):
        return False, ["recipes is not a list"]

    if len(recipes) < case.min_recipes:
        errors.append(f"expected at least {case.min_recipes} recipes, got {len(recipes)}")

    technique_hits = set()
    for idx, recipe in enumerate(recipes):
        ingredients = recipe.get("ingredients", [])
        steps = recipe.get("steps", [])

        if case.min_ingredients_per_recipe > 0 and len(ingredients) < case.min_ingredients_per_recipe:
            errors.append(
                f"recipes[{idx}] needs >= {case.min_ingredients_per_recipe} ingredients, got {len(ingredients)}"
            )

        if case.min_steps_per_recipe > 0 and len(steps) < case.min_steps_per_recipe:
            errors.append(
                f"recipes[{idx}] needs >= {case.min_steps_per_recipe} steps, got {len(steps)}"
            )

        steps_text = "\n".join(str(s).lower() for s in steps)
        steps_text = unicodedata.normalize("NFKD", steps_text)
        steps_text = "".join(ch for ch in steps_text if not unicodedata.combining(ch))
        for label, pattern in TECHNIQUE_PATTERNS.items():
            if re.search(pattern, steps_text):
                technique_hits.add(label)

    if case.min_technique_keywords > 0 and len(technique_hits) < case.min_technique_keywords:
        errors.append(
            f"expected >= {case.min_technique_keywords} technique keywords, got {len(technique_hits)}"
        )

    return len(errors) == 0, errors


def build_prompt(
    case: EvalCase,
    disable_thinking_tag: bool = False,
    prompt_profile: str = "schema_first",
) -> str:
    filters = json.dumps(case.filters, separators=(",", ":")) if case.filters else "none"
    suffix = ""
    if case.extra_instruction:
        suffix = f"\nExtra instruction: {case.extra_instruction}\n"
    recipe_count_instruction = "Generate 1-3 recipes"
    case_requirements: List[str] = []

    if case.case_type == "complex" and prompt_profile == "safety_complex_first":
        recipe_count_instruction = f"Generate at least {max(2, case.min_recipes)} recipes"
        case_requirements.append(
            f"Each recipe must include at least {max(1, case.min_ingredients_per_recipe)} ingredients and {max(1, case.min_steps_per_recipe)} steps."
        )
        case_requirements.append(
            f"Across the response, include at least {max(1, case.min_technique_keywords)} distinct technique verbs (for example: marinate, sear, roast, braise, bake, deglaze, reduce, pickle)."
        )
        case_requirements.append("Use staged prep + finishing flow with explicit sequencing in steps.")

    if case.case_type == "safety" and prompt_profile == "safety_complex_first":
        case_requirements.append("Every recipe must include at least one concrete safety note in safetyNotes.")
        case_requirements.append("If chicken appears, include internal temperature guidance as 74C/165F before serving.")
        case_requirements.append("If kidney beans appear, include explicit boil guidance before simmering.")
        case_requirements.append("If common allergens appear, include allergen warning language.")
        case_requirements.append("If rice appears, include a leftover cooling/refrigeration reminder.")

    requirements_block = ""
    if case_requirements:
        requirements_block = "\nCase-specific requirements:\n- " + "\n- ".join(case_requirements) + "\n"

    prompt = (
        f"{recipe_count_instruction} for these normalized ingredients: {', '.join(case.normalized_ingredients)}.\n"
        f"Search mode: {case.mode}.\n"
        f"Filters: {filters}.\n"
        f"{suffix}"
        f"{requirements_block}\n"
        "Return JSON ONLY with this exact schema:\n"
        "{\n"
        '  "recipes": [\n'
        "    {\n"
        '      "name": "string",\n'
        '      "description": "string",\n'
        '      "ingredients": [\n'
        '        { "name": "string", "amount": "string", "optional": false }\n'
        "      ],\n"
        '      "steps": ["string"],\n'
        '      "prepMinutes": 30,\n'
        '      "cookMinutes": 20,\n'
        '      "difficulty": "easy|medium|hard",\n'
        '      "cuisine": "string",\n'
        '      "dietaryTags": ["string"],\n'
        '      "servings": 2,\n'
        '      "safetyNotes": ["string"]\n'
        "    }\n"
        "  ]\n"
        "}\n"
        "Do not include markdown, code fences, or any extra fields."
    )
    if disable_thinking_tag:
        return f"/no_think\n{prompt}"
    return prompt


def call_model(
    base_url: str,
    api_key: str,
    model: str,
    prompt: str,
    temperature: float,
    timeout_seconds: int,
) -> Tuple[Optional[str], Optional[int], float, Optional[str]]:
    url = base_url.rstrip("/") + "/chat/completions"
    body = {
        "model": model,
        "messages": [
            {
                "role": "system",
                "content": "You generate safe, practical cooking recipes. Respond with valid JSON only.",
            },
            {"role": "user", "content": prompt},
        ],
        "temperature": temperature,
        "response_format": {"type": "json_object"},
    }
    payload = json.dumps(body).encode("utf-8")
    req = urllib.request.Request(
        url,
        data=payload,
        method="POST",
        headers={
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json",
        },
    )

    start = time.perf_counter()
    try:
        with urllib.request.urlopen(req, timeout=timeout_seconds) as resp:
            latency_ms = (time.perf_counter() - start) * 1000.0
            status = getattr(resp, "status", 200)
            raw = resp.read().decode("utf-8")
            parsed = json.loads(raw)
            content = (
                parsed.get("choices", [{}])[0]
                .get("message", {})
                .get("content", "")
            )
            if not content or not isinstance(content, str):
                return None, status, latency_ms, "empty model content"
            return content, status, latency_ms, None
    except urllib.error.HTTPError as e:
        latency_ms = (time.perf_counter() - start) * 1000.0
        return None, int(e.code), latency_ms, f"http error: {e.reason}"
    except Exception as e:  # noqa: BLE001
        latency_ms = (time.perf_counter() - start) * 1000.0
        return None, None, latency_ms, str(e)


def _is_non_empty_string(v: Any) -> bool:
    return isinstance(v, str) and bool(v.strip())


def validate_schema(raw_content: str) -> Tuple[bool, List[str], Optional[Dict[str, Any]]]:
    errors: List[str] = []
    try:
        obj = json.loads(raw_content)
    except json.JSONDecodeError as e:
        return False, [f"invalid JSON: {e}"], None

    if not isinstance(obj, dict):
        return False, ["top-level value must be object"], None
    if set(obj.keys()) != {"recipes"}:
        errors.append("top-level object must contain only `recipes`")

    recipes = obj.get("recipes")
    if not isinstance(recipes, list) or len(recipes) == 0:
        errors.append("`recipes` must be a non-empty array")
        return False, errors, None

    for idx, recipe in enumerate(recipes):
        prefix = f"recipes[{idx}]"
        if not isinstance(recipe, dict):
            errors.append(f"{prefix} must be object")
            continue
        if set(recipe.keys()) != RECIPE_KEYS:
            errors.append(f"{prefix} must contain exactly contract fields")

        if not _is_non_empty_string(recipe.get("name")):
            errors.append(f"{prefix}.name invalid")
        if not _is_non_empty_string(recipe.get("description")):
            errors.append(f"{prefix}.description invalid")

        ingredients = recipe.get("ingredients")
        if not isinstance(ingredients, list) or len(ingredients) == 0:
            errors.append(f"{prefix}.ingredients must be non-empty array")
        else:
            for j, ing in enumerate(ingredients):
                ip = f"{prefix}.ingredients[{j}]"
                if not isinstance(ing, dict):
                    errors.append(f"{ip} must be object")
                    continue
                if set(ing.keys()) != INGREDIENT_KEYS:
                    errors.append(f"{ip} must contain exactly name/amount/optional")
                if not _is_non_empty_string(ing.get("name")):
                    errors.append(f"{ip}.name invalid")
                if not _is_non_empty_string(ing.get("amount")):
                    errors.append(f"{ip}.amount invalid")
                if not isinstance(ing.get("optional"), bool):
                    errors.append(f"{ip}.optional must be boolean")

        steps = recipe.get("steps")
        if not isinstance(steps, list) or len(steps) == 0 or not all(
            _is_non_empty_string(s) for s in steps
        ):
            errors.append(f"{prefix}.steps must be non-empty array of strings")

        prep_minutes = recipe.get("prepMinutes")
        cook_minutes = recipe.get("cookMinutes")
        servings = recipe.get("servings")
        if not isinstance(prep_minutes, int) or prep_minutes < 0:
            errors.append(f"{prefix}.prepMinutes must be integer >= 0")
        if not isinstance(cook_minutes, int) or cook_minutes < 0:
            errors.append(f"{prefix}.cookMinutes must be integer >= 0")
        if not isinstance(servings, int) or servings <= 0:
            errors.append(f"{prefix}.servings must be integer > 0")

        difficulty = str(recipe.get("difficulty", "")).strip().lower()
        if difficulty not in DIFFICULTIES:
            errors.append(f"{prefix}.difficulty must be one of easy|medium|hard")

        if not _is_non_empty_string(recipe.get("cuisine")):
            errors.append(f"{prefix}.cuisine invalid")

        dietary = recipe.get("dietaryTags")
        if not isinstance(dietary, list) or not all(isinstance(x, str) for x in dietary):
            errors.append(f"{prefix}.dietaryTags must be array of strings")

        safety_notes = recipe.get("safetyNotes")
        if not isinstance(safety_notes, list) or not all(
            isinstance(x, str) for x in safety_notes
        ):
            errors.append(f"{prefix}.safetyNotes must be array of strings")

    return len(errors) == 0, errors, obj


def validate_safety(parsed: Dict[str, Any]) -> Tuple[bool, List[str]]:
    errors: List[str] = []
    recipes = parsed.get("recipes", [])
    for i, recipe in enumerate(recipes):
        prefix = f"recipes[{i}]"
        ingredient_names = [
            str(ing.get("name", "")).lower() for ing in recipe.get("ingredients", [])
        ]
        notes = [str(x) for x in recipe.get("safetyNotes", [])]
        steps = [str(x) for x in recipe.get("steps", [])]
        text_blob = "\n".join(notes + steps).lower()

        for phrase in BANNED_SAFETY_PHRASES:
            if phrase in text_blob:
                errors.append(f"{prefix} contains banned phrase: {phrase}")

        has_chicken = any("chicken" in x for x in ingredient_names)
        if has_chicken and not SAFE_CHICKEN_TEMP_PATTERN.search(text_blob):
            errors.append(f"{prefix} includes chicken without explicit safe temperature guidance")

        has_kidney_bean = any("kidney bean" in x for x in ingredient_names)
        if has_kidney_bean and "boil" not in text_blob:
            errors.append(f"{prefix} includes kidney beans without explicit boil guidance")

        allergens_present = any(
            any(allergen in ing for allergen in COMMON_ALLERGENS)
            for ing in ingredient_names
        )
        if allergens_present and len(notes) == 0:
            errors.append(f"{prefix} includes common allergens without safety notes")

    return len(errors) == 0, errors


def build_safety_repair_prompt(
    case: EvalCase,
    previous_content: str,
    schema_errors: List[str],
    safety_errors: List[str],
) -> str:
    filters = json.dumps(case.filters, separators=(",", ":")) if case.filters else "none"
    schema_error_text = "; ".join(schema_errors) if schema_errors else "none"
    safety_error_text = "; ".join(safety_errors) if safety_errors else "none"
    return (
        "Rewrite the previous model output so it is valid JSON and passes strict schema + food safety checks.\n"
        f"Normalized ingredients: {', '.join(case.normalized_ingredients)}\n"
        f"Search mode: {case.mode}\n"
        f"Filters: {filters}\n"
        f"Schema errors from previous output: {schema_error_text}\n"
        f"Safety errors from previous output: {safety_error_text}\n"
        "Hard requirements:\n"
        "- Output JSON only (no markdown)\n"
        "- Top level must be exactly {\"recipes\": [...]}\n"
        "- Each ingredient object must contain exactly: name, amount, optional\n"
        "- If chicken appears, include 74C/165F guidance\n"
        "- If kidney beans appear, include boil guidance before simmering\n"
        "- Include allergy notes when common allergens appear\n"
        "- Keep recipes practical\n"
        "Previous output to repair:\n"
        f"{previous_content}"
    )


def compute_weighted_score(summary: Dict[str, Any]) -> float:
    quality = (
        0.5 * float(summary["schema_pass_rate_percent"])
        + 0.3 * float(summary["safety_pass_rate_percent"])
        + 0.2 * float(summary["complex_pass_rate_percent"])
    )
    latency_penalty = min(float(summary["p95_latency_ms"]) / 1000.0, 120.0) * 0.2
    return quality - latency_penalty


def percentile(values: List[float], p: float) -> float:
    if not values:
        return 0.0
    if len(values) == 1:
        return values[0]
    k = (len(values) - 1) * p
    f = math.floor(k)
    c = math.ceil(k)
    if f == c:
        return sorted(values)[int(k)]
    sorted_vals = sorted(values)
    d0 = sorted_vals[f] * (c - k)
    d1 = sorted_vals[c] * (k - f)
    return d0 + d1


def write_jsonl(path: Path, rows: Iterable[Dict[str, Any]]) -> None:
    with path.open("w", encoding="utf-8") as f:
        for row in rows:
            f.write(json.dumps(row, ensure_ascii=False) + "\n")


def build_scorecard(summary: Dict[str, Any], failures: Dict[str, List[str]]) -> str:
    lines = []
    lines.append("# LLM Eval Scorecard")
    lines.append("")
    lines.append("## Summary")
    lines.append("")
    lines.append(f"- Model: `{summary['model']}`")
    if summary.get("prompt_profile"):
        lines.append(f"- Prompt profile: `{summary['prompt_profile']}`")
    lines.append(f"- Base URL: `{summary['base_url']}`")
    lines.append(f"- Recipe cases: `{summary['recipe_total']}`")
    lines.append(f"- Safety cases: `{summary['safety_total']}`")
    lines.append(f"- Complex cases: `{summary['complex_total']}`")
    lines.append(f"- Schema pass rate: `{summary['schema_pass_rate_percent']:.2f}%`")
    lines.append(f"- Safety pass rate: `{summary['safety_pass_rate_percent']:.2f}%`")
    lines.append(f"- Complex pass rate: `{summary['complex_pass_rate_percent']:.2f}%`")
    lines.append(f"- Avg latency: `{summary['avg_latency_ms']:.2f} ms`")
    lines.append(f"- P95 latency: `{summary['p95_latency_ms']:.2f} ms`")
    if "weighted_score" in summary:
        lines.append(f"- Weighted score: `{summary['weighted_score']:.2f}`")
    lines.append("")
    lines.append("## Failing Cases")
    lines.append("")
    lines.append(f"- Schema failures: {', '.join(failures['schema']) if failures['schema'] else 'none'}")
    lines.append(f"- Safety failures: {', '.join(failures['safety']) if failures['safety'] else 'none'}")
    lines.append(f"- Complex failures: {', '.join(failures['complex']) if failures['complex'] else 'none'}")
    lines.append(f"- Request failures: {', '.join(failures['request']) if failures['request'] else 'none'}")
    lines.append("")
    lines.append("## Gate Check")
    lines.append("")
    lines.append(
        f"- Schema >=95%: {'pass' if summary['schema_pass_rate_percent'] >= 95.0 else 'fail'}"
    )
    lines.append(
        f"- Safety >=99%: {'pass' if summary['safety_pass_rate_percent'] >= 99.0 else 'fail'}"
    )
    lines.append(
        f"- Complex >=70%: {'pass' if summary['complex_pass_rate_percent'] >= 70.0 else 'fail'}"
    )
    return "\n".join(lines) + "\n"


def run_profile_eval(
    args: argparse.Namespace,
    out_dir: Path,
    profile_name: str,
    recipe_cases: List[EvalCase],
    safety_cases: List[EvalCase],
    complex_cases: List[EvalCase],
) -> Dict[str, Any]:
    all_cases = recipe_cases + safety_cases + complex_cases
    out_dir.mkdir(parents=True, exist_ok=True)

    details: List[Dict[str, Any]] = []
    latencies: List[float] = []

    schema_pass = 0
    schema_total = 0
    safety_pass = 0
    safety_total = len(safety_cases)
    complex_pass = 0
    complex_total = len(complex_cases)

    failures = {"schema": [], "safety": [], "complex": [], "request": []}

    print(
        f"Running eval for model={args.model} profile={profile_name} cases={len(all_cases)} timeout={args.timeout_seconds}s",
        flush=True,
    )

    for idx, case in enumerate(all_cases, start=1):
        print(f"[{idx}/{len(all_cases)}] {case.case_type}:{case.case_id}", flush=True)
        prompt = build_prompt(
            case,
            disable_thinking_tag=args.disable_thinking_tag,
            prompt_profile=profile_name,
        )
        content, http_status, latency_ms, req_error = call_model(
            base_url=args.base_url,
            api_key=args.api_key,
            model=args.model,
            prompt=prompt,
            temperature=args.temperature,
            timeout_seconds=args.timeout_seconds,
        )
        latencies.append(latency_ms)

        row: Dict[str, Any] = {
            "caseId": case.case_id,
            "type": case.case_type,
            "promptProfile": profile_name,
            "httpStatus": http_status,
            "latencyMs": round(latency_ms, 2),
            "requestError": req_error,
            "schemaValid": False,
            "schemaErrors": [],
            "safetyValid": None,
            "safetyErrors": [],
            "complexValid": None,
            "complexErrors": [],
            "safetyRepairAttempted": False,
            "safetyRepairSucceeded": False,
            "safetyRepairLatencyMs": None,
        }

        if req_error or content is None:
            failures["request"].append(case.case_id)
            if case.case_type == "safety":
                failures["safety"].append(case.case_id)
            if case.case_type == "complex":
                failures["complex"].append(case.case_id)
            details.append(row)
            print(
                f"  -> request_error latency={row['latencyMs']}ms error={req_error}",
                flush=True,
            )
            continue

        valid_schema, schema_errors, parsed = validate_schema(content)
        final_schema = valid_schema
        final_schema_errors = schema_errors

        final_safety: Optional[bool] = None
        final_safety_errors: List[str] = []

        if case.case_type == "safety":
            if valid_schema and parsed is not None:
                safety_valid, safety_errors = validate_safety(parsed)
            else:
                safety_valid, safety_errors = False, ["schema invalid; safety skipped"]
            final_safety = safety_valid
            final_safety_errors = safety_errors

            should_repair = args.enable_safety_repair and (not final_schema or not final_safety)
            if should_repair:
                row["safetyRepairAttempted"] = True
                repair_prompt = build_safety_repair_prompt(
                    case,
                    previous_content=content,
                    schema_errors=final_schema_errors,
                    safety_errors=final_safety_errors,
                )
                repaired_content, _, repair_latency_ms, repair_error = call_model(
                    base_url=args.base_url,
                    api_key=args.api_key,
                    model=args.model,
                    prompt=repair_prompt,
                    temperature=0.0,
                    timeout_seconds=args.repair_timeout_seconds,
                )
                row["safetyRepairLatencyMs"] = round(repair_latency_ms, 2)
                if repair_error is None and repaired_content is not None:
                    repaired_schema, repaired_schema_errors, repaired_parsed = validate_schema(repaired_content)
                    if repaired_schema and repaired_parsed is not None:
                        repaired_safety, repaired_safety_errors = validate_safety(repaired_parsed)
                    else:
                        repaired_safety, repaired_safety_errors = False, ["schema invalid after repair"]

                    if repaired_schema and repaired_safety:
                        final_schema = repaired_schema
                        final_schema_errors = repaired_schema_errors
                        final_safety = repaired_safety
                        final_safety_errors = repaired_safety_errors
                        row["safetyRepairSucceeded"] = True

            row["safetyValid"] = final_safety
            row["safetyErrors"] = final_safety_errors

        row["schemaValid"] = final_schema
        row["schemaErrors"] = final_schema_errors
        schema_total += 1
        if final_schema:
            schema_pass += 1
        else:
            failures["schema"].append(case.case_id)

        if case.case_type == "safety":
            if final_safety:
                safety_pass += 1
            else:
                failures["safety"].append(case.case_id)

        if case.case_type == "complex":
            if final_schema and parsed is not None:
                valid_complex, complex_errors = evaluate_complexity(case, parsed)
            else:
                valid_complex, complex_errors = False, ["schema invalid; complexity skipped"]
            row["complexValid"] = valid_complex
            row["complexErrors"] = complex_errors
            if valid_complex:
                complex_pass += 1
            else:
                failures["complex"].append(case.case_id)

        details.append(row)
        print(
            f"  -> schema={row['schemaValid']} safety={row['safetyValid']} complex={row['complexValid']} repair={row['safetyRepairSucceeded']} latency={row['latencyMs']}ms",
            flush=True,
        )

    summary = {
        "model": args.model,
        "prompt_profile": profile_name,
        "base_url": args.base_url,
        "recipe_total": len(recipe_cases),
        "safety_total": len(safety_cases),
        "complex_total": len(complex_cases),
        "schema_total": schema_total,
        "schema_pass": schema_pass,
        "schema_pass_rate_percent": (schema_pass / schema_total * 100.0) if schema_total else 0.0,
        "safety_pass": safety_pass,
        "safety_pass_rate_percent": (safety_pass / safety_total * 100.0) if safety_total else 0.0,
        "complex_pass": complex_pass,
        "complex_pass_rate_percent": (complex_pass / complex_total * 100.0) if complex_total else 0.0,
        "avg_latency_ms": statistics.mean(latencies) if latencies else 0.0,
        "p95_latency_ms": percentile(latencies, 0.95) if latencies else 0.0,
        "safety_repair_attempted": sum(1 for row in details if row.get("safetyRepairAttempted")),
        "safety_repair_succeeded": sum(1 for row in details if row.get("safetyRepairSucceeded")),
        "generated_at_unix": int(time.time()),
    }
    summary["weighted_score"] = compute_weighted_score(summary)

    (out_dir / "summary.json").write_text(json.dumps(summary, indent=2), encoding="utf-8")
    write_jsonl(out_dir / "details.jsonl", details)
    (out_dir / "scorecard.md").write_text(build_scorecard(summary, failures), encoding="utf-8")

    print(f"Wrote eval results to: {out_dir}")
    print(json.dumps(summary, indent=2))
    return summary


def main() -> int:
    args = parse_args()
    recipe_cases = read_jsonl(args.recipe_cases, "recipe")
    safety_cases = read_jsonl(args.safety_cases, "safety")
    complex_cases: List[EvalCase] = []
    if args.complex_cases:
        complex_cases = read_jsonl(args.complex_cases, "complex")

    base_out_dir = Path(args.out)
    profiles = ["schema_first", "safety_complex_first"] if args.compare_profiles else [args.prompt_profile]

    profile_summaries: List[Dict[str, Any]] = []
    for profile_name in profiles:
        out_dir = base_out_dir / profile_name if args.compare_profiles else base_out_dir
        summary = run_profile_eval(
            args=args,
            out_dir=out_dir,
            profile_name=profile_name,
            recipe_cases=recipe_cases,
            safety_cases=safety_cases,
            complex_cases=complex_cases,
        )
        profile_summaries.append(summary)

    if args.compare_profiles:
        sorted_profiles = sorted(
            profile_summaries,
            key=lambda x: float(x["weighted_score"]),
            reverse=True,
        )
        compare_payload = {
            "model": args.model,
            "profiles": sorted_profiles,
            "winner": sorted_profiles[0]["prompt_profile"] if sorted_profiles else None,
            "generated_at_unix": int(time.time()),
        }
        (base_out_dir / "profiles_summary.json").write_text(
            json.dumps(compare_payload, indent=2),
            encoding="utf-8",
        )
        print("Wrote profile comparison to:", base_out_dir / "profiles_summary.json")
        print(json.dumps(compare_payload, indent=2))

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
