# User flows

Flows trace to screens in [screen-inventory.md](screen-inventory.md).

## UF-01 — Guest / login → home

```mermaid
flowchart LR
  A[S01 Login] -->|credentials| B[S05 Home]
  A -->|guest| B
  A -->|register| C[S02-04 Register]
  C --> B
```

**API:** `POST /auth/login` or register; guest = no token, catalog read-only until book.

## UF-02 — Search and book listing

```mermaid
flowchart TD
  H[S05 Home] --> S[S06 Search]
  S --> D[S07 Listing detail]
  D --> B[S08 Booking summary]
  B --> P[S09 Payment]
  P --> C[S10 Confirmation]
  C --> H
  C --> BK[S27 Bookings]
```

**API sequence**

1. `GET /catalog/listings?q=&type=`
2. `GET /catalog/listings/{id}`
3. `POST /bookings` (+ optional coupon validation)
4. `POST /payments/intents` → poll or webhook
5. `GET /bookings/{id}`

**Edge cases:** wallet insufficient (REQ-PAY-02), coupon invalid, payment failed retry.

## UF-03 — AI planner → save → book trip package

```mermaid
flowchart TD
  T[S11 AI Planner] -->|generate| R[Itinerary result]
  R -->|save| TR[S12 Trips]
  R -->|book| B[S08 Booking]
  TR --> TD[S13 Trip detail]
```

**API**

1. `POST /ai/itineraries` → `202`
2. `GET /ai/itineraries/jobs/{id}` until completed
3. `POST /trips/saved`
4. Booking uses `tripPackage` on `POST /bookings`

## UF-04 — Cancel booking → wallet refund

```mermaid
flowchart LR
  BK[S27 Bookings] --> BD[S28 Booking detail]
  BD --> CN[S29 Cancel]
  CN --> BK
```

**API:** `POST /bookings/{id}/cancel` → wallet credit + notifications.

## UF-05 — Wallet top-up

```mermaid
flowchart LR
  PR[S19 Profile] --> W[S20 Wallet]
  W -->|top up| W
```

**API:** `POST /wallet/top-ups` (mock instant in DEV; QR flow in prod).

## UF-06 — Community: consume → engage → copy trip

```mermaid
flowchart TD
  CM[S14 Community] --> PD[S15 Post detail]
  CM --> GD[S18 Group detail]
  CM --> CP[S16 Create post]
  PD -->|copy trip| TR[S12 Trips]
  PD -->|report/hide| CM
```

**API:** feed, post detail, comments, like/save/follow/report, `POST .../copy-trip`.

## UF-07 — Support

```mermaid
flowchart LR
  PR[S19 Profile] --> HP[S25 Help]
  HP --> ST[S26 Support ticket]
```

**API:** `GET /support/faqs`, `POST /support/tickets`.

## UF-08 — Reviews after completed booking

```mermaid
flowchart LR
  BD[S28 Booking detail] --> WR[S23 Write review]
  WR --> RV[S22 Reviews]
```

**API:** `POST /reviews` linked to `listingId` from booking line item.

## Permission matrix (MVP)

| Action | Guest | Traveler |
|--------|-------|----------|
| Browse catalog | Yes | Yes |
| Book / pay | No | Yes |
| Wallet | No | Yes |
| Community post | No | Yes |
| Settings / delete account | No | Yes |

Server enforces; guest mode in prototype should map to read-only API in Flutter.
