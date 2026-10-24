# "Travel Super App" board → screens, APIs, MVP status

Source: product board image (30 traveler screens, Partner / Admin / Staff portals, main menu, roles, tech stack, user flow).
Visual tokens sampled from the board: header navy `#002765`, bar `#082D66`, page bg `#F4F8FC`, success green `#23925D`, partner teal `#3988A2`, admin violet `#9192C0`, staff orange.

## Main menu (12) and roles
Home · Search · My Bookings · My Trips · Tickets & Pass · Wallet · Coupons · Notifications · Messages · Profile · Settings · Support.
Roles: User (traveler) · Partner · Admin · Staff. **Sprint 1–2 implements User only.**

## Traveler screens
| Board | Screen | API | MVP |
|-------|--------|-----|-----|
| 1.1 | Home | `GET /catalog/destinations`, `GET /catalog/listings` | P0 |
| 1.2 | Search | `GET /catalog/listings?q&type` | P0 |
| 1.3 | Search results | `GET /catalog/listings` | P0 |
| 1.4 | Listing detail | `GET /catalog/listings/{id}` | P0 |
| 1.5 | Choose option (flight variant deferred) | listing + `POST /bookings` | P0 |
| 1.6 | Choose date | client state | P0 |
| 1.7 | Guests / traveler info | client state | P0 |
| 1.8 | Pay | `POST /payments/intents` | P0 |
| 1.9 | Booking success | `GET /bookings/{id}` | P0 |
| 1.10 | My bookings | `GET /bookings` | P0 |
| 1.11 | Booking detail / cancel | `GET /bookings/{id}`, `POST …/cancel` | P0 |
| 1.12 | Hotels | `GET /catalog/listings?type=hotel` | P0 |
| 1.13 | Hotel detail | `GET /catalog/listings/{id}` | P0 |
| 1.14 | Restaurants | `?type=restaurant` | P1 |
| 1.15 | Reviews & ratings | `GET/POST /reviews` | P1 |
| 1.16 | Map / tickets & pass | deep link only / deferred | Deferred |
| 1.17 | Itinerary (My Trips) | `GET /trips/saved` | P1 |
| 1.18 | Coupons | `GET /coupons` | P1 |
| 1.19 | Wallet | `GET /wallet`, `POST /wallet/top-ups` | P0 |
| 1.20 | Profile | `GET /me` | P0 |
| 1.21 | Notifications | `GET /notifications` | P1 |
| 1.22 | Messages | stub (WebSocket deferred) | Stub |
| 1.23 | Help center | `GET /support/faqs` | P1 |
| 1.24 | Contact | `POST /support/tickets` | P1 |
| 1.25 | Community feed | `GET /community/feed` | P1 |
| 1.26 | Settings | `GET/PATCH /me/settings` | P1 |
| 1.27 | AI assistant | `POST /ai/itineraries` | P1 |
| 1.28 | City guide | `GET /catalog/destinations` | P1 |
| 1.29 | Nearby map | maps deep link | Deferred |
| 1.30 | Weather | external | Deferred |

Portals (Partner 2.1–2.4, Admin 3.1–3.4, Staff 4.1–4.4): deferred; no API/ERD (see `requirements/mvp-scope.md`).

## Sprint 1 UI delivered
Shell with bottom nav (Home, Search, Bookings, Wallet, Profile), Home, Search+results, Listing detail, Bookings list, Wallet, Profile, Login — running on `ApiClient` with `mock` and `http` implementations.
