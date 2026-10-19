# QA, security, performance, and release checklist (Phase 04)

Use before BETA promotion and before PRODUCTION. Mark each item PASS / FAIL / NOT VERIFIED with evidence links (CI run, report, screenshot).

## Automated tests

| Layer | Scope | Tool (planned) |
|-------|--------|----------------|
| Unit | Go domain logic, Flutter repositories | `go test`, `flutter test` |
| API contract | OpenAPI examples vs responses | Schemathesis / Dredd |
| Integration | API + MySQL + Redis | testcontainers-go |
| E2E | UF-01–UF-08 critical paths | Flutter integration / Maestro |
| Regression | Booking + wallet + cancel | CI on `main` |

## Security

- [ ] OWASP API Top 10 review on auth, booking, wallet, webhooks
- [ ] IDOR tests on bookings, wallet, posts (user A cannot access user B)
- [ ] Rate limits on login, post create, report
- [ ] Webhook signature verification tests
- [ ] Secrets scan (gitleaks) in CI
- [ ] Dependency audit (`govulncheck`, `flutter pub outdated`)

## Performance

- [ ] p95 latency targets: catalog read < 300ms, booking write < 800ms (BETA)
- [ ] Load test k6 on search + feed read paths
- [ ] Redis cache hit ratio monitored for listing detail

## Data and backup

- [ ] MySQL automated backup daily; restore drill documented
- [ ] Migration tested on BETA clone before PROD
- [ ] Object storage lifecycle rules for orphaned uploads

## Deployment

- [ ] DEV → BETA → PROD config separation verified
- [ ] Health checks `/health/live`, `/health/ready`
- [ ] Rollback procedure for API + migrations
- [ ] Post-deploy smoke: login, search, mock payment, one booking cancel in BETA

## Release gates (must pass)

1. Acceptance criteria for MVP P0 screens (see screen inventory)
2. Contract tests green
3. No critical/high open security findings
4. UAT sign-off on BETA URL
5. Explicit approval for PRODUCTION deploy

## Current status

| Environment | Deploy | Tests | Notes |
|-------------|--------|-------|-------|
| DEV | NOT VERIFIED | NOT VERIFIED | Docs + prototype only |
| BETA | NOT IMPLEMENTED | — | — |
| PRODUCTION | NOT IMPLEMENTED | — | — |
