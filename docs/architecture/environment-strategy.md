# Environment strategy (DEV → BETA → PRODUCTION)

| | DEV | BETA | PRODUCTION |
|---|-----|------|------------|
| **Purpose** | Fast UI + API iteration | UAT / stakeholder preview | Live users |
| **Flutter API mode** | `mock` or local Gin | `beta` | `production` |
| **Database** | Local MySQL or none (mock) | Dedicated MySQL instance | Managed MySQL + replicas |
| **Redis** | Local / optional | Dedicated | Cluster |
| **Object storage** | MinIO | Beta bucket | Prod bucket + CDN |
| **Payment** | Mock adapter | PSP sandbox | PSP live |
| **LLM** | Mock / template | Beta key, spend cap | Production key |
| **Customer data** | Synthetic fixtures only | Synthetic + test accounts | Real PII |
| **Deploy** | Local / optional preview branch | CI deploy on `beta` tag | CI deploy on release tag |

## Release gates (summary)

See [../qa/release-checklist.md](../qa/release-checklist.md).

## Configuration validation

API process fails startup if:

- `APP_ENV=production` and mock payment adapter enabled
- JWT secret is default placeholder
- Database SSL mode disabled in production
