# ADR 002 — Mock-first delivery

## Status

Accepted

## Context

Universal environment policy requires DEV without production DB, and incremental replacement of mocks with real backend modules.

## Decision

1. Flutter uses repository interfaces with three implementations: `Mock`, `HttpBeta`, `HttpProduction`.
2. Gin API exposes the same OpenAPI; DEV may run `cmd/api` with in-memory adapters for rapid testing.
3. Contract tests validate mock fixtures and live API against OpenAPI schemas.

## Consequences

- Phase 03 starts with OpenAPI + mock server; module swap (e.g. booking → real MySQL) is isolated.
- Duplicate effort to maintain mocks until Beta — mitigated by schema-generated fixtures.
