# System Architecture

## Verified stack

- Mobile: Flutter
- API: Go `net/http` for current Sprint implementation
- Database: MySQL 8
- Cache: Redis
- Object storage: S3-compatible MinIO for local development
- Container runtime: Docker Compose

## Runtime boundaries

- Mobile uses `MockApiClient` or `HttpApiClient` behind a shared `ApiClient` interface.
- Go API handlers validate HTTP requests and enforce authentication/authorization.
- Domain repositories support in-memory and SQL implementations.
- SQL migrations are embedded and applied at startup when configured.
- External payment, LLM, push, and media integrations are mock-first and must be credentials/configuration gated.

## Phase status

Sprint 1 and Sprint 2 foundation are present. Enterprise partner/admin/staff operations, external gateways, and production infrastructure remain controlled implementation phases rather than claimed completed features.
