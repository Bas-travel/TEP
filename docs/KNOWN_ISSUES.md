# Known Issues

_Last reviewed: 2026-11-22 (branch `feat/phase-03-sprint-1`)._

## Not verified yet (needs the user's machine)
- `docker compose up --build` on a real Docker daemon (the authoring sandbox has no daemon). MySQL 8 was verified separately with the real API binary.
- Flutter on a device/emulator/browser: CI (analyze + 47 tests + contract test) passes, but no human UI pass after the `main` UI merge.
- PowerShell scripts (`setup-env.ps1`, `db-backup.ps1`, `db-restore.ps1`) have not been executed.

## Deferred by scope (`docs/requirements/mvp-scope.md`)
- Partner payouts / bank accounts (earnings are reported gross / commission / net only).
- Live chat (staff support uses tickets; live chat needs a WebSocket service).
- Moderator portal (role kept in the schema, unused by the portals).
- Real payment gateway, FCM push, S3 media lifecycle, external weather/LLM services (need credentials). Payment runs on a mock PSP + signed webhook.
- Map and weather use deferred integrations or external deep links.

## Technical debt
- `package:flutter/material.dart` is deprecated in the newest Flutter stable; CI runs `flutter analyze --no-fatal-infos` until the migration is done as its own change.
- Mock-mode traveler catalog is static (portal listing edits are not reflected to travelers in mock mode; HTTP mode is correct).
- Rate limiting is in-process (not Redis-backed); refresh-token reuse hardening is covered by tests but not by a Redis store.
- Secure token storage on devices is not implemented (tokens live in memory only).
- Go test coverage (in-memory run): api 62.8%, auth 74.5%, config 75.8%, db 17.1%, httpx 0%. Target is >= 80% for business logic.

## Environments
- No BETA or PRODUCTION deployment exists or is claimed.
