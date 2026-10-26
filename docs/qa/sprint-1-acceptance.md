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

## Update — Sprint 2 slice (2026-10-27): full API + MySQL persistence

Verified locally (MySQL 8.0.46, Go 1.24):

| Check | Result |
|-------|--------|
| OpenAPI vs routes | 55 operations in spec; 53 `/v1` routes match 1:1, the other 2 are `/health/live` and `/health/ready` (served at the root). 0 routes outside the spec |
| `go vet`, `gofmt`, `go test -race` — in-memory repo | PASS |
| Same `internal/api` suite on MySQL (`TEP_TEST_MYSQL_DSN`, tables truncated + reseeded per test, `-race -p 1`) | PASS — 19 top-level tests, 0 FAIL, no data race |
| Migrations 0001→0004 up, 0004→0003 down, up again, then all down to an empty schema | PASS (found and fixed an FK-index error in the 0004 down script) |
| Real binary, `API_MODE=db`: create booking (dates + 3 guests), post, settings, saved trip with itinerary, join group → restart process → read back | PASS — all data survived the restart |

CI run 40378313563 on `67e9b53`: green, including the new "Same test-suite on the MySQL repository" step in the API job (real MySQL 8 service).

## Not done / carried over
- `docker compose up` was never run (no Docker daemon in the sandbox); compose file is parse-checked only.
- Flutter has not been executed on a device/emulator; the UI for the new operations (create post, comments, groups/members, change password, delete account) is not built yet — the Dart `ApiClient` does not expose them.
- Catalog (destinations/listings) is still served from Go fixtures; the `listings` table is only seeded when `MIGRATE_SEED=true`.
- Real PSP and S3 are not exercised (mock PSP; presign verified against the AWS SigV4 test vector only).
- Owner decisions recorded: brand **Sabai jai Travel**, ADR-004 (net/http), theme **Sunset Journey**.
