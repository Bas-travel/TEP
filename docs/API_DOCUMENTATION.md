# API Documentation

Source contract: `docs/api/openapi.yaml`.

Conventions: `docs/api/conventions.md`.

Authentication uses bearer access tokens and rotating opaque refresh tokens. JSON uses camelCase, database uses snake_case, and errors use RFC 7807-compatible `application/problem+json` responses.

Health endpoints are served at `/health/live` and `/health/ready`. API resources are served under `/v1`.

Implementation parity and route coverage must be verified with automated scripts and integration tests after every route change.
