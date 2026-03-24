#!/usr/bin/env python3
"""Create a filled QLoRA training record from eval artifacts.

This does not train; it assembles run metadata and eval comparison.
"""

from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any, Dict


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Create filled QLoRA training record")
    parser.add_argument("--template", required=True)
    parser.add_argument("--run-id", required=True)
    parser.add_argument("--date", required=True)
    parser.add_argument("--dataset-name", required=True)
    parser.add_argument("--dataset-version", required=True)
    parser.add_argument("--train-count", type=int, required=True)
    parser.add_argument("--val-count", type=int, required=True)
    parser.add_argument("--test-count", type=int, required=True)
    parser.add_argument("--dataset-hash", required=True)
    parser.add_argument("--adapter-path", required=True)
    parser.add_argument("--logs-path", required=True)
    parser.add_argument("--eval-output-path", required=True)
    parser.add_argument("--eval-summary", required=True, help="Path to fine-tuned summary.json")
    parser.add_argument(
        "--baseline-profile-summary",
        required=True,
        help="Path to baseline profiles_summary.json",
    )
    parser.add_argument(
        "--baseline-profile",
        default="safety_complex_first",
        help="Baseline prompt profile to compare against",
    )
    parser.add_argument("--timeout-rate-percent", type=float, default=0.0)
    parser.add_argument("--decision-outcome", default="iterate", choices=["promote", "iterate", "rollback"])
    parser.add_argument("--decision-notes", default="")
    parser.add_argument("--out", required=True)
    return parser.parse_args()


def load_json(path: Path) -> Dict[str, Any]:
    with path.open("r", encoding="utf-8") as f:
        data = json.load(f)
    if not isinstance(data, dict):
        raise ValueError(f"expected JSON object at {path}")
    return data


def pick_baseline(summary: Dict[str, Any], profile: str) -> Dict[str, Any]:
    profiles = summary.get("profiles", [])
    if not isinstance(profiles, list):
        raise ValueError("baseline profiles_summary.json missing 'profiles' list")

    for item in profiles:
        if isinstance(item, dict) and item.get("prompt_profile") == profile:
            return item

    raise ValueError(f"baseline profile '{profile}' not found")


def pct(value: Any) -> float:
    try:
        return float(value)
    except Exception:
        return 0.0


def main() -> int:
    args = parse_args()

    template = load_json(Path(args.template))
    eval_summary = load_json(Path(args.eval_summary))
    baseline_summary = load_json(Path(args.baseline_profile_summary))
    baseline = pick_baseline(baseline_summary, args.baseline_profile)

    out = dict(template)
    out["runId"] = args.run_id
    out["date"] = args.date

    dataset = dict(out.get("dataset", {}))
    dataset.update(
        {
            "name": args.dataset_name,
            "version": args.dataset_version,
            "trainCount": args.train_count,
            "valCount": args.val_count,
            "testCount": args.test_count,
            "datasetHash": args.dataset_hash,
        }
    )
    out["dataset"] = dataset

    artifacts = dict(out.get("artifacts", {}))
    artifacts.update(
        {
            "adapterPath": args.adapter_path,
            "logsPath": args.logs_path,
            "evalOutputPath": args.eval_output_path,
        }
    )
    out["artifacts"] = artifacts

    metrics = dict(out.get("metrics", {}))
    metrics.update(
        {
            "schemaPassRatePercent": pct(eval_summary.get("schema_pass_rate_percent")),
            "safetyPassRatePercent": pct(eval_summary.get("safety_pass_rate_percent")),
            "complexPassRatePercent": pct(eval_summary.get("complex_pass_rate_percent")),
            "p95LatencyMs": pct(eval_summary.get("p95_latency_ms")),
            "timeoutRatePercent": args.timeout_rate_percent,
        }
    )
    out["metrics"] = metrics

    baseline_compare = {
        "profile": args.baseline_profile,
        "schemaPassDelta": metrics["schemaPassRatePercent"] - pct(baseline.get("schema_pass_rate_percent")),
        "safetyPassDelta": metrics["safetyPassRatePercent"] - pct(baseline.get("safety_pass_rate_percent")),
        "complexPassDelta": metrics["complexPassRatePercent"] - pct(baseline.get("complex_pass_rate_percent")),
        "p95LatencyDeltaMs": metrics["p95LatencyMs"] - pct(baseline.get("p95_latency_ms")),
    }
    out["baselineComparison"] = baseline_compare

    decision = dict(out.get("decision", {}))
    decision.update({"outcome": args.decision_outcome, "notes": args.decision_notes})
    out["decision"] = decision

    out_path = Path(args.out)
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(out, indent=2), encoding="utf-8")
    print(f"Wrote training record: {out_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
