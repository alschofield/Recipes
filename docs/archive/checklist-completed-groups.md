# Checklist Completed Groups Archive

Archived on 2026-03-23.

## Source: `llm/CHECKLIST.md`

### Define Constraints First (completed)

- Confirmed target use (inference now, QLoRA next), deployment target (self-hosted local Docker), budget ceilings, and acceptance gates.

### Workspace Standardization (completed)

- Compartmentalized LLM workspace into `evals/`, `train/`, and `judge/` lanes.
- Added dedicated trainer compose stack at `llm/train/docker-compose.train.yml`.
- Added judge output schemas for ingredient metadata and recipe quality.

### GitHub/Release Hygiene (completed)

- Consolidated ignore policy and release checks.
- Added production gate doc `llm/PRODUCTION-READINESS.md`.
- Ran pre-push readiness checks (`py_compile`, `go test ./pkg/search`, `docker compose config`).

### Model Validation Before Promotion (completed)

- Verified license/commercial posture and context details for `qwen3:8b` + fallback candidate.
- Verified JSON reliability behavior on latest eval artifacts.
- Added latency/cost demand-assumption validation snapshot.
- Added safety regression notes from latest risk-suite outcomes.
- Documented decision to keep hosted OpenAI-compatible DR provider lane.

### Judge Model and Metadata Enrichment (V1+) (completed)

- Defined lightweight judge candidate list for metadata and quality lanes.
- Added calibration dataset + acceptance thresholds.
- Added deterministic-baseline calibration tests and manual spot-check guardrails.

## Source: `server/CHECKLIST.md`

### Mobile API Readiness (completed)

- Implemented rotating refresh-token sessions with session and family revoke controls.
- Added offline-safe favorites mutation semantics and reconciliation contract.
- Added mobile-focused test coverage for refresh/retry/conflict behavior.

## Source: `web/CHECKLIST.md`

### Web UX Modernization and Trust (completed)

- Implemented strict/inclusive/complex clarity and blend trust signals.
- Added modernized recipes UI shell/components with density and pantry ergonomics.
- Added recipes-surface accessibility and reliability hardening.

## Source: `mobile/CHECKLIST.md`

### Mobile Planning and Release Foundations (completed)

- Locked native-platform direction (SwiftUI + Compose).
- Defined product scope, UX/navigation, design parity, and API/data policies.
- Defined release-engineering, store-readiness, and QA/operations baselines.
