# TEP — แอพท่องเที่ยวครบวงจร (Sabai Travel)

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

## Local setup (future)

Implementation scaffold (Phase 03) will document:

- `services/api-go` — Gin API
- `apps/mobile-flutter` — Flutter client
- `docker compose` — MySQL, Redis, MinIO (local object storage)
