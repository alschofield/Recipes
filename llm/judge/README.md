# Judge Lane

This lane covers lightweight judge-model quality controls and drift monitoring.

## Drift Check

Use `check_drift.py` to compare current judge-related snapshot data against baseline priors:

```bash
python llm/judge/check_drift.py \
  --snapshot llm/evals/results/nightly-123/judge-snapshot.json \
  --baseline llm/judge/data-priors.summary.json \
  --out-json llm/evals/results/nightly-123/judge-drift.json \
  --out-md llm/evals/results/nightly-123/judge-drift.md
```

Supported snapshot shapes:

- JSON object with `items` list (for example, ingredient catalog payload)
- JSON object with `records` list
- JSON list of records

Drift checks include:

- category distribution L1 distance vs baseline categories
- low-confidence share (default threshold `< 0.65`)
- confidence p50 delta vs baseline p50 proxy

Exit code is non-zero when thresholds are breached.

## Calibration assets

- `calibration-dataset-v1.json` - seed manual/automated judge calibration cases.
- `calibration-thresholds-v1.json` - acceptance thresholds and decision policy.
- `calibration-template.json` - run record template for calibration outcomes.
