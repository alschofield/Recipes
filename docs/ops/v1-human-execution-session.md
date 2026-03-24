# V1 Human Execution Session Plan

Use this for a live blocker-clearing session with owner access to infra, stores, and approvals.

## Session Goal

Close as many external/manual gates as possible in one pass and update `v1-launch-blocker-evidence.md` live.

## Before Session (already automated)

- Confirm latest reports are fresh:
  - `docs/ops/v1-local-preflight-latest.md`
  - `docs/ops/v1-runtime-smoke-latest.md`
  - `docs/ops/v1-open-gates-snapshot.md`
  - `docs/ops/v1-gate-dashboard-latest.md`

## Live Session Order

1. **External inputs sweep**
   - Fill `v1-external-inputs-checklist.md`
   - Confirm missing credentials/access owners

2. **Staging verification sweep**
   - Run deployed smoke from `v1-server-web-smoke-command-pack.md`
   - Record results in `v1-launch-blocker-evidence.md`

3. **Web manual sweep**
   - Run `docs/product/v1-web-usability-sanity-sheet.md`
   - Capture any P0/P1 defects and disposition

4. **Mobile manual sweep**
   - Run `docs/mobile/native-accessibility-manual-test-sheet-v1.md`
   - Run `docs/mobile/voice-stt-validation-matrix-v1.md`
   - Run `docs/mobile/v1-mobile-ux-sanity-sheet.md`

5. **Store/compliance sweep**
   - Complete `docs/mobile/store-submission-pack-template-v1.md` (archive dated working copy when finalized)
   - Confirm policy URLs and declarations in store consoles

6. **Final gate review**
   - Refresh `v1-open-gates-snapshot.md`
   - Update final decision block in `v1-launch-blocker-evidence.md`

## What to collect during session

- Workflow run URLs
- Staging/prod deploy URLs
- Lighthouse report URLs
- Store console screenshot links (or internal refs)
- Approver names + timestamps

## Third-Party Escalation Pack

If specialized security/legal expertise is needed, share:

- `../archive/deliverables/security-pre-audit-report-2026-03-23.md`
- `../archive/deliverables/legal-compliance-pre-audit-report-2026-03-23.md`
- `../deliverables/third-party-security-audit-brief-v1.md`
- `../deliverables/third-party-legal-review-brief-v1.md`
- `../deliverables/third-party-audit-options-v1.md`
- `../deliverables/third-party-design-ux-brief-v1.md`
- `../deliverables/third-party-data-curation-brief-v1.md`
