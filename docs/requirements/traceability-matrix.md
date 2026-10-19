# Requirements traceability (MVP)

| REQ ID | Summary | Screens | API | DB entities | Test (Phase 04) |
|--------|---------|---------|-----|-------------|-----------------|
| REQ-AUTH-01 | Login / guest policy | S01 | `/auth/login` | users, user_sessions | UF-01 E2E |
| REQ-AUTH-02 | Registration profile | S02–S04 | `/auth/register` | user_profiles | Integration |
| REQ-AUTH-03 | Change password | S31 | `/auth/password` | users | Unit |
| REQ-HOM-01 | Home destinations | S05 | `/catalog/destinations` | destinations | Contract |
| REQ-SRH-01 | Search/filter | S06 | `/catalog/listings` | listings | Contract |
| REQ-SRH-02 | Listing detail | S07 | `/catalog/listings/{id}` | listings | E2E UF-02 |
| REQ-BKG-01 | Booking draft | S08 | `POST /bookings` | bookings | Integration |
| REQ-BKG-03 | Confirmation | S10 | bookings | bookings, payments | E2E |
| REQ-BKG-04 | My bookings | S27–S28 | `GET /bookings` | bookings | E2E |
| REQ-BKG-05 | Cancel + refund | S29 | cancel endpoint | wallet_transactions | E2E UF-04 |
| REQ-PAY-01 | Payment methods | S09 | payment intents | payments | Integration |
| REQ-PAY-02 | Insufficient wallet | S09 | 402/validation | wallet_accounts | Unit |
| REQ-WLT-01 | Wallet + top-up | S20 | `/wallet` | wallet_* | UF-05 |
| REQ-AI-01 | AI itinerary job | S11 | `/ai/itineraries` | ai_planner_jobs | Contract |
| REQ-AI-02 | Save trip | S11–S12 | `/trips/saved` | saved_trips | UF-03 |
| REQ-TRP-01 | Trip list | S12 | GET saved trips | saved_trips | — |
| REQ-COM-01 | Community feed | S14 | `/community/feed` | community_posts | UF-06 |
| REQ-COM-02 | Post detail / comments | S15 | posts, comments | community_* | Integration |
| REQ-NOT-01 | Notifications | S21 | `/notifications` | notifications | — |
| REQ-REV-02 | Write review | S23 | `POST /reviews` | reviews | UF-08 |
| REQ-CPN-01 | Coupons | S24 | `/coupons` | coupons | — |
| REQ-SUP-01 | FAQ + tickets | S25–S26 | support | support_tickets | UF-07 |
| REQ-PRF-02 | Settings / theme | S30 | `/me/settings` | user prefs JSON | Widget |

Classification today: requirements **documented**; implementation **NOT IMPLEMENTED** except prototype UI mock.
