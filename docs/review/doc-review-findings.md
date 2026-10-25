# Document review — baseline findings (Step 1)

Reviewed commit: `7f28ef1` (main). Reviewer: Claude. Date: 2026-10-23.
Scope: `docs/api/openapi.yaml`, `docs/database/erd.md`, architecture/security/UX docs, MVP scope vs. the
"Travel Super App" screen board (30 traveler screens + Partner / Admin / Staff portals).

Status legend: **FIXED** (changed in this branch) · **OPEN** (needs a decision or later sprint).

## A. Blocking / high

| # | Finding | Evidence | Status |
|---|---------|----------|--------|
| A1 | `README.md` contains leftover merge-conflict markers (`=======`, `>>>>>>> 8fcdc60…`) and a duplicate stub title. | README.md tail | FIXED |
| A2 | Error model is incomplete. Only `409`/`401` are declared on 2 operations; **50 of 52 operations declare no 4xx**. Docs promise RFC 7807 but `Problem` has no `code`, `requestId` or field-level `errors`. `402` (wallet insufficient, REQ-PAY-02), `403`, `404`, `422`, `429` absent. | openapi.yaml `components.responses` has only `Problem` | FIXED (shared responses + `Problem` fields added; per-operation refs added via default responses) |
| A3 | Architecture says `Idempotency-Key` on `POST /payments/confirm`, but no such endpoint exists (contract has `POST /payments/intents`). Idempotency header is declared only on `POST /bookings` and the intent endpoint. | architecture/README.md L97 vs openapi | FIXED (README wording aligned to `/payments/intents`) |
| A4 | Health endpoints `/health/live` and `/health/ready` are required by the release checklist and environment strategy but are not in OpenAPI. | qa/release-checklist.md | FIXED (added, unauthenticated, outside `/v1`) |
| A5 | **MVP scope mismatch with the screen board.** Board shows Partner (4 screens), Admin (4), Staff (4) portals and traveler features (flights, restaurants, tickets/pass, messages, AI assistant, weather, map) that have no API/ERD/screen-ID. Docs list 31 screens of a different numbering (S01–S31) and only roles `traveler/moderator/admin`; architecture also lists `local_guide`. Role set differs between architecture README, security-model and the board (User / Partner / Admin / Staff). | board vs screen-inventory.md | OPEN → see `docs/requirements/mvp-scope.md` (proposal; needs owner confirmation) |

## B. Medium

| # | Finding | Status |
|---|---------|--------|
| B1 | ERD diagram relations reference entities that have no table block: `booking_line_items`, `trip_days`, `trip_day_items`, `reviews`, `notifications`, `support_tickets`, `user_sessions`, `device_push_tokens`, `community_groups`, `community_group_members`, `user_follows`, `post_reports`. | FIXED in `erd.md` (MVP additions section + DDL in migration `0001`) |
| B2 | Tables implied by the API but absent from ERD: `faqs`, `post_likes`, `post_saves`, `user_settings`, `payment_intents` (API has an intent object; ERD has 1:1 `payments`), `webhook_events` (idempotent webhook processing), `idempotency_keys`, `audit_logs` (security-model requires wallet audit log). | FIXED (added to ERD list + migration) |
| B3 | `POST /notifications` is used for "mark all read" — non-RESTful and ambiguous with create. | FIXED → `POST /notifications/read-all` |
| B4 | ID typing inconsistent: DB `bigint`; API uses `int64` for most ids but `string` for `authorId`, `jobId`, payment intent `id`. Decision: external ids are opaque strings for jobs/intents; numeric for the rest. | OPEN (documented in `docs/api/conventions.md`) |
| B5 | Field naming drift DB↔API (`rating_avg`↔`rating`, `location_text`↔`location`, `hero_image_key`↔`imageUrl`, `name_th`↔`name`). Acceptable only with an explicit mapping. | FIXED (mapping table in `conventions.md`) |
| B6 | Pagination: cursor on 2 endpoints; bookings, notifications, trips, wallet transactions, reviews lists are unbounded. | OPEN → Sprint 4+ |
| B7 | Auth gaps: no logout / session revoke endpoint, no guest-token policy in API (docs say "no token"), refresh endpoint has no 401 response, no rate-limit response. Refresh rotation + reuse detection not specified. | FIXED partly (`POST /auth/logout`, 401/429 added); reuse-detection OPEN |
| B8 | `BookingSummary.status` enum is `confirmed/completed/cancelled` but booking draft + payment pending state is implied by `POST /bookings` ("draft") → no `pending_payment` state. | FIXED (added `pending_payment`) |
| B9 | `UserSettings.language` + profile `language` duplicated; `theme` has `system`, prototype only light/dark. | OPEN (low) |
| B10 | Servers use placeholder `api.example.com` domains. | OPEN (needs real domains) |
| B11 | Brand naming: README "Sabai Travel", board "Travel Super App". | FIXED — owner decision 2026-10-25: **Sabai jai Travel** (repo/code name stays `TEP`) |

