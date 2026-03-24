# Monetization Strategy v1

Goal: establish a debt-reduction-first revenue plan that can fund infrastructure and generate meaningful owner income.

## 1) ICP and willingness-to-pay segments

### Primary segments

1. Busy professionals
   - Pain: low time, decision fatigue after work.
   - WTP signal: convenience and reliability beats novelty.
2. Family meal planners
   - Pain: budget constraints, repeatable weekly planning.
   - WTP signal: saved plans and shopping utility reduce friction.
3. Fitness-focused users
   - Pain: macro/protein targets + variety.
   - WTP signal: high for precise filters and repeat tracking.
4. Dietary-restricted users
   - Pain: safe recipe discovery and substitution confidence.
   - WTP signal: high when trust and safety notes are explicit.

## 2) Paid feature boundaries

### Free tier

- Limited daily recipe generations.
- Basic search/filter and favorites.
- Standard recipe detail.

### Pro tier

- Higher generation limits and priority fallback handling.
- Advanced filters (dietary complexity, quality/ranking controls).
- Saved meal plans and shopping-list export.
- Recipe history + reusable pantry profiles.

## 3) Pricing hypothesis

- Monthly: `$9.99`
- Annual: `$89.99` (about 25% discount)
- Target gross margin guardrail: `>=70%` after inference + infra.

Assumption: pro users trigger higher fallback usage; maintain margin by improving DB recall and fallback canary controls.

## 4) Billing infrastructure decision

Decision: Stripe subscriptions (hosted checkout + customer portal).

Implementation plan:

1. Stripe product + price IDs for monthly/annual plans.
2. Webhook handling for subscription state, trial end, and payment failure.
3. Entitlement table in server DB (`plan`, `status`, `currentPeriodEnd`, `cancelAtPeriodEnd`).
4. Grace + dunning policy for failed payments.

## 5) Activation funnel metrics

Track funnel with weekly reporting:

- visit -> signup
- signup -> first recipe search
- first search -> first successful recipe detail view
- first detail -> first favorite save
- favorite save -> paywall exposure
- paywall exposure -> paid conversion

Core KPIs:

- activation rate (signup to first successful recipe)
- 7-day retention
- trial-to-paid conversion
- paid churn

## 6) Paywall experiment plan

Initial experiments:

1. Timing: paywall after N successful generations vs after first favorite save.
2. Messaging: convenience-first vs safety/trust-first copy.
3. Limits: daily generation limit variants.
4. CTA framing: monthly default vs annual default.

Success criteria:

- conversion uplift with no severe retention drop
- stable search success and latency under increased pro usage

## 7) Native ad monetization investigation

Policy anchors (see `ad-monetization-policy.md`):

- sponsored items must be clearly disclosed
- ad placements cannot obscure safety-critical recipe information
- no dark patterns that mimic organic results

Evaluation metrics:

- RPM uplift
- conversion impact on pro subscriptions
- retention impact on active users

## 8) Retention loop plan

- Weekly meal-plan reminder cadence.
- Pantry refresh nudges tied to prior successful recipes.
- Streak-style progress for saved/favorited recipes.
- Reactivation campaigns for dormant users (email/push).

## 9) Unit economics dashboard spec

Weekly dashboard fields:

- active users
- conversion rate
- churn rate
- ARPU
- LTV proxy
- inference cost per active user
- blended infra cost per active user
- contribution margin

Threshold alerts:

- inference cost/user rises >20% week-over-week
- conversion drops >15% week-over-week
- churn rises above target band
