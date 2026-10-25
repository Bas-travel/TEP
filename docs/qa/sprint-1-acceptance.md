# Sprint 1 — Acceptance report (Phase 03)

Branch `feat/phase-03-sprint-1` · verified 2026-10-25 · CI run on `8a8b43f`: **all 3 jobs green**
(API + migrations on MySQL 8 · Flutter analyze + test · Contract Flutter `HttpApiClient` ↔ Go API).

| # | Acceptance criterion | Result | Evidence |
|---|----------------------|--------|----------|
| AC-1 | `docker-compose.yml` defines MySQL 8, Redis 7, MinIO, API; config parses | PASS (parse only) | `docker compose config` with `.env.example`. **`docker compose up` NOT run** — no Docker daemon in the sandbox |
| AC-2 | `services/api-go` builds, `go vet` + `gofmt` clean, tests pass with `-race` | PASS | 21 test/subtest PASS lines, 0 FAIL; CI job green |
| AC-3 | Health: `/health/live`, `/health/ready` (mysql/redis state), `-healthcheck` flag | PASS | live curl in mock mode → `{"status":"ok","checks":{"mysql":"disabled","redis":"disabled"}}`; ready with `mysql: ok, redis: ok` verified in CI against MySQL 8 |
| AC-4 | Migrations `0001`/`0002` up and down on MySQL 8 | PASS | CI job + manual run (2026-10-25) |
| AC-5 | `.env.example` has placeholders only; `.env` ignored; startup refuses placeholder JWT secret outside dev | PASS | `config_test.go`, secret scan of the tree (no findings) |
| AC-6 | Flutter app runs with `API_MODE=mock` (no backend) and `http` | PASS (CI) | `flutter analyze`: no issues; widget + unit tests green in CI. Not run on a device/emulator |
| AC-7 | Mock client and API agree on the contract | PASS | Contract job runs the real `HttpApiClient` against the real Go API (book → pay with Wallet → cancel → refund) |
| AC-8 | Every Go route exists in `docs/api/openapi.yaml` | PASS | script: 37 routes, 0 outside the spec |
| AC-9 | UI covers the board's 30 traveler screens + 12-entry main menu | PASS (functional MVP) | Menu test asserts all 12 entries; Map / Weather / Tickets-map are explicit "coming soon" per `mvp-scope.md`; Partner/Admin/Staff are a read-only preview |
| AC-10 | Error model: RFC 7807 with `code`, `requestId`, field `errors` | PASS | live: `POST /v1/reviews` rating 9 → 422 `VALIDATION_FAILED` with `errors[0].field=rating`; no token → 401 |

## Not done / carried over
- 18 of 55 OpenAPI operations are not implemented in the API (community write/groups/follow, push tokens, media presign, payment webhook, password change, `DELETE /me`) — post-MVP.
- Engagement endpoints (coupons, notifications, trips, reviews, tickets, community feed) run on the in-memory store only; no MySQL persistence yet (migration tables exist).
- Booking date/guests not persisted (finding F-1) — owner decision needed.
- Flutter never executed on a device; visual QA against the board still needed.
- Branding (`Sabai Travel` vs `Travel Super App`), ADR-004 (net/http instead of Gin) and design-token decision await owner confirmation.