## C. Authentication / security consistency

- JWT 15 min + opaque rotating refresh: consistent between security-model and OpenAPI (`bearerAuth`).
- `security: []` correctly set for register, login, refresh, catalog reads, FAQ, webhook.
- Webhook is `security: []` — correct, but contract must state signature header (added as `X-Signature`).
- Admin prefix `/admin/v1` described in security-model but **no admin paths** exist in OpenAPI → OPEN (Phase 2 with Partner/Staff).
- Default JWT secret startup check is specified in environment-strategy → implemented in Sprint 1 (`config.Validate`).

## D. Migration plan

- Tool chosen: "golang-migrate **or** goose" → decision recorded: **golang-migrate-compatible SQL files** (`NNNN_name.up.sql` / `.down.sql`) applied by a tiny in-repo runner, so no external binary is needed in DEV. ERD says "forward-only" but also wants rollback scripts → kept `.down.sql` for DEV only.
- Seed data: ERD says seeds mirror prototype constants → `0002_seed_dev.up.sql` (DEV/BETA only; guarded by `APP_ENV`).

## E. Traceability

- Traceability matrix lists 23 REQs; screen-inventory references REQ-IDs not in the matrix (REQ-TRP-02, REQ-COM-03/04/05, REQ-PRF-01, REQ-REV-01, REQ-SUP-02, REQ-WLT-02/03, REQ-BKG-02). → OPEN, tracked for Sprint 2.

## Verification note

Checked mechanically: all `$ref` targets resolve; YAML parses; operation list extracted (**55 ops** — corrected in review round 2; round 1 said 52).
NOT verified: OpenAPI against a validator tool (none installable offline) — only structural checks by script.


## Round 2 (2026-10-24) — independent re-verification

Re-ran structural checks (YAML parse, `$ref` resolution, per-operation response codes) instead of trusting round 1.

| # | Finding | Status |
|---|---------|--------|
| R2-1 | Round 1 reported 52 operations; actual count is **55**. | FIXED (text corrected) |
| R2-2 | `POST /auth/refresh` had no `401` although B7 marked it fixed (invalid / reused refresh token). | FIXED |
| R2-3 | `GET /bookings?status=` enum lacked `pending_payment` (added in B8 to `BookingSummary` only). | FIXED |
| R2-4 | `UserProfile` was `allOf RegisterRequest`, so `GET /me` advertised a required `password` field. | FIXED (explicit schema, no password; `avatarUrl` added for board screen 1.20) |
| R2-5 | `ListingSummary` has no thumbnail but the board's list screens (1.3, 1.12, 1.14) show images. | FIXED (`imageUrl`) |
| R2-6 | `Problem` had no `required` fields; clients cannot rely on `code`. | FIXED |
| R2-7 | **Design tokens conflict.** `screen-inventory.md` documents pine `#10403D` / amber `#E3A23C` / Fraunces (old React prototype). The "Travel Super App" board uses navy `#002765`-`#082D66`, teal `#23925D` success, light `#F4F8FC` background. | SUPERSEDED — owner chose theme **Sunset Journey** on 2026-10-25 (see `docs/ux/design-tokens.md`). Board navy palette and old pine/amber tokens are legacy. |
| R2-8 | Migration file `0001_init.up.sql` is referenced by `erd.md` as authoritative DDL but did not exist. | FIXED in Sprint 1 (`services/api-go/migrations/`) |
| R2-9 | ADR-001 specifies Gin. `golang.org/x/*` modules cannot be fetched in the build sandbox, so Gin cannot be compiled there. | DECISION: ADR-004 (net/http in Sprint 1, handlers framework-agnostic). ACCEPTED by owner 2026-10-25. |
| R2-10 | Board lists WebSocket real-time and Stripe/PromptPay; docs defer WebSocket (live chat) and have PSP behind mock adapter. | Consistent with `mvp-scope.md`; no change. |

