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
| Partner payout runs / bank accounts | Earnings are reported (gross, commission, net) but money is not transferred in MVP |
| Staff live chat 4.3 | Needs WebSocket service; tickets with replies cover support in MVP |
| Flights booking (1.5 flight variant), Tickets/Pass 1.16 | No supplier integration in docs |
| Weather 1.30, Map 1.16/1.29 | MVP = maps deep links only (integrations.md) |
| Live chat 1.22 / 4.3 | Needs WebSocket service |

## Role model decision (needs confirmation)

Board: User · Partner · Admin · Staff. **Decision (owner request, Phase 3): the three portals are in MVP scope.**
Roles: `traveler` · `partner` · `staff` · `admin` (`moderator` is kept in the schema, unused by the portals).

## Portals in MVP (Partner 2.1–2.4, Admin 3.1–3.4, Staff 4.1–4.4)

| Portal | Menu page | Data management | API |
|--------|-----------|-----------------|-----|
| Partner 2.1 | Dashboard | Counters: listings by status, bookings to complete, this-month net | `GET /partner/v1/dashboard` |
| Partner 2.2 | Listings | Create (draft), edit, publish, archive own listings | `/partner/v1/listings` |
| Partner 2.3 | Bookings | Complete a confirmed booking; reject (refund) | `/partner/v1/bookings` |
| Partner 2.4 | Earnings & profile | Gross / commission / net per month; business profile | `/partner/v1/earnings`, `/profile` |
| Admin 3.1 | Dashboard | Platform counters, revenue | `GET /admin/v1/dashboard` |
| Admin 3.2 | Users | Search, change role (not partner), suspend / reactivate | `/admin/v1/users` |
| Admin 3.3 | Partners & listings | Approve / reject / suspend applications, commission, moderate listings | `/admin/v1/partners`, `/listings` |
| Admin 3.4 | Audit & system | Audit log of every portal mutation, dependency status | `/admin/v1/audit-logs`, `/system` |
| Staff 4.1 | Dashboard | Open tickets, unchecked bookings, checked / flagged today, avg first response | `GET /staff/v1/dashboard` |
| Staff 4.2 | Bookings | Verify or flag paid bookings (note required when flagged) | `/staff/v1/bookings` |
| Staff 4.3 | Tickets | Reply, resolve / reopen (replaces live chat in MVP) | `/staff/v1/tickets` |
| Staff 4.4 | Daily report | Day summary in Asia/Bangkok | `GET /staff/v1/reports/daily` |

Traveler side: apply to become a partner (`POST /v1/partners/apply`, `GET /v1/partners/me`), see ticket status and staff replies and answer them.
Partner lifecycle: pending → approved | rejected; approved ↔ suspended; rejected → approved. Suspended partners' listings leave the public catalog.
Commission defaults to 12 % (0–50 %, set by admin). The board's partner "ยืนยัน / ปฏิเสธ" buttons map to **complete / reject** (bookings are already confirmed by payment).
