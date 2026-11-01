# Database Design

Authoritative logical ERD: `docs/database/erd.md`.

Selected migrations:

- `0001` initial travel entities and MVP additions
- `0002` DEV seed data
- `0003` booking trip details
- `0004` persistence and engagement additions

Critical rules include numeric booking IDs, opaque async/payment IDs, decimal monetary values, UTC timestamps, refresh-token hash storage, and transactional booking/wallet operations.

Migration and seed verification must be repeated against MySQL 8 before Phase completion.
