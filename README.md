# TEP — แอพท่องเที่ยวครบวงจร (Sabai jai Travel)

Repository: [Bas-travel/TEP](https://github.com/Bas-travel/TEP.git)

## Stack (approved)

| Layer | Technology |
|-------|------------|
| Mobile | Flutter |
| API | Go + Gin |
| Database | MySQL 8 |
| Cache / sessions | Redis |
| Media | S3-compatible object storage |
| Integrations | Payment (PromptPay / card PSP), FCM, LLM (AI planner), maps deep links |

## Environments

`DEV` (mock / local) → `BETA` (isolated DB + preview) → `PRODUCTION`

See [docs/architecture/README.md](docs/architecture/README.md) and [docs/architecture/environment-strategy.md](docs/architecture/environment-strategy.md).

## Documentation index

| Phase | Artifact |
|-------|----------|
| 01 — Architecture | [docs/architecture/](docs/architecture/) · [OpenAPI](docs/api/openapi.yaml) · [ERD](docs/database/erd.md) |
| 02 — UX/UI | [Screen inventory](docs/ux/screen-inventory.md) · [User flows](docs/ux/user-flows.md) |
| 03 — Implementation | `apps/` and `services/` (planned) |
| 04 — QA / Release | [docs/qa/release-checklist.md](docs/qa/release-checklist.md) |

## Prototype (reference UI)

Interactive React prototype (DEV mock only): [sabai-travel-app-prototype.jsx](sabai-travel-app-prototype.jsx)

Flutter implementation should preserve UX parity documented in `docs/ux/`.

## Local setup (Phase 03 · Sprint 1)

| Path | What |
|------|------|
| `services/api-go` | Go API (net/http in Sprint 1, see ADR-004): health, auth, catalog, bookings, payments (mock PSP), wallet; embedded SQL migrations |
| `apps/mobile-flutter` | Flutter client, `API_MODE=mock` (offline) or `http` (real API) — see its README |
| `docker-compose.yml` | MySQL 8, Redis 7, MinIO, API |
| `scripts/gen_seed.py` | One source for seed data → SQL migration, Go fixtures, Flutter mock fixtures (`--check` detects drift) |

```bash
cp .env.example .env              # DEV placeholders only; .env is git-ignored — never commit it
docker compose up --build         # API on http://localhost:8080
curl localhost:8080/health/ready  # {"checks":{"mysql":"ok","redis":"ok"},"status":"ok",...}

# API without Docker (in-memory mock mode)
cd services/api-go && APP_ENV=dev JWT_SECRET=$(openssl rand -hex 24) go run ./cmd/api
go test ./...                     # unit + HTTP tests

# Mobile
cd apps/mobile-flutter && flutter pub get && flutter test && flutter run   # login demo@tep.dev / Passw0rd!
```

CI (`.github/workflows/ci.yml`): Go vet/test + migrations on MySQL 8, Flutter analyze/test, and a contract test that runs the
Flutter `HttpApiClient` against the real Go API.
