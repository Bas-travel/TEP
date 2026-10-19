# Security model

## Authentication

- Passwords: Argon2id (or bcrypt with cost ≥ 12 if Argon2 unavailable on platform).
- Access token: JWT, 15m, claims: `sub`, `roles`, `session_id`.
- Refresh token: opaque random, stored hashed in MySQL + metadata in Redis; rotation on each refresh.
- Guest mode: optional anonymous device ID for browse-only; booking requires registered user.

## Authorization (RBAC)

| Role | Capabilities |
|------|----------------|
| `traveler` | Own profile, bookings, wallet, UGC on own posts |
| `moderator` | Hide/delete reported content, view reports |
| `admin` | Full operational APIs (separate admin router prefix `/admin/v1`) |

**Rule:** Every mutating handler validates resource ownership or role on the **server**. Client UI hiding is not security.

## API hardening

- TLS everywhere outside local DEV.
- CORS allowlist per environment.
- Request size limits; JSON depth limits.
- Input validation via struct tags + domain rules.
- SQL: prepared statements only (GORM/sqlx).
- Output encoding for UGC fields in clients.

## Secrets

- Never commit secrets. Templates: `.env.example` per service.
- Production: secret manager / CI injected env.
- PSP webhook signing keys rotated per provider docs.

## Wallet and payments

- Ledger append-only; balance derived or stored with row locks (`SELECT FOR UPDATE`).
- Refunds reference original payment / booking ID (idempotent).
- Audit log table for wallet mutations.

## Community moderation

- Report reasons enum matches prototype.
- Auto-hide post for reporter pending review (configurable).
- Rate limits on post/create comment.

## Compliance notes

- Document PDPA-relevant flows: account export/delete, consent for marketing notifications.
- Legal compliance claims require organizational review — not asserted here.
