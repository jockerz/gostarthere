# Agent Guide — Go Start There

Fullstack web application: **Go (Fiber + Huma)** backend, **Svelte 5 (SvelteKit + shadcn-svelte)** frontend.

## Quick Reference

| | Backend | Frontend |
|---|---|---|
| **Stack** | Go 1.25, Fiber v3, Huma v2, GORM (SQLite/PostgreSQL), asynq (Redis) | SvelteKit 2, Svelte 5 (runes), Tailwind v4, shadcn-svelte, Vite 8 |
| **Entry** | `backend/main.go` | `frontend/src/routes/` |
| **Dev** | `make run` (hot-reload via air) | `npm run dev` |
| **Build** | `make build` | `npm run build` |
| **Test** | `make test` | `npm run check` (type-check only) |
| **Guidelines** | `backend/AGENTS.md` | `frontend/AGENTS.md` |

> **Always read the relevant sub-directory AGENTs.md before editing code in that area.**

## Docker

### Setup
1. Copy `.env.example` to `.env` and set `POSTGRES_PASSWORD` and `REDIS_PASSWORD`
2. Copy `backend/.env.docker.example` to `backend/.env.docker` and configure

### Local Development
```bash
docker compose up -d              # Start all services
docker compose logs -f backend    # Follow backend logs
docker compose down               # Stop all services
docker compose down -v            # Stop and remove volumes (resets DB)
```

Services: Frontend (`http://localhost`), API (`http://localhost/api/v1/...`), Traefik Dashboard (`http://localhost:8081`)

### Production
```bash
ACME_EMAIL=you@example.com docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

## Project Structure

```
gostarthere/
├── backend/                  # Go API + worker
│   ├── main.go               # Cobra CLI: `api` / `worker`
│   ├── apps/api/             # Fiber app, handlers, routes, middleware
│   ├── apps/workers/         # asynq background workers
│   ├── pkg/                  # Business logic (entities, auth, user, oauth2)
│   ├── extensions/           # DB (GORM SQLite), email (SMTP), task queue
│   ├── internal/             # Config
│   └── Makefile
├── frontend/                 # SvelteKit SPA
│   ├── src/lib/              # Shared: components (ui/custom), services, hooks, utils
│   ├── src/routes/           # Routes: auth/* (public), app/* (protected)
│   └── package.json
├── README.md
└── AGENT.md                  # This file
```

## Architecture Principles

### Backend — Clean Architecture

Layers are strictly separated. Dependencies flow inward only:

```
apps/api/handler → pkg/<feature>/service → pkg/entities
                   pkg/<feature>/repository → extensions/database
```

**Import rules (never violate):**
- `pkg/` must never import `apps/*`, Fiber, or Huma
- `pkg/entities` must never import `apps/*` or `pkg/<feature>/*`
- `apps/api/presenter` and `apps/api/schema` must never import GORM
- `apps/api/routes` must never import `pkg/entities` or GORM

**Handler pattern:** Closures returning `func(ctx, input) (output, error)` — business logic lives in services, not handlers.

**Response format:** All API responses use `{ success: bool, message: string, data?: any }`.

**Feature generation order:**
1. `pkg/entities/` — domain models
2. `apps/api/presenter/` — request/response DTOs
3. `pkg/<feature>/repository.go` — data access
4. `pkg/<feature>/service.go` — business logic
5. `apps/api/handler/` — HTTP handlers
6. `apps/api/schema/` — Huma operation definitions
7. `apps/api/routes/` — route registration

### Frontend — SvelteKit SPA

- SSR disabled, prerender enabled (adapter-static, pure SPA)
- Svelte 5 runes only: `$props()`, `$state()`, `$derived()`, `$effect()`
- Use `{@render children()}`, not `<slot />`
- Auth guard in `app/+layout.ts` (localStorage token check)
- API client: `src/lib/services/api.ts` with typed `apiCall<T>()`

## Conventions

### Backend

- **Indentation:** tabs
- **Error handling:** domain errors as sentinel `var` in services; handlers translate to HTTP codes
- **DI:** constructor injection (`NewRepository(db)`, `NewService(repo)`)
- **Passwords/tokens:** bcrypt-hashed, dual-token scheme (`prefix.secret` wrapped in JWT)
- **DB migration:** GORM AutoMigrate on startup (no separate migration files)

### Frontend

- **Indentation:** tabs
- **Quotes:** double quotes
- **Semicolons:** required
- **Components:** shadcn in `ui/`, custom in `custom/[feature]/`
- **Naming:** PascalCase components, kebab-case routes, camelCase services
- **Path aliases:** `$lib/` or `@/` → `src/lib`

## Environment

### Backend (`.env`)

Key variables: `SECRET`, `DB_URL`, `REDIS_HOST/PORT/DB/PASS`, `SMTP_HOST/PORT`, `OAUTH_REDIRECT_BASE`, `OAUTH_GITHUB_CLIENT_ID/SECRET`, `OAUTH_GOOGLE_CLIENT_ID/SECRET`.

### Frontend (`.env`)

- `PUBLIC_API_URL` — backend API base URL
- `OAUTH_REDIRECT_BASE` — OAuth callback base

## Testing

### Backend

- **Repository tests:** in-memory SQLite (`:memory:`)
- **Service tests:** hand-written mock repositories (no framework), `SetSkipTaskQueue(true)` to skip Redis
- **Handler tests:** `humatest.New(t)` + `testify/assert`
- Run: `make test`

### Frontend

- No test framework configured; run `npm run check` for type checking

## Agent Rules

1. Never place business logic in `apps/` (backend) or `routes/` (frontend)
2. Respect layer import boundaries — check `backend/AGENTS.md` table
3. Use constructor injection for all dependencies
4. Return domain errors from services; handlers translate to HTTP
5. Keep presenter DTOs JSON-only — no GORM tags
6. Prefer editing existing files over creating new ones
7. Run `make test` (backend) or `npm run check` (frontend) before claiming work is done
8. Do not commit `.env` files or secrets