Still NOT verified: OpenAPI validated by a real validator (none installable offline); Flutter code could not be compiled in the sandbox (see Sprint 1 report).


## Round 3 (2026-10-25) — Sprint 1 integration findings (UI ↔ API)

| # | Finding | Status |
|---|---------|--------|
| F-1 | **No stay/visit date or guest count in `POST /bookings`.** Board screens 1.5–1.7 (choose date, guests) cannot be persisted: `BookingLine` has only `listingId` + `quantity`. The Flutter UI collects the date but does not send it. | OPEN → proposal: add `startDate`, `endDate?` (date), `guests` to the booking line + `booking_line_items` columns; needs owner decision before Sprint 3 |
| F-2 | No media/image API: listing/destination have no image; board shows photos. UI uses typed gradient thumbnails and honours optional `imageUrl`. | OPEN (MinIO + signed URLs, Sprint 4+) |
| F-3 | Booking number format differs: API `TEP-yyMMdd-00001`; mock used `TEP-yyyy-00001`. | FIXED (mock aligned) |
| F-4 | Timestamps are emitted with nanoseconds and the server's local offset (`2026-10-25T03:06:10.423782765+07:00`). Dart parses it, but contract should be RFC 3339 UTC with milliseconds. | OPEN (low) — change in `httpx`/types when DB mode lands |
| F-5 | Flutter refresh-token rotation: client refreshes once on 401 and retries; reuse-detection still OPEN (B7). Tokens are in memory only. | OPEN (Sprint 2: secure storage) |
| F-6 | Wallet balance in DEV seed is ฿1,500 (demo user); journeys that pay by Wallet must pick a listing ≤ ~฿1,450 (tests use listing 7). | NOTE |

### Verification performed (2026-10-25)
- `go vet`, `go test -race ./...` — pass. `gofmt -l` — clean. `python3 scripts/gen_seed.py --check` — no drift.
- Migrations applied to a real MySQL 8.0.46: `0001`+`0002` up, then `0002`+`0001` down — all succeed; API started with `MIGRATE_ON_START`, `/health/ready` returned `mysql: ok, redis: ok`; `-healthcheck` exit 0.
- Live curl journey on the running API: login → create booking (idempotent) → pay with Wallet → cancel → refund → 401 without token → 422 problem+json with field errors.
- `docker compose config` parses with `.env.example` values.
- NOT verified locally: `docker compose up` (no Docker daemon in the sandbox) and **all Flutter code** (Flutter/Dart SDK downloads are blocked in the sandbox). Flutter analyze/test and the HttpApiClient↔API contract test are delegated to CI.


## Round 4 (2026-10-25) — CI repair + engagement features

| # | Finding | Status |
|---|---------|--------|
| R4-1 | CI red on `155f212`: Home search bar `Row` overflowed (no `Expanded`) and category labels wrapped past the 72×94 tile. | FIXED (`home_screen.dart`) |
| R4-2 | Bookings and Wallet tabs stayed stale after paying because `IndexedStack` keeps tabs alive; the journey test caught it. | FIXED (tabs reload on switch, `shell.dart`) |
| R4-3 | Local WIP was inconsistent (`shell.dart` passed `onTab` that `HomeScreen` did not accept). | FIXED before commit |
| R4-4 | Route/OpenAPI parity script: 37 Go routes, 0 outside the spec, 18 spec operations still unimplemented (listed in `qa/sprint-1-acceptance.md`). | NOTE |
| R4-5 | `master` on the remote has unrelated history (README with merge-conflict markers). Not touched. | OPEN — owner decision |
