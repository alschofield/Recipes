# API Gateway (Single Domain)

This workspace holds lightweight gateway options so the web app can use a single API origin
(`https://api.yourdomain.com`) while backend services stay separately deployed.

Current cheapest-first option included:

- `cloudflare-worker/` - path-based routing Worker (free-tier friendly for low traffic)

Why this helps:

- Web only needs one public API URL
- Keeps recipes/users/favorites services independently deployable
- Avoids running an additional paid gateway container in many cases
