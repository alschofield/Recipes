# Open Items Execution Plan (2026-03-23)

This plan covers the remaining active items in `llm/CHECKLIST.md` with execution order and blockers.

## Execution Order

1. **Legal/compliance closeout (dataset lanes)**
   - Inputs ready:
     - `llm/train/datasets/provenance-manifest.v1.json`
     - `llm/train/datasets/source-denylist.txt`
     - `docs/archive/llm/model-validation-report-2026-03-23.md`
   - Remaining action: legal approver decision for unresolved third-party lane entries.
   - Blocker type: external approval.

2. **Staging rollback validation**
   - Run staged provider/model switch using env-only changes (`LLM_MODEL`, `LLM_BASE_URL`, `LLM_API_KEY`, `LLM_FALLBACK_CANARY_PERCENT`, `LLM_FALLBACK_DISABLED`).
   - Validate health + metrics endpoints and rollback timing.
   - Required evidence: command log + metrics snapshot + signoff note in `llm/PRODUCTION-READINESS.md`.
   - Blocker type: staging environment access.

3. **Exit-gate verification**
   - Gate automation is now available:
     - `llm/evals/check_readiness_gates.py`
     - nightly workflow `.github/workflows/nightly-llm-evals.yml`
   - Current known status from latest baseline artifacts:
     - schema gate: fail
     - safety gate: pass
     - complex gate: pass
     - p95 latency gate: fail
   - Blocker type: model/data improvements to pass gates.

4. **First QLoRA pilot run record**
   - Record helper is ready: `llm/train/qlora/create_training_record.py`.
   - Remaining action: produce first real pilot eval output path, then generate run record with baseline deltas.
   - Blocker type: pilot training/eval execution.

## Immediate Work That Is Already Unblocked

- Keep nightly eval + drift + readiness workflows active.
- Keep provenance manifest current for any new dataset lane.
- Prepare pilot dataset in `llm/train/datasets/processed/` and run dedup/safety scripts before training.

## Definition of “Safe to Move to New Checklist”

Treat this checklist as complete only after:

- All remaining active items in `llm/CHECKLIST.md` are checked.
- External signoffs (legal + staging rollback) are attached in docs.
- Exit gates are passing on current production candidate profile.
