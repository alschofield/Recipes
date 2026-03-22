# Hosting Strategy: Cheapest First, Scalable Later

This doc compares practical hosting layouts for Recipes and recommends a staged path.

## Recommendation summary

## Stage 1 (low cost / launch fast)

- **Web:** Vercel Hobby (free start, custom domain ready)
- **API containers:** Railway (or Render/Fly equivalent low-cost container runtime)
- **Postgres:** managed free/low-cost tier (Neon/Supabase/Railway)
- **Redis:** Upstash or managed Redis low-cost tier

## Stage 2 (custom domain + better reliability)

- Web: `www.yourdomain.com`
- API: `api.yourdomain.com`
- Move to paid plans where needed to remove sleep/cold-start limits

## Stage 3 (growth)

- Scale API services independently
- Add stronger monitoring and SLO alerts
- Add pre-prod environment parity and progressive rollouts

---

## Single-provider vs split-provider tradeoff

Does single API domain cost more?

- **Not necessarily.**
- If you add a dedicated paid gateway service, cost usually increases.
- If you use a lightweight edge router (for example Cloudflare Worker free tier for low traffic), single-domain API can stay near-zero extra cost.
- Cheapest pattern early: split compute/datastores + edge gateway on free tier.

## Option A: same provider for API + DB + Redis

Benefits:

- Simpler operations (one control plane, fewer credentials)
- Potentially lower network latency between app and DB/cache
- Fewer moving parts for initial setup

Tradeoffs:

- Vendor lock-in increases
- Outage blast radius is larger (one provider impacts everything)
- You may pay more for one layer than best-of-breed alternatives

When it makes sense:

- Small team, fast iteration, lowest ops overhead priority

## Option B: split providers (example: Vercel web + Railway API + Neon DB + Upstash Redis)

Benefits:

- Strong price/performance flexibility
- Best tool for each layer
- Better fault isolation across providers

Tradeoffs:

- More setup/configuration complexity
- More secrets and dashboards to manage
- Potential cross-provider network latency considerations

When it makes sense:

- You want low cost now with clean scaling options later

---

## Practical guidance for this repo

Given current architecture:

- Keep web and API deployment independent
- Keep users/recipes/favorites as independently deployable services
- Keep DB migrations as a dedicated release step

If you prefer operational simplicity now, run API + DB + Redis on the same provider.
If you prefer cost optimization now, split DB/Redis onto free tiers and keep API compute minimal.

For this repository, the recommended cheapest Option B implementation is:

- Web on Vercel Hobby
- API services on low-cost container host
- Managed low-cost Postgres + Redis
- Cloudflare Worker gateway for `api.yourdomain.com`
