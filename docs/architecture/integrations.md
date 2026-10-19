# External integrations

## Payment (PromptPay + card)

**Flow (server-authoritative)**

1. Client `POST /payments/intents` with booking ID and method.
2. API returns intent ID; for PromptPay, QR payload URL or EMV string; for card, client secret / redirect URL from PSP.
3. PSP webhook `POST /webhooks/payments/{provider}` verifies signature, updates intent status.
4. Client polls `GET /payments/intents/{id}` or receives push on success.
5. API confirms booking in one transaction (booking status + wallet debit if applicable).

**DEV / BETA**

- Mock PSP adapter returns deterministic success after delay (mirrors prototype UX).
- No real money movement in DEV.

## Object storage

- Bucket per environment: `tep-dev`, `tep-beta`, `tep-prod`.
- Upload: `POST /media/presign` → client PUT → `POST /media/complete`.
- Public listing images: CDN URL; EXIF strip on upload pipeline (phase 2).

**Local DEV:** MinIO via Docker Compose.

## Redis

| Key pattern | TTL | Purpose |
|-------------|-----|---------|
| `session:{userId}:{sessionId}` | refresh lifetime | Refresh token metadata |
| `ratelimit:{scope}:{id}` | 1m | Rate limiting |
| `cache:listing:{id}` | 5m | Hot listing detail |
| `ai:job:{id}` | 24h | AI planner job status |
| `otp:{purpose}:{phone}` | 5m | Future phone verify |

## Firebase Cloud Messaging

- Register device: `POST /devices/push-tokens`.
- Server sends on booking confirmed, payment success, trip reminder, moderation outcome.

## AI trip planner (LLM)

**Async contract**

1. `POST /ai/itineraries` → `202` + `jobId`.
2. `GET /ai/itineraries/jobs/{jobId}` → `pending | completed | failed` + result.

**Prompt inputs:** destination, days, budget, styles[], travelers, locale.

**Fallback:** deterministic template engine (same rules as prototype `buildItinerary`) when LLM unavailable.

## Maps

MVP matches prototype: generate Google Maps search / directions URLs server-side or client-side from place names — no Maps API billing in MVP.

## Email / SMS (optional phase 2)

Password reset, receipt email — stub interfaces in `internal/notify`.
