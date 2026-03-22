# Operations Runbook

## Reverse Proxy and Routing

- Local routing config: `server/etc/nginx/nginx.local.conf`
- Production routing config: `server/etc/nginx/nginx.production.conf`
- Docker compose picks config via `NGINX_CONF_FILE` (defaults to local):

```bash
NGINX_CONF_FILE=./server/etc/nginx/nginx.production.conf docker compose up -d
```

Route mapping:

- `/api/users/*` -> users service
- `/api/favorites/*` -> favorites service
- `/api/recipes/*` -> recipes service
- `/api/ingredients/*` -> recipes service
- `/` -> web service

Metrics endpoints behind Nginx:

- `/metrics/users`
- `/metrics/favorites`
- `/metrics/recipes`

## Structured Logging and Request IDs

All services emit JSON logs with:

- `requestId`
- `service`
- `method`
- `path`
- `status`
- `durationMs`

The request ID is accepted from `X-Request-ID` or generated server-side and echoed back.

## Error Tracking and Alerting

Set `ERROR_WEBHOOK_URL` in each service environment to enable automatic webhook alerts on `5xx` responses.

Payload includes:

- service
- requestId
- method
- path
- status
- time

## Backup and Restore (Postgres)

Backup:

```bash
server/scripts/db/backup.sh
```

Restore:

```bash
server/scripts/db/restore.sh backups/recipes_YYYYMMDD_HHMMSS.dump
```

Recovery validation checklist:

1. Run restore script against fresh DB.
2. Apply any new migrations if needed.
3. Start services and hit `/users/health`, `/favorites/health`, `/recipes/health`.
4. Smoke test login + search + favorites flows.

## Release Checklist

1. `go test ./... && go build ./...` in `server/`
2. `pnpm run lint && pnpm run test && pnpm run build` in `web/`
3. Verify migrations are applied in staging.
4. Deploy services and Nginx config.
5. Validate metrics endpoints and error webhook delivery.
6. Run smoke journey: signup/login -> search -> favorite -> detail.

## Rollback Instructions

1. Roll back to previous app image/tag for web and API services.
2. Revert Nginx to last known good config.
3. If schema migration caused issue, run migration down or restore latest DB backup.
4. Re-run smoke test and monitor 5xx error rate.
