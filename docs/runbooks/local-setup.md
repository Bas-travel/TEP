# Local setup (Docker stack)

Runs MySQL 8, Redis 7, MinIO and the Go API with persistence (`API_MODE=db`) and demo data.

## 1. Prerequisites
Docker Desktop (Compose v2) running. On Windows use PowerShell.

## 2. Create `.env` (once)
Secrets are generated randomly — never type or reuse them, never commit `.env` (it is git-ignored).

| OS | Command |
|----|---------|
| Windows | `powershell -ExecutionPolicy Bypass -File scripts\setup-env.ps1` |
| macOS / Linux | `scripts/setup-env.sh` |

The script refuses to overwrite an existing `.env` (add `-Force` / `--force` to regenerate). It fills `JWT_SECRET` (64 chars), `MYSQL_PASSWORD`, `MYSQL_ROOT_PASSWORD`, `MINIO_ROOT_USER`, `MINIO_ROOT_PASSWORD` and keeps the rest from `.env.example`:

| Variable | Dev value | Meaning |
|----------|-----------|---------|
| `API_MODE` | `db` | `db` = MySQL, `mock` = in-memory (no DB) |
| `MIGRATE_ON_START` | `true` | applies `services/api-go/migrations` on start |
| `MIGRATE_SEED` | `true` | creates synthetic demo data and demo accounts; ignored when `APP_ENV=production` |
| `MOCK_PAYMENT` | `true` | fake payment provider; **must be `false` in production** |

## 3. Start and verify
```powershell
docker compose up --build -d
docker compose ps                                  # all services healthy
curl.exe http://localhost:8080/health/ready        # {"status":"ok", ... "mysql":"ok", "redis":"ok"}
```
Demo accounts (password `Passw0rd!`): `demo@tep.dev` (traveler), `partner@tep.dev`, `staff@tep.dev`, `admin@tep.dev`.

## 4. Run the mobile app against it
```powershell
cd apps\mobile-flutter
flutter create .        # once: generates windows/web/android platform folders
flutter run -d chrome --dart-define=API_MODE=http --dart-define=API_BASE_URL=http://localhost:8080
```
Without the `--dart-define` flags the app runs in mock mode (no backend).

## 5. Everyday commands
| Task | Command |
|------|---------|
| Logs | `docker compose logs -f api` |
| Stop | `docker compose down` |
| Reset database and files (destroys data) | `docker compose down -v` |
| Change passwords | `-Force` regenerate **and** `docker compose down -v` (MySQL keeps the password it was first created with) |

## 6. Rules that keep it safe
- Real environments (beta / production) never use `.env` from this script: secrets come from the host's secret store; the API refuses placeholder secrets and `MIGRATE_SEED` is ignored in production (see `docs/architecture/environment-strategy.md`).
- Secrets are alphanumeric so they are safe inside the MySQL DSN and compose interpolation; do not add `$`, `#` or spaces if you edit by hand.
- Ports (`API_PORT`, `MYSQL_PORT` (default 3307), `REDIS_PORT`, `MINIO_PORT`) can be changed in `.env` if they clash with local services.

## Troubleshooting
| Symptom | Cause / fix |
|---------|-------------|
| `required variable MYSQL_PASSWORD is missing` | `.env` not created — run step 2 |
| `Access denied for user 'tep'` after regenerating `.env` | old volume has old password — `docker compose down -v` |
| API restarts, log says `JWT_SECRET` | value shorter than 32 chars or still a placeholder — regenerate |
| `No supported devices` in Flutter | run `flutter create .` (step 4) |
