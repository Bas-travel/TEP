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
| B11 | Brand naming: README "Sabai Travel", board "Travel Super App". | OPEN (owner decision; code uses `TEP Travel Super App`) |

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

Checked mechanically: all `$ref` targets resolve; YAML parses; operation list extracted (52 ops).
NOT verified: OpenAPI against a validator tool (none installable offline) — only structural checks by script.
