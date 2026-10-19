# Screen inventory — prototype vs TEP (Flutter)

Reference prototype: [sabai-travel-app-prototype.jsx](../../sabai-travel-app-prototype.jsx)  
API mapping: [openapi.yaml](../api/openapi.yaml)

## Design tokens (preserve in Flutter)

| Token | Light | Usage |
|-------|-------|--------|
| `--pine` | #10403D | Primary actions, active tab |
| `--amber` | #E3A23C | Accents, AI highlights |
| `--paper` / `--surface` | #FAFAF8 / #FFFFFF | Backgrounds |
| Display font | Fraunces | Headlines |
| UI font | Plus Jakarta Sans | Body |

Dark theme: `.theme-dark` CSS block → Flutter `ThemeData` extensions.

## Screen catalog

| ID | Screen (TH) | Prototype route | Primary APIs | REQ | Flutter priority |
|----|-------------|-----------------|--------------|-----|------------------|
| S01 | เข้าสู่ระบบ | `login` | `POST /auth/login` | REQ-AUTH-01 | P0 |
| S02–S04 | สมัครสมาชิก 1–3 | `register-*` | `POST /auth/register` | REQ-AUTH-02 | P0 |
| S05 | หน้าแรก | `home` | `GET /catalog/destinations`, notifications count | REQ-HOM-01 | P0 |
| S06 | ค้นหา | `search` | `GET /catalog/listings` | REQ-SRH-01 | P0 |
| S07 | รายละเอียด listing | `listing-detail` | `GET /catalog/listings/{id}` | REQ-SRH-02 | P0 |
| S08 | สรุปการจอง | `booking` | `POST /bookings` (draft) | REQ-BKG-01 | P0 |
| S09 | ชำระเงิน | `payment` | `POST /payments/intents` | REQ-PAY-01 | P0 |
| S10 | ยืนยันการจอง | `confirmation` | `GET /bookings/{id}` | REQ-BKG-03 | P0 |
| S11 | AI วางแผนทริป | `ai` | `POST /ai/itineraries`, poll job | REQ-AI-01 | P0 |
| S12 | ทริปของฉัน | `trips` | `GET /trips/saved` | REQ-TRP-01 | P0 |
| S13 | รายละเอียดทริป | `trip-detail` | `GET /trips/saved/{id}` | REQ-TRP-02 | P1 |
| S14 | ชุมชน (feed/QA/groups) | `community` | `GET /community/feed`, `/groups` | REQ-COM-01 | P1 |
| S15 | รายละเอียดโพสต์ | `post-detail` | `GET /community/posts/{id}` | REQ-COM-02 | P1 |
| S16 | สร้างโพสต์ | `create-post` | `POST /community/posts` | REQ-COM-03 | P1 |
| S17 | โปรไฟล์สมาชิก | `member` | `GET /community/members/{id}` | REQ-COM-04 | P1 |
| S18 | รายละเอียดกลุ่ม | `group-detail` | `GET /community/groups/{id}` | REQ-COM-05 | P1 |
| S19 | โปรไฟล์ | `profile` | `GET /me` | REQ-PRF-01 | P0 |
| S20 | Sabai Wallet | `wallet` | `GET /wallet`, `POST /wallet/top-ups` | REQ-WLT-01 | P0 |
| S21 | การแจ้งเตือน | `notifications` | `GET /notifications`, mark read | REQ-NOT-01 | P1 |
| S22 | รีวิวของฉัน | `reviews` | `GET /reviews` | REQ-REV-01 | P1 |
| S23 | เขียนรีวิว | `write-review` | `POST /reviews` | REQ-REV-02 | P1 |
| S24 | คูปอง | `coupons` | `GET /coupons`, apply | REQ-CPN-01 | P1 |
| S25 | ช่วยเหลือ | `help` | `GET /support/faqs`, tickets list | REQ-SUP-01 | P1 |
| S26 | ตั๋วสนับสนุน | `support-ticket` | `POST /support/tickets` | REQ-SUP-02 | P2 |
| S27 | การจองของฉัน | `bookings` | `GET /bookings` | REQ-BKG-04 | P0 |
| S28 | รายละเอียดการจอง | `booking-detail` | `GET /bookings/{id}` | REQ-BKG-04 | P0 |
| S29 | ยกเลิกการจอง | `cancel-booking` | `POST /bookings/{id}/cancel` | REQ-BKG-05 | P0 |
| S30 | ตั้งค่า | `settings` | `GET/PATCH /me/settings` | REQ-PRF-02 | P1 |
| S31 | เปลี่ยนรหัสผ่าน | `change-password` | `PUT /auth/password` | REQ-AUTH-03 | P1 |

## Navigation model

| Pattern | Prototype | Flutter recommendation |
|---------|-----------|---------------------------|
| Main tabs | 6 tabs (home, search, ai, trips, community, profile) | `StatefulShellRoute` / bottom nav |
| Subflows | Stack `push/pop` | Nested `Navigator` or go_router stacks |
| Modals | Community bottom sheet | `showModalBottomSheet` |
| Toast | Ephemeral message | `SnackBar` / overlay |

## States per screen (acceptance)

Each screen must implement: **loading**, **empty**, **success**, **error**, **permission-denied** (where auth required). Prototype implements loading/error partially (payment, AI, wallet insufficient) — Flutter must complete the matrix.

## Gap analysis (prototype → production)

| Area | Prototype | Target |
|------|-----------|--------|
| Auth | UI only | Real JWT + secure storage |
| Data | In-memory | API + local cache |
| AI | setTimeout + rules | Async job + LLM with fallback |
| Payment | Simulated delay | PSP + webhook |
| Community | Local state | Server moderation + pagination |
| Images | Placeholder gradients | Object storage URLs |
| i18n | Thai only | ARB files; EN phase 2 optional |

Status: **UX parity PASS (visual/flow)** · **Backend integration NOT IMPLEMENTED**
