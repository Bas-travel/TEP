# ADR 003 — Monolith-first backend

## Status

Accepted

## Context

Early product needs speed and transactional consistency across booking, wallet, and payments.

## Decision

Single deployable Gin application with internal packages per domain module. Async AI jobs may use goroutines + Redis in MVP; extract worker process only when CPU/latency requires.

## Consequences

- Simple deployment for BETA/PROD.
- Must enforce module boundaries in code review to avoid spaghetti.
- Future extraction of `community` or `payments` to services is possible with clear API seams.
