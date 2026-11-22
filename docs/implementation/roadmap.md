# TEP (Sabai jai Travel) — Audit และ Implementation Roadmap

วันที่ audit: 2026-11-22 · ขอบเขต: อ่านอย่างเดียว (ไม่ลบ/เขียนทับข้อมูล ไม่ต่อ production ไม่ deploy)
อ้างอิง: Project Instructions "Professional Local-First Full-Stack Development Standard" (14 หัวข้อ)

## 1. สถานะ Repository

| รายการ | ค่าจริง |
|---|---|
| Repo / branch | `Bas-travel/TEP` · `feat/phase-03-sprint-1` @ `9d320f6` (ตรงกับ remote) |
| PR #2 | เปิดอยู่ · `mergeable_state = clean` · CI ผ่านทั้ง 3 jobs |
| งานที่ยังไม่ commit (local) | 8 ไฟล์แก้ + 5 ไฟล์ใหม่ = **demo dataset 20–30 รายการ/โมดูล** (ทดสอบแล้วบน MySQL 8 จริง) |
| Commit ที่ GitHub ขึ้น Unverified | 3 commit ของผู้ใช้ (ไม่กระทบโค้ด; แก้ด้วย Squash merge หรือ commit signing) |

## 2. Technology Stack (ตรวจจากโค้ดจริง)

| ชั้น | เทคโนโลยี | ขนาด/หมายเหตุ |
|---|---|---|
| Backend | Go 1.26 `net/http` ล้วน (ไม่มี framework) | 9.4k LOC + 1.7k LOC เทสต์ · packages: api, auth, config, db, httpx, objstore |
| Database | MySQL 8.0 · migrations 0001–0006 (up/down) | repo 2 แบบ: in-memory + SQL รัน test-suite เดียวกัน |
| Cache/Queue | Redis 7 (health เท่านั้น) | ยังไม่ใช้ rate-limit/session store จริง |
| Object storage | MinIO (S3 presigned URL) | อัปโหลดรูป post ใช้ได้; listing image ยังไม่ต่อ |
| Mobile/Web | Flutter (mock + http ApiClient) | 47 test cases · ไม่มี android/ios/web ใน repo (สร้างด้วย `flutter create .` ในเครื่อง) |
| Auth | HS256 JWT + PBKDF2 (120k) · role อ่านจาก DB ทุก request · refresh-token rotation/reuse | stdlib ล้วน |
| Payment | Mock PSP (`MOCK_PAYMENT`) + webhook HMAC-SHA256 + idempotency + dedupe event | ยังไม่มี gateway sandbox จริง |
| CI | GitHub Actions: Go vet/race/test (+MySQL), OpenAPI parity (90 ops), Flutter analyze/test, contract test | |
| Docker | compose: mysql, redis, minio, api · **ยังไม่เคยยืนยันบน Docker daemon จริง** | sandbox นี้มี docker CLI แต่ไม่มี daemon |

## 3. API และ Modules/Menus ที่มีจริง

OpenAPI **90 operations** (routes = spec ครบ ตรวจอัตโนมัติ): Health 2 · Auth 5 · Profile 4 · Catalog 3 · Bookings 4 · Payments 2 · Wallet 2 · AI Planner 2 · Trips 3 · Reviews 2 · Coupons 2 · Notifications 3 · Support 4 · Community 15 · Media 1 · Devices 1 · Webhooks 1 · Partners 2 · **Partner portal 12 · Admin portal 11 · Staff portal 9**

เมนูใน UI (ทำงานจริงกับ API):
- Traveler ≈ 18: หน้าแรก, ค้นหา, การจอง, ทริป, ตั๋ว/พาส, Wallet, คูปอง, แจ้งเตือน, ข้อความ, โปรไฟล์, ตั้งค่า, ช่วยเหลือ, AI planner, ชุมชน, ไกด์เมือง, รีวิว, สมัครพาร์ทเนอร์, แผนที่/อากาศ (deferred)
- Partner 4: ภาพรวม, สินค้า/บริการ, การจอง, รายได้
- Admin 4 (+แท็บย่อย): ภาพรวม, ผู้ใช้, พาร์ทเนอร์ (+ตรวจรายการ), ระบบ (+audit log)
- Staff 4: ภาพรวม, ตรวจการจอง, ช่วยเหลือ (tickets), รายงาน

