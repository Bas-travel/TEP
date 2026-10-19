# Phase 03 — Implementation plan (next)

Prerequisite: Phase 01–02 artifacts approved.

## Repository layout (target)

```
TEP/
  apps/mobile-flutter/
  services/api-go/
    cmd/api/
    internal/identity/
    internal/catalog/
    ...
    migrations/
  docs/
  deploy/
    docker-compose.dev.yml
  sabai-travel-app-prototype.jsx   # reference until Flutter P0 parity
```

## Sprint order (mock → real)

| Sprint | Deliverable | API mode |
|--------|-------------|----------|
| 1 | Monorepo scaffold, Docker MySQL/Redis/MinIO, health checks | Mock Gin or Prism |
| 2 | Flutter shell: theme, router, 6 tabs, auth screens | Mock repositories |
| 3 | Catalog + search + listing (UF-02 part 1) | Mock → `GET` real |
| 4 | Bookings + payments mock (UF-02 complete) | MySQL bookings module |
| 5 | Wallet + cancel/refund (UF-04, UF-05) | MySQL ledger |
| 6 | AI jobs + saved trips (UF-03) | Redis job + template engine |
| 7 | Community MVP (UF-06) | MySQL UGC |
| 8 | Notifications, reviews, coupons, support | Remaining modules |
| 9 | Phase 04 hardening | CI, tests, BETA deploy |

## Definition of done (Phase 03 milestone 1)

- `docker compose up` runs API + dependencies
- Flutter `API_MODE=mock` demos full UF-02 without backend
- OpenAPI contract test job in CI (schemas valid)

No production deploy without Phase 04 gates.
