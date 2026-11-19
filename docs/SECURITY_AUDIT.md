# Security Audit

## Current verified controls

- JWT-backed authenticated routes.
- Password and refresh-token handling tests.
- Problem+JSON error responses.
- DEV-only secret placeholders rejected outside DEV.
- SQL repository transaction paths for critical operations.

## Not yet verified

- External penetration testing.
- Device secure-token storage.
- Production CORS and TLS configuration.
- Redis-backed refresh reuse controls.
- Real webhook signatures and secret rotation.
- Dependency vulnerability and full OWASP API review.

Evidence and dates must not be marked PASS until commands or automated tests execute successfully.
