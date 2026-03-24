# V1 Server/Web Smoke Command Pack

Use these commands against staging or production domains during final launch gating.

Set environment first:

```bash
export WEB_ORIGIN="https://www.yourdomain.com"
export API_ORIGIN="https://api.yourdomain.com"
```

## Health checks

```bash
curl -fsS "$WEB_ORIGIN/api/health"
curl -fsS "$API_ORIGIN/recipes/health"
curl -fsS "$API_ORIGIN/users/health"
curl -fsS "$API_ORIGIN/favorites/health"
```

## Auth/session checks (requires test account)

```bash
curl -sS -X POST "$API_ORIGIN/users/login" -H "content-type: application/json" -d '{"username":"<user>","password":"<pass>"}'
curl -sS -X POST "$API_ORIGIN/users/refresh" -H "content-type: application/json" -d '{"refreshToken":"<refresh-token>"}'
curl -sS "$API_ORIGIN/users/<user-id>/sessions" -H "authorization: Bearer <access-token>"
curl -sS -X POST "$API_ORIGIN/users/logout/session" -H "content-type: application/json" -d '{"refreshToken":"<refresh-token>"}'
```

## Search/detail checks

```bash
curl -sS -X POST "$API_ORIGIN/recipes/search" -H "content-type: application/json" -d '{"ingredients":["chicken","rice","garlic"],"mode":"strict","complex":false,"pagination":{"page":1,"pageSize":5}}'
curl -sS "$API_ORIGIN/recipes/detail/<recipe-id>"
```

## Favorites checks

```bash
curl -sS "$API_ORIGIN/favorites/<user-id>" -H "authorization: Bearer <access-token>"
curl -sS -X POST "$API_ORIGIN/favorites/<user-id>/<recipe-id>" -H "authorization: Bearer <access-token>"
curl -sS -X DELETE "$API_ORIGIN/favorites/<user-id>/<recipe-id>" -H "authorization: Bearer <access-token>"
```

## Web path smoke

```bash
curl -I "$WEB_ORIGIN/"
curl -I "$WEB_ORIGIN/recipes"
curl -I "$WEB_ORIGIN/recipes/<recipe-id>"
curl -I "$WEB_ORIGIN/favorites"
curl -I "$WEB_ORIGIN/account"
```

Record outputs and timestamps in `v1-launch-blocker-evidence.md`.
