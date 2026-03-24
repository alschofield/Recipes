# Recipes Monorepo

Recipes is a multi-service app for ingredient-based recipe discovery, account management, and favorites.
This repository is organized as a monorepo so backend services, web app, docs, and operations assets can evolve together.

## Table of contents

- [Monorepo map](#monorepo-map)
- [Prerequisites](#prerequisites)
- [Global tooling setup](#global-tooling-setup)
- [Quick start](#quick-start)
- [Root command runners](#root-command-runners)
- [Environment variables](#environment-variables)
- [CI/CD at a glance](#cicd-at-a-glance)
- [Public GitHub metadata](#public-github-metadata)
- [GitHub collaboration files](#github-collaboration-files)
- [Recommended workflow](#recommended-workflow)
- [Deep links](#deep-links)

## Monorepo map

| Path | Purpose | Readme |
|---|---|---|
| `server/` | Go APIs, migrations, seed utilities, service Dockerfiles, nginx config, and API gateway config | [`server/README.md`](server/README.md) |
| `web/` | Next.js 16 App Router frontend (SSR-first) | [`web/README.md`](web/README.md) |
| `docs/` | Architecture, contracts, and runbooks | [`docs/README.md`](docs/README.md) |
| `scripts/` | Cross-platform local bootstrap helpers | [`scripts/README.md`](scripts/README.md) |
| `mobile/` | Native mobile workspace (SwiftUI + Jetpack Compose lanes) | [`mobile/README.md`](mobile/README.md) |
| `llm/` | Reserved LLM tooling/evaluation workspace | [`llm/README.md`](llm/README.md) |
| `.github/` | CI workflow, issue/PR templates, ownership, dependency automation | See "GitHub collaboration files" below |

Other root files:

- `docker-compose.yml` - local stack orchestration
- `Makefile` - root command runner for make users
- `Taskfile.yml` - root command runner for Task users
- `.gitignore` - monorepo-wide ignore rules for generated assets/secrets

---

## Prerequisites

- Docker Desktop (or Docker Engine + Compose)
- Go 1.25+
- Node.js 20+
- pnpm 10+

Optional:

- `make` (GNU Make) for `Makefile` commands
- `task` (go-task) for `Taskfile.yml` commands

## Global tooling setup

This repo expects a few globally available tools for local operations:

- Docker CLI + Docker Compose plugin (`docker`, `docker compose`)
- Go CLI (`go`)
- pnpm (`pnpm`) via Corepack
- Task runner (`task`) recommended for cross-platform command execution
- GNU Make (`make`) optional alternative runner

Quick setup references:

- Run bootstrap helper:

```bash
bash scripts/bootstrap.sh
```

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\bootstrap.ps1
```

- Enable pnpm via Corepack:

```bash
corepack enable
corepack prepare pnpm@10 --activate
```

- Verify Docker tooling:

```bash
docker --version
docker compose version
```

Bootstrap scripts:

```bash
bash scripts/bootstrap.sh
```

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\bootstrap.ps1
```

---

## Quick start

### Option A: Make

```bash
make setup
```

Then in separate terminals:

```bash
make server-run-recipes
make server-run-users
make server-run-favorites
make web-dev
```

### Option B: Task

```bash
task setup
```

Then in separate terminals:

```bash
task server-run-recipes
task server-run-users
task server-run-favorites
task web-dev
```

Web app runs on `http://localhost:3000`.

---

## Root command runners

The root runners are the preferred way to execute common operations from repo root.

### Makefile commands

```bash
make help
```

Main targets:

- `setup` - install web deps, start infra, migrate DB, seed sample data
- `check` - backend tests + web build + web tests
- `infra-up` / `infra-down` / `infra-ps` / `infra-logs` / `infra-restart`
- `migrate-up` / `migrate-down` / `migrate-version`
- `seed` / `db-counts`
- `server-run-recipes` / `server-run-users` / `server-run-favorites`
- `server-test` / `server-test-search` / `server-test-auth`
- `web-install` / `web-dev` / `web-build` / `web-start` / `web-lint` / `web-test` / `web-e2e`
- `health`

### Taskfile commands

```bash
task --list
```

Task names mirror the Makefile targets for portability.

---

## Environment variables

Use per-project env files:

- `server/.env.local` (local backend)
- `web/.env.local` (local web)

Never commit real secrets. Commit only examples:

- `server/.env.example`
- `web/.env.example`

Common local backend values:

- `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_DB`
- `JWT_SECRET`, `JWT_ISSUER`, `JWT_ACCESS_TTL`
- `REDIS_URL`
- `CORS_ALLOWED_ORIGINS`

Common web values:

- `NEXT_PUBLIC_API_BASE_URL` (preferred when using one API domain)
- `NEXT_PUBLIC_API_URL`
- `NEXT_PUBLIC_API_RECIPES_PORT`
- `NEXT_PUBLIC_API_USERS_PORT`
- `NEXT_PUBLIC_API_FAVORITES_PORT`

---

## CI/CD at a glance

CI workflow: `.github/workflows/ci.yml`

Deploy workflows:

- `.github/workflows/deploy-staging.yml`
- `.github/workflows/deploy-prod.yml`

On PRs and pushes to `main`, CI runs:

1. Backend `go vet`, `go test`, `go build`
2. Web dependency install, lint, test, and build

On pushes to `main`, CI also builds backend Docker images.

Deployment planning guide: [`docs/ops/deployment-plan.md`](docs/ops/deployment-plan.md)
Hosting strategy comparison: [`docs/ops/hosting-strategy.md`](docs/ops/hosting-strategy.md)
Gateway setup: [`server/gateway/README.md`](server/gateway/README.md)
Provider onboarding checklist: [`docs/ops/provider-onboarding-checklist.md`](docs/ops/provider-onboarding-checklist.md)
Provider fill-in template: [`docs/ops/provider-setup-template.md`](docs/ops/provider-setup-template.md)

### Deploy workflow secrets

Set these repository/environment secrets before enabling deploy workflows:

- `WEB_STAGING_DEPLOY_HOOK`
- `API_RECIPES_STAGING_DEPLOY_HOOK`
- `API_USERS_STAGING_DEPLOY_HOOK`
- `API_FAVORITES_STAGING_DEPLOY_HOOK`
- `WEB_PROD_DEPLOY_HOOK`
- `API_RECIPES_PROD_DEPLOY_HOOK`
- `API_USERS_PROD_DEPLOY_HOOK`
- `API_FAVORITES_PROD_DEPLOY_HOOK`

These are provider deploy webhooks (for example Vercel/Render/Railway service hooks where available).

---

## Public GitHub metadata

This repo includes public collaboration and governance files:

- `LICENSE` - all rights reserved (no use without permission)
- `CONTRIBUTING.md` - contribution workflow and standards
- `SECURITY.md` - vulnerability reporting guidance
- `.github/ISSUE_TEMPLATE/bug_report.md` - structured bug intake (repro steps, expected/actual behavior, environment)
- `.github/ISSUE_TEMPLATE/feature_request.md` - structured feature intake (problem, proposal, scope)
- `.github/ISSUE_TEMPLATE/config.yml` - disables blank issues so reports use templates
- `.github/PULL_REQUEST_TEMPLATE.md` - PR checklist for scope, validation, and deployment notes
- `.github/CODEOWNERS` - review ownership mapping by path
- `.github/dependabot.yml` - weekly dependency update PRs for Go, npm, and GitHub Actions
- `.github/workflows/ci.yml` - CI checks for server and web on PRs and `main`

## GitHub collaboration files

How these files are used in practice:

1. Open a bug with the bug template and include reproducible steps.
2. Open enhancement ideas with the feature template to define scope early.
3. Use the PR template to document what changed, why, testing, and rollout/rollback notes.
4. Follow `CONTRIBUTING.md` for branch/test conventions and `SECURITY.md` for private vulnerability reporting.

---

## Recommended workflow

1. Create a branch from `main`
2. Use `make` or `task` commands from repo root
3. Run checks before opening a PR:

```bash
make check
make web-e2e
```

Task equivalent:

```bash
task check
task web-e2e
```

4. Open PR with migration/deployment notes if applicable

---

## Deep links

- Backend details: [`server/README.md`](server/README.md)
- Frontend details: [`web/README.md`](web/README.md)
- Docs index: [`docs/README.md`](docs/README.md)
- Deployment checklist: [`docs/ops/deployment-plan.md`](docs/ops/deployment-plan.md)
- Hosting tradeoffs: [`docs/ops/hosting-strategy.md`](docs/ops/hosting-strategy.md)
- Gateway setup: [`server/gateway/README.md`](server/gateway/README.md)
- Provider onboarding: [`docs/ops/provider-onboarding-checklist.md`](docs/ops/provider-onboarding-checklist.md)
- Provider setup template: [`docs/ops/provider-setup-template.md`](docs/ops/provider-setup-template.md)
