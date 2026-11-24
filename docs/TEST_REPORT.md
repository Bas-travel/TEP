# Test Report

_Last updated: 2026-11-22. Evidence below was produced by commands actually run; nothing is marked PASS otherwise._

## Backend (Go 1.26)
| Check | Result |
|---|---|
| `go vet ./...` | pass |
| `go test ./...` (in-memory repository) | pass — all packages |
| Same suite on **MySQL 8.0.46** (`TEP_TEST_MYSQL_DSN`, repository `API_MODE=db`) | pass — all packages |
| `python3 scripts/openapi_parity.py` | 90 routes == 90 spec operations |
| `python3 scripts/gen_seed.py --check` | no drift |
| Demo dataset test (`TestDemoDatasetCoversEveryMenu`) | pass in memory and on MySQL |
| Real binary + MySQL: base seed -> restart with dataset (top-up) -> restart again (idempotent) | verified (40 users / 33 bookings, unchanged on 3rd start) |
| MySQL backup -> restore round trip (`scripts/db-backup.sh`/`db-restore.sh`) | 39 tables / 1,633 rows identical; refusal guards verified |

### Coverage (statements, in-memory run)
api 62.8% · auth 74.5% · config 75.8% · db 17.1% · objstore 97.4% · httpx 0% · cmd 0% — **below the 80% target** (see roadmap P5).

## Mobile (Flutter) — CI only
`flutter analyze` (infos non-fatal), 47 test cases (contract tests skipped locally, run in the `contract` CI job against the Go API). Not run on a device/emulator.

## Not verified
Docker Compose on a daemon, PowerShell scripts, device UI, E2E, performance, security scans (govulncheck/gosec), payment gateway sandbox.
