# System context and boundaries

## Actors

| Actor | Description |
|-------|-------------|
| Traveler | End user: search, book, pay, plan trips, community |
| Guest | Browse limited catalog; must register to book (configurable) |
| Moderator | Review reported community content (phase 2) |
| Admin | Operations, coupons, listing onboarding (phase 2) |
| Payment provider | External PSP / PromptPay QR flow |
| LLM provider | Generates or enriches itineraries |
| FCM | Push delivery to Flutter |

## System boundary

**In scope (MVP parity with prototype)**

- Authentication and registration (multi-step profile)
- Home, search, listing detail, booking checkout, payment, confirmation
- My bookings, cancel + wallet refund
- AI trip planner (async job + poll; rule-based fallback in DEV mock)
- Saved trips and trip detail
- Wallet balance, top-up (mock in DEV; real PSP in prod)
- Notifications (in-app; push via FCM when configured)
- Reviews and coupons
- Help FAQ + support tickets
- Community: feed, Q&A filter, groups, posts, comments, follow, save, report, hide

**Out of scope (initial release unless added to requirements)**

- Native airline/hotel GDS live inventory (use merchant-managed listings)
- In-app map navigation SDK (MVP: external map deep links as in prototype)
- Real-time chat support (tickets only in MVP)
- Multi-language content CMS (UI strings i18n only; content Thai-first)

## External dependencies

| Integration | MVP usage | Failure mode |
|-------------|-----------|--------------|
| MySQL | System of record | API unhealthy; no write degradation |
| Redis | Sessions, rate limit, cache | Fall back to DB for session if configured; stricter rate limits |
| Object storage | Listing/post images | Placeholder images; block upload |
| Payment PSP | Card / PromptPay | User selects wallet or retry |
| LLM | Itinerary generation | Return rule-based itinerary + `degraded: true` |
| FCM | Push notifications | In-app notifications still work |

## Data classification

| Class | Examples | Controls |
|-------|----------|----------|
| Public | Destinations, listing marketing copy | CDN cache |
| Internal | Aggregated analytics | Access restricted |
| PII | Name, email, phone, DOB | Encrypt at rest where supported; minimize logs |
| Financial | Wallet ledger, payment refs | Strong audit trail; no PAN storage (PSP tokenization) |
| UGC | Community posts | Moderation, report workflow |
