# apps/mobile-flutter — TEP Travel Super App (Sprint 1)

Flutter client that follows the product board (`docs/ux/board-mapping.md`). It talks to the backend only through
`lib/data/api_client.dart`, with two interchangeable implementations:

| `API_MODE` | Implementation | Needs backend? |
|------------|----------------|----------------|
| `mock` (default) | `MockApiClient` — in-memory, same business rules as the API (3 % fee, `WELCOME10`, wallet) | No |
| `http` | `HttpApiClient` → `services/api-go` | Yes (`docker compose up`) |

## Screens in Sprint 1 (board ids)
Login/Register · Home 1.1 · Search 1.2–1.3 (type chips, destination filter, paging) · Listing detail 1.4/1.13 ·
Booking 1.5–1.9 (date, quantity, coupon, price, PromptPay/card/Wallet, success) · My bookings 1.10 · Booking detail + cancel 1.11 ·
Wallet 1.19 (top-up, history) · Profile 1.20 · Help center 1.23.

Deferred (see `docs/requirements/mvp-scope.md`): flights seat selection, tickets/pass, map, weather, messages, community, AI planner, Partner/Admin/Staff portals.

## Run
```bash
cd apps/mobile-flutter
flutter create --platforms=android,ios .   # once: generates android/ ios/ (git-ignored)
flutter pub get
flutter run                                 # mock mode, login demo@tep.dev / Passw0rd!

# against the real API (Android emulator reaches host via 10.0.2.2)
docker compose up --build                   # from repo root
flutter run --dart-define=API_MODE=http --dart-define=API_BASE_URL=http://10.0.2.2:8080
```

## Test
```bash
flutter analyze
flutter test                                              # models, MockApiClient rules, widget journey
TEP_API_URL=http://127.0.0.1:8080 flutter test test/integration   # contract vs a running Go API
```

`lib/data/mock_fixtures.dart` is generated from the same seed as the Go API: `python3 scripts/gen_seed.py`.

## Known limits
- Tokens are kept in memory only (secure storage + auto-login: Sprint 2).
- Listing images: API has no media yet, so the UI renders typed gradient thumbnails (`imageUrl` is honoured when present).
- The booking date chosen in the UI is not sent: the API `POST /bookings` has no stay/visit date field yet (tracked in `docs/review/doc-review-findings.md`, F-1).
