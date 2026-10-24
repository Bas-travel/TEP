# ADR-004 — net/http for Sprint 1 (Gin deferred)

Status: **Proposed** (needs owner confirmation) · Date: 2026-10-24 · Amends ADR-001

## Context
ADR-001 selects Go + Gin. The Sprint 1 build environment can reach GitHub but not `golang.org`, so Gin's
transitive dependencies (`golang.org/x/net`, `x/sys`, `x/crypto`) cannot be downloaded and Gin code cannot be compiled or tested there.

## Decision
Sprint 1 API uses Go 1.22+ `net/http` `ServeMux` (method + path patterns) and `database/sql` with
`github.com/go-sql-driver/mysql` v1.7.1 (no transitive deps). Handlers are plain `http.Handler`s with
middleware as `func(http.Handler) http.Handler`, so they can be mounted in Gin with `gin.WrapH` or ported mechanically.

## Consequences
- Zero third-party runtime deps except the MySQL driver; small attack surface.
- Router/middleware conveniences (binding, validation) are hand-written in `internal/httpx`.
- Revisit when CI has normal module access: switch to Gin if the team still wants it.