**รวม ≈ 30 เมนูหลัก** — อยู่ในช่วงเป้า 30–40 อยู่แล้ว ช่องว่างคือ "ความลึก" ของ back-office (ดู §5)

## 4. ข้อมูลจริงใน MySQL (หลัง demo dataset, ตรวจผ่าน HTTP + SQL)

users 40 · partners 9 · listings 30 · destinations 8 · bookings 33 · payments 28 · wallet tx 33+ · coupons 20 · reviews 30 · saved trips 26 · notifications 34 · tickets 29 (+ replies 27) · groups 10 · posts 35 · comments 60 · likes 231 · follows 118 · audit logs 38 · faqs 20
บัญชีเดโม: `demo/admin/staff/staff2/partner/partner2–6/moderator@tep.dev` รหัส `Passw0rd!` (เฉพาะ DEV)

## 5. ผลตรวจ vs มาตรฐาน (Gap Analysis)

| # | หัวข้อมาตรฐาน | สถานะ | หลักฐาน / ช่องว่าง |
|---|---|---|---|
| 1 | Core principles (reuse, no duplicate) | ✅ | สถาปัตยกรรม Repo+Tx เดียว, ไม่มีตารางซ้ำ |
| 2 | Local env + แยก LOCAL/TEST/STAGING/PROD | 🟡 | มี `.env.example` ชุดเดียว; ยังไม่มี `.env.test/.staging` template, compose override |
| 3 | MySQL: schema, FK, index, migration, seed, tx | ✅ | 6 migrations + down; seed + dataset; tx ผ่าน `Repo.Tx` |
| 3 | Backup/Restore procedure | ❌ | ไม่มีสคริปต์/ runbook backup–restore |
| 4 | โมดูล 30–40 | 🟡 | ≈30 เมนู; ขาด: เส้นทาง CRUD ฝั่ง admin (coupons, FAQ, destinations), รายงาน/analytics, payments/refund console, moderation queue (reports), settings แก้ไขได้, roles & permissions UI, moderator portal |
| 5 | FE→API→DB จริง | ✅/🟡 | โหมด http ครบ; โหมด mock ฝั่ง traveler ใช้ catalog คงที่ |
| 6 | Payment sandbox | 🟡 | Mock PSP + webhook HMAC + idempotency ✅; **ยังไม่มี PSP sandbox จริง** (Omise/Stripe/2C2P test mode), ยังไม่มี reconciliation job |
| 7 | Security | 🟡 | ✅ RBAC server-side, PBKDF2, JWT, rate limit, CORS, upload type/size, audit log, SQL parameterized. ❌ ยังไม่ทำ: govulncheck/gosec, CSRF (ไม่จำเป็นกับ bearer token แต่ต้องบันทึกเหตุผล), secure token storage ฝั่งมือถือ, security headers ตรวจสอบ, rate limit ผ่าน Redis (ตอนนี้ in-process) |
| 8 | UI/UX responsive, states | 🟡 | มี loading/empty/error ในพอร์ทัล; ยังไม่ตรวจ responsive บน browser/อุปกรณ์จริง |
| 9 | QA | 🟡 | Go coverage วัดจริง (in-memory): **api 62.8% · auth 74.5% · config 75.8% · db 17.1% · objstore 97.4% · httpx/cmd 0%** → เป้า ≥80% ยังไม่ถึง. ไม่มี E2E, performance, security scan |
| 10 | Workflow phases | — | ดู roadmap §6 |
| 11 | Git | ✅ | ไม่ force push, branch ชัดเจน, ไม่มี secret ใน git |
| 12 | Production restrictions | ✅ | ไม่มี deploy; config ปฏิเสธ placeholder/mock payment บน production |
| 13 | Docs ตรงความจริง | ❌ | `docs/KNOWN_ISSUES.md`, `DEVELOPMENT_PROGRESS.md`, `TEST_REPORT.md`, `SECURITY_AUDIT.md` **ล้าสมัย** (ยังบอกว่าพอร์ทัลเป็น preview, ยังไม่รัน Docker/Flutter) |

