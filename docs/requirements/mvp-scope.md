# MVP scope — proposal (derived from the "Travel Super App" board)

Status: **PROPOSAL — needs product-owner confirmation.** Source: screen board (30 traveler screens, Partner/Admin/Staff portals).

## In MVP (traveler app, P0)

Home 1.1 · Search 1.2–1.3 · Listing detail 1.4 · Booking 1.5–1.9 (choose, date, guests, payment, success) ·
My Bookings 1.10–1.11 · Hotels/Stays 1.12–1.13 · Restaurants 1.14 · Reviews 1.15 · Itinerary/AI plan 1.17 ·
Coupons 1.18 · Wallet 1.19 · Profile 1.20 · Notifications 1.21 · Messages 1.22 (read-only inbox stub) ·
Help 1.23 · Contact 1.24 · Settings 1.26 · AI assistant 1.27 (reuse `/ai/itineraries`).

Sprint 1 UI delivers the **shell + 6 main tabs + Home, Search, Listing detail, Bookings, Wallet, Profile** on the mock client; remaining screens follow Sprints 2–8.

## Deferred (post-MVP)

| Item | Reason |
|------|--------|
| Partner portal (2.1–2.4), Staff (4.1–4.4) | No API/ERD yet; needs `partners` + RBAC roles |
| Admin dashboard (3.1–3.4) | `/admin/v1` not in OpenAPI |
| Flights booking (1.5 flight variant), Tickets/Pass 1.16 | No supplier integration in docs |
| Weather 1.30, Map 1.16/1.29 | MVP = maps deep links only (integrations.md) |
| Live chat 1.22 / 4.3 | Needs WebSocket service |

## Role model decision (needs confirmation)

Board: User · Partner · Admin · Staff. Docs: traveler · moderator · admin (+ future local_guide).
Sprint 1 implements `traveler` only; role claim carried in JWT for later extension.
