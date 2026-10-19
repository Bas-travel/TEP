# ADR 001 — Technology stack

## Status

Accepted

## Context

TEP requires a mobile-first travel app with booking, wallet, community UGC, and AI planning. Team selected Go + Gin, Flutter, MySQL, Redis, and object storage.

## Decision

- **Mobile:** Flutter (single codebase for iOS/Android; design parity with existing React prototype).
- **API:** Go 1.22+ with Gin; modular monolith.
- **Database:** MySQL 8 (InnoDB, utf8mb4).
- **Cache:** Redis 7 for sessions, rate limits, hot cache, AI job status.
- **Files:** S3-compatible storage + CDN in production.

## Consequences

- Strong typing on mobile and API; shared OpenAPI for contract tests.
- Ops must run MySQL + Redis + bucket per environment.
- Team needs Go and Flutter skills; React prototype remains reference only until Flutter reaches parity.