### ความเสี่ยงและสิ่งที่ยืนยันไม่ได้ใน sandbox
- Docker daemon ไม่มี → `docker compose up` ต้องยืนยันบนเครื่องคุณ
- Flutter SDK ไม่มี → ยืนยันได้ผ่าน CI เท่านั้น (CI ปัจจุบันตั้ง `--no-fatal-infos` ชั่วคราวเพราะ Flutter รุ่นใหม่ประกาศ `material.dart` deprecated)
- Go proxy (`proxy.golang.org`) ถูกบล็อกใน sandbox; ใช้ไดรเวอร์ MySQL จากแพ็กเกจ apt (v1.7.1 ตรงกับ go.mod) เพื่อทดสอบกับ MySQL จริง

## 6. Implementation Roadmap (เสนอ — รออนุมัติ)

หลักการ: ทุก Phase ทำ local → ทดสอบจริง (MySQL จริง) → commit ในเครื่อง → push รวมตอนคุณสั่ง. ไม่ลบข้อมูลเดิม; schema เปลี่ยนผ่าน migration ใหม่เท่านั้น.

| Phase | งาน | เกณฑ์ผ่าน (Exit criteria) |
|---|---|---|
| **P0 Baseline** | commit demo dataset (local), อัปเดต 4 เอกสารที่ล้าสมัยให้ตรงความจริง, สร้าง `.env.test/.env.staging.example` + `docker-compose.override` สำหรับ LOCAL/TEST | git สะอาด · เอกสารไม่ขัดกับโค้ด |
| **P1 Local stack จริง** | ยืนยัน `docker compose up` บนเครื่องคุณ (คุณรัน/ฉันแก้ตาม log), สคริปต์ backup/restore MySQL + runbook + ทดสอบ restore | health ready ทั้ง mysql/redis/minio · restore ได้จริง |
| **P2 Back-office ให้ครบ ~36 เมนู** | Admin: coupons CRUD, FAQ/CMS, destinations CRUD, moderation queue (post reports), payments/refund console, รายงาน/analytics (รายได้, conversion), system settings; Partner: อัปโหลดรูป listing (MinIO); Moderator portal; Roles & permissions (อ่าน/อธิบายสิทธิ์) | ทุกเมนูมี CRUD/ค้นหา/แบ่งหน้า/สถานะ · RBAC matrix test ครอบคลุมทุก endpoint ใหม่ · OpenAPI parity |
| **P3 Payment Sandbox** | adapter PSP (เลือกเจ้า: Omise/Stripe/2C2P test mode), webhook signature จริง, refund workflow, reconciliation job + ตารางบันทึก | E2E จ่าย→webhook→confirm→refund ผ่านใน test mode · ไม่มีคีย์ live |
| **P4 Security hardening** | govulncheck + gosec ใน CI, Redis rate-limit, security headers, secure token storage (flutter_secure_storage), refresh-reuse test, secret scan, เอกสาร threat model | ไม่มี critical/high ที่ยังเปิด · รายงานผลสแกนจริง |
| **P5 QA เข้ม** | เพิ่ม unit/integration เพื่อ **coverage ≥80% ของ business logic** (api/db/httpx), E2E (flutter integration_test บน chrome + Playwright สำหรับ API flow), performance (k6/vegeta baseline), regression ใน CI | coverage gate ใน CI · E2E flow สำคัญผ่านทั้งหมด |
| **P6 Staging** | `docker-compose.staging.yml`/Terraform-ready config, env แยก, deploy script (ยังไม่ deploy), smoke test | พร้อม deploy เมื่อคุณอนุมัติเท่านั้น |
| **P7 Production readiness** | เอกสาร deploy/rollback/monitoring, release checklist ครบ | **ไม่ deploy จนกว่าคุณสั่ง** |

ลำดับแนะนำ: P0 → P1 → P2 → P3 → P4 → P5 → P6 (P7 เป็นเอกสารเท่านั้น)

## 7. สิ่งที่ต้องการให้คุณตัดสินใจ
1. อนุมัติเริ่ม **P0** (commit ในเครื่อง + แก้เอกสาร + env templates)?
2. PSP sandbox ที่ต้องการ: Omise / Stripe / 2C2P (กระทบ P3)
3. ยืนยันขอบเขต P2: รวม **payout พาร์ทเนอร์** หรือไม่ (ตอนนี้ `mvp-scope.md` ระบุ post-MVP — ถ้าจะรวมต้องแก้สเปกก่อน)
4. จะ merge PR #2 ก่อน (แนะนำ Squash and merge) แล้วเปิด branch ใหม่ต่อ หรือทำต่อบน branch เดิม?
