# LLM Fine-Tune Candidate Shortlist (Recipes)

Purpose: choose one base model family to fine-tune for recipe generation quality while keeping inference practical.

## Recommendation

Start with a two-track bakeoff:

- Track A (recommended default): `Qwen3-8B`
- Track B: `Qwen3-4B` (efficiency/control baseline)

Why this pairing:

- Same family means simpler prompt/template compatibility and migration.
- `Qwen3-8B` is the best current quality candidate for this stack.
- `Qwen3-4B` is useful as a smaller-cost control, even though current local runs show timeout issues on long prompts.

## Candidate Matrix

| Candidate | Role | Strengths | Risks | Fit for Recipes |
|---|---|---|---|---|
| `Qwen3-8B` | Primary candidate | Best current local behavior for target prompts, strong structured output potential | Heavier than 4B for fine-tune/inference | High |
| `Qwen3-4B` | Efficiency baseline | Lower footprint and faster iteration for QLoRA experiments | Current local inference timed out on full eval prompts | Medium |
| `Llama-3.1-8B-Instruct` | Stretch candidate | Strong general quality | License and usage constraints must be reviewed carefully | Medium-High |

## Decision Gates

Choose a winner only if all are true:

- Schema-valid JSON rate >= 95% on recipe eval set
- Safety pass rate >= 99% on cooking-risk prompts
- Cost/latency stays within target budget
- License/compliance approved for intended use

## Evaluation Plan

1. Run zero-shot baseline on all candidates using current prompt contract.
2. Run lightweight prompt tuning pass (no fine-tune) and re-score.
3. Fine-tune top 2 candidates with the same training split.
4. Compare on held-out eval set and choose winner.

## Fine-Tune Strategy (Initial)

- Method: QLoRA (adapter-based) before any full fine-tune.
- Training objective: improve recipe structure adherence + culinary coherence.
- Keep provenance metadata for each training run (dataset hash, prompt version, hyperparameters).

## Data Needed Before Fine-Tuning

- Licensed/public-domain recipe corpus with attribution metadata.
- Structured examples matching app output schema.
- Safety-focused adversarial prompts (food safety, allergens, undercooking risks).

## Output of This Phase

- Final selected base model ID
- Eval scorecard and comparison notes
- License/compliance sign-off note
- Rollout plan (canary + rollback)
