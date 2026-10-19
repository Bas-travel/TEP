# Data dictionary (selected entities)

Requirement IDs trace to [../ux/screen-inventory.md](../ux/screen-inventory.md).

## users / user_profiles

| Field | Description | REQ |
|-------|-------------|-----|
| email, password_hash | Login credentials | REQ-AUTH-01 |
| profile.* | Registration steps 1–3 | REQ-AUTH-02 |
| travel_styles (JSON) | Chip selections | REQ-AUTH-02 |
| budget_level, travel_with | Personalization | REQ-AUTH-02, REQ-AI-01 |

## listings

| Field | Description | REQ |
|-------|-------------|-----|
| type | hotel, restaurant, flight, attraction | REQ-SRH-01 |
| price_from | Starting price THB | REQ-SRH-02, REQ-BKG-01 |
| rating_avg, review_count | Display metrics | REQ-SRH-02 |

## bookings / payments

| Field | Description | REQ |
|-------|-------------|-----|
| booking_number | Human-readable e.g. SB###### | REQ-BKG-03 |
| service_fee | Prototype uses ~5% | REQ-PAY-01 |
| status | confirmed, completed, cancelled | REQ-BKG-04 |
| refund_amount | Wallet refund on cancel | REQ-BKG-05 |

## wallet_accounts / wallet_transactions

| Field | Description | REQ |
|-------|-------------|-----|
| balance | Current THB balance | REQ-WLT-01 |
| direction | in / out | REQ-WLT-02 |
| reference_* | Link to booking/payment | REQ-WLT-03 |

## saved_trips / ai_planner_jobs

| Field | Description | REQ |
|-------|-------------|-----|
| budget_breakdown | JSON array label/value | REQ-AI-02 |
| trip_days | Normalized itinerary | REQ-AI-02, REQ-TRP-01 |
| ai_planner_jobs.result | LLM or template output | REQ-AI-01 |

## community_posts

| Field | Description | REQ |
|-------|-------------|-----|
| post_type | story, tip, question, trip | REQ-COM-01 |
| shared_trip | JSON days/budget/styles for trip posts | REQ-COM-02 |
| group_id | Nullable for global feed | REQ-COM-03 |

## coupons / user_coupons

| status | available, used, expired | REQ-CPN-01 |

## notifications

| type | booking, payment, promotion, trip, system, community | REQ-NOT-01 |

## support_tickets

| subject, message | User-submitted help | REQ-SUP-01 |
