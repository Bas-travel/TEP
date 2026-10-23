# API conventions (baseline)

- Base path `/v1`; health probes at server root (`/health/live`, `/health/ready`).
- JSON `camelCase` on the wire, `snake_case` in MySQL. Money = decimal THB, serialized as JSON number with 2 dp.
- Errors: `application/problem+json` — `{type,title,status,detail,code,requestId,errors[]}`.
  Codes: `AUTH_UNAUTHENTICATED, AUTH_FORBIDDEN, NOT_FOUND, VALIDATION_FAILED, WALLET_INSUFFICIENT, RATE_LIMITED, CONFLICT, INTERNAL`.
- Auth: `Authorization: Bearer <JWT>`; JWT 15 min; refresh token opaque, rotated on each `/auth/refresh`.
- Ids: numeric `int64` for DB-backed entities; opaque strings for async objects (`jobId`, payment `intent id` = `pi_…`).
- Idempotency: `Idempotency-Key` header on `POST /bookings` and `POST /payments/intents`.
- Pagination: `?cursor=&limit=` → `{items, nextCursor}`.

## DB ↔ API field mapping

| DB | API |
|----|-----|
| `listings.rating_avg` | `rating` |
| `listings.location_text` | `location` |
| `listings.hero_image_key` | `imageUrl` (resolved via storage base URL) |
| `destinations.name_th` | `name` |
| `community_posts.body` | `text` |
