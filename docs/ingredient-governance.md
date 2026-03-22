# Ingredient Governance and Community Curation Plan

Status: proposed execution plan (next implementation track)

## Why this matters

Ingredient identity quality drives search relevance, deduplication, ranking confidence, and long-term maintainability.
LLM generation and user suggestions must improve coverage without fragmenting canonical ingredient data.

## Data Model (authoritative)

- `ingredients`: canonical ingredient entities only (`canonical_name` unique).
- `ingredient_aliases`: accepted variants that map to canonical ingredients (`alias` unique).
- `ingredient_candidates` (new): unresolved ingredient proposals from users/LLM.
- `ingredient_candidate_votes` (optional phase 2): user votes on candidate mapping/new canonical decision.

## Candidate Workflow

1. Incoming name is normalized (trim, lowercase, singularization).
2. Deterministic lookup checks:
   - exact alias match
   - exact canonical match
   - normalized equivalent match
3. If confidently matched, reuse canonical `ingredient_id`.
4. If unmatched/low confidence, create `ingredient_candidate` (do not auto-create canonical by default).
5. Admin/mod resolves candidate as:
   - map to existing canonical (alias addition),
   - promote to new canonical,
   - reject/noise.

## LLM Integration Rules

- LLM output must pass schema validation (already implemented).
- For each generated ingredient:
  - high-confidence match => attach existing canonical ingredient;
  - low-confidence/unmatched => create candidate entry.
- Avoid canonical auto-creation from raw LLM output unless deterministic exact match path is satisfied.

## User Contribution UX

- Add user-facing "Suggest Ingredient" flow:
  - input ingredient string,
  - show likely existing canonical/aliases,
  - user chooses "same as existing" or "new ingredient".
- Add review feed for community input (optional phase 2):
  - vote/signal likely mapping,
  - surface confidence to admins.
- Add admin moderation page/API:
  - approve alias mapping,
  - create canonical,
  - reject duplicate/spam.

## Anti-duplicate Safeguards

- Keep unique constraints on `ingredients.canonical_name` and `ingredient_aliases.alias`.
- Introduce normalized key checks before inserts.
- Run similarity checks for near-matches and force candidate queue for ambiguous names.
- Add periodic hygiene job to detect likely duplicates and stale candidates.

## Metrics / Quality Gates

- Canonicalization hit rate (% ingredient references mapped without candidate creation).
- Pending candidate backlog size + age.
- Duplicate merge rate over time.
- Search recall improvement after candidate resolutions.

## Suggested implementation order

1. Add DB tables for candidates (+ indexes/status fields).
2. Implement matcher service with deterministic + low-risk fuzzy checks.
3. Integrate matcher into LLM recipe persistence and seed ingestion.
4. Add user suggestion endpoint + page.
5. Add admin resolution endpoints + basic UI.
6. Add tests and monitoring metrics.
