package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

// sqlRepo is the MySQL 8 implementation of Repo (API_MODE=db). Schema: migrations/0001…0004.
// Times are stored as UTC DATETIME; money as DECIMAL(12,2). Wallet, booking and intent rows are locked with
// SELECT … FOR UPDATE inside the transaction whenever a handler asks for lock=true.
type sqlRepo struct{ db *sql.DB }

// NewSQLRepo wraps an opened *sql.DB (see internal/db.Open).
func NewSQLRepo(d *sql.DB) Repo { return &sqlRepo{db: d} }

func (r *sqlRepo) Close() error { return r.db.Close() }

func (r *sqlRepo) Tx(ctx context.Context, fn func(Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(&sqlTx{tx: tx, ctx: ctx}); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

type sqlTx struct {
	tx  *sql.Tx
	ctx context.Context
}

func (t *sqlTx) exec(q string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(t.ctx, q, args...)
}
func (t *sqlTx) query(q string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(t.ctx, q, args...)
}
func (t *sqlTx) row(q string, args ...any) *sql.Row { return t.tx.QueryRowContext(t.ctx, q, args...) }

func isDup(err error) bool {
	var me *mysql.MySQLError
	return errors.As(err, &me) && me.Number == 1062
}

func notFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func utc(t time.Time) time.Time { return t.UTC() }

func lockClause(lock bool) string {
	if lock {
		return " FOR UPDATE"
	}
	return ""
}

func jsonOrNil(v any, empty bool) (any, error) {
	if empty {
		return nil, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// ---- users & sessions ----

const userSelect = `SELECT u.id, u.email, u.phone, u.password_hash, u.role, u.status, u.handle,
  p.first_name, p.last_name, p.gender, DATE_FORMAT(p.date_of_birth,'%Y-%m-%d'), p.nationality, p.language, p.budget_level,
  p.travel_styles, p.favorite_destination, p.travel_with, p.level, p.bio
  FROM users u JOIN user_profiles p ON p.user_id = u.id`

func scanUser(row interface{ Scan(...any) error }) (*userRec, error) {
	var u userRec
	var phone, gender, dob, nat, lang, budget, fav, with, bio sql.NullString
	var styles []byte
	if err := row.Scan(&u.ID, &u.Profile.Email, &phone, &u.PassHash, &u.Role, &u.Status, &u.Handle,
		&u.Profile.FirstName, &u.Profile.LastName, &gender, &dob, &nat, &lang, &budget, &styles, &fav, &with, &u.Profile.Level, &bio); err != nil {
		return nil, notFound(err)
	}
	p := &u.Profile
	p.ID, p.Phone, p.Gender, p.DateOfBirth, p.Nationality = u.ID, phone.String, gender.String, dob.String, nat.String
	p.Language, p.BudgetLevel, p.FavoriteDestination, p.TravelWith = lang.String, budget.String, fav.String, with.String
	u.Bio = bio.String
	if len(styles) > 0 {
		_ = json.Unmarshal(styles, &p.TravelStyles)
	}
	return &u, nil
}

func profileArgs(p UserProfile) ([]any, error) {
	styles, err := jsonOrNil(p.TravelStyles, len(p.TravelStyles) == 0)
	if err != nil {
		return nil, err
	}
	return []any{p.FirstName, p.LastName, nullStr(p.Gender), nullStr(p.DateOfBirth), nullStr(p.Nationality), nullStr(p.Language),
		nullStr(p.BudgetLevel), styles, nullStr(p.FavoriteDestination), nullStr(p.TravelWith)}, nil
}

func (t *sqlTx) CreateUser(u *userRec, now time.Time) error {
	if u.Status == "" {
		u.Status = "active"
	}
	if u.Role == "" {
		u.Role = "traveler"
	}
	if u.Profile.Level == "" {
		u.Profile.Level = "Explorer"
	}
	handle := u.Handle
	if handle == "" {
		handle = "tmp_" + randID("") // replaced below once the id is known
	}
	res, err := t.exec(`INSERT INTO users (email, phone, handle, password_hash, role, status, created_at) VALUES (?,?,?,?,?,?,?)`,
		u.Profile.Email, nullStr(u.Profile.Phone), handle, u.PassHash, u.Role, u.Status, utc(now))
	if err != nil {
		if isDup(err) {
			return ErrDuplicate
		}
		return err
	}
	id, _ := res.LastInsertId()
	if u.Handle == "" {
		u.Handle = fmt.Sprintf("m_%d", id)
		if _, err := t.exec(`UPDATE users SET handle=? WHERE id=?`, u.Handle, id); err != nil {
			return err
		}
	}
	args, err := profileArgs(u.Profile)
	if err != nil {
		return err
	}
	args = append([]any{id}, args...)
	args = append(args, u.Profile.Level, nullStr(u.Bio))
	if _, err := t.exec(`INSERT INTO user_profiles (user_id, first_name, last_name, gender, date_of_birth, nationality, language,
		budget_level, travel_styles, favorite_destination, travel_with, level, bio) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`, args...); err != nil {
		return err
	}
	if _, err := t.exec(`INSERT INTO wallet_accounts (user_id) VALUES (?)`, id); err != nil {
		return err
	}
	u.ID, u.Profile.ID = id, id
	return nil
}

func (t *sqlTx) UserByID(id int64, lock bool) (*userRec, error) {
	return scanUser(t.row(userSelect+` WHERE u.id=? AND u.status='active'`+lockClause(lock), id))
}

func (t *sqlTx) UserByEmail(email string) (*userRec, error) {
	return scanUser(t.row(userSelect+` WHERE u.email=? AND u.status='active'`, strings.TrimSpace(email)))
}

func (t *sqlTx) SaveUser(u *userRec) error {
	if _, err := t.exec(`UPDATE users SET password_hash=?, phone=? WHERE id=?`, u.PassHash, nullStr(u.Profile.Phone), u.ID); err != nil {
		if isDup(err) {
			return ErrDuplicate
		}
		return err
	}
	args, err := profileArgs(u.Profile)
	if err != nil {
		return err
	}
	args = append(args, u.Profile.Level, nullStr(u.Bio), u.ID)
	_, err = t.exec(`UPDATE user_profiles SET first_name=?, last_name=?, gender=?, date_of_birth=?, nationality=?, language=?, budget_level=?,
		travel_styles=?, favorite_destination=?, travel_with=?, level=?, bio=? WHERE user_id=?`, args...)
	return err
}

func (t *sqlTx) DeleteUser(id int64, now time.Time) error {
	if _, err := t.UserByID(id, true); err != nil {
		return err
	}
	// Personal and engagement data goes; bookings, payments and wallet history stay (financial records) but anonymised.
	for _, q := range []string{
		`DELETE FROM user_sessions WHERE user_id=?`,
		`DELETE FROM device_push_tokens WHERE user_id=?`,
		`DELETE FROM user_settings WHERE user_id=?`,
		`DELETE FROM user_coupons WHERE user_id=?`,
		`DELETE FROM notifications WHERE user_id=?`,
		`DELETE FROM saved_trips WHERE user_id=?`,
		`DELETE FROM ai_planner_jobs WHERE user_id=?`,
		`DELETE FROM reviews WHERE user_id=?`,
		`DELETE FROM support_tickets WHERE user_id=?`,
		`DELETE FROM community_posts WHERE author_id=?`,
		`DELETE FROM community_comments WHERE author_id=?`,
		`DELETE FROM community_group_members WHERE user_id=?`,
		`DELETE FROM user_follows WHERE follower_id=? OR followee_id=?`,
		`DELETE FROM post_likes WHERE user_id=?`,
		`DELETE FROM post_saves WHERE user_id=?`,
		`DELETE FROM post_hides WHERE user_id=?`,
		`DELETE FROM post_reports WHERE user_id=?`,
	} {
		args := []any{id}
		if strings.Count(q, "?") == 2 {
			args = append(args, id)
		}
		if _, err := t.exec(q, args...); err != nil {
			return err
		}
	}
	if _, err := t.exec(`UPDATE user_profiles SET first_name='ผู้ใช้ที่ลบแล้ว', last_name='', gender=NULL, date_of_birth=NULL, nationality=NULL,
		budget_level=NULL, travel_styles=NULL, favorite_destination=NULL, travel_with=NULL, avatar_key=NULL, bio=NULL WHERE user_id=?`, id); err != nil {
		return err
	}
	_, err := t.exec(`UPDATE users SET status='deleted', email=?, phone=NULL, handle=?, password_hash='' WHERE id=?`,
		fmt.Sprintf("deleted+%d@tep.invalid", id), fmt.Sprintf("deleted_%d", id), id)
	return err
}

func (t *sqlTx) AddRefresh(hash string, uid int64, expires, now time.Time) error {
	_, err := t.exec(`INSERT INTO user_sessions (user_id, refresh_token_hash, expires_at, created_at) VALUES (?,?,?,?)`, uid, hash, utc(expires), utc(now))
	return err
}

func (t *sqlTx) UseRefresh(hash string, now time.Time) (int64, RefreshState, error) {
	var id, uid int64
	var exp time.Time
	var revoked sql.NullTime
	err := t.row(`SELECT id, user_id, expires_at, revoked_at FROM user_sessions WHERE refresh_token_hash=? FOR UPDATE`, hash).Scan(&id, &uid, &exp, &revoked)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, RefreshUnknown, nil
	}
	if err != nil {
		return 0, RefreshUnknown, err
	}
	if revoked.Valid {
		return uid, RefreshReused, nil
	}
	if _, err := t.exec(`UPDATE user_sessions SET revoked_at=? WHERE id=?`, utc(now), id); err != nil {
		return 0, RefreshUnknown, err
	}
	if now.After(exp) {
		return uid, RefreshExpired, nil
	}
	return uid, RefreshOK, nil
}

func (t *sqlTx) RevokeRefresh(hash string, uid int64) error {
	_, err := t.exec(`DELETE FROM user_sessions WHERE refresh_token_hash=? AND user_id=?`, hash, uid)
	return err
}

func (t *sqlTx) RevokeAllRefresh(uid int64) error {
	_, err := t.exec(`DELETE FROM user_sessions WHERE user_id=?`, uid)
	return err
}

func (t *sqlTx) UpsertPushToken(uid int64, token, platform string, now time.Time) error {
	_, err := t.exec(`INSERT INTO device_push_tokens (user_id, token, platform, created_at) VALUES (?,?,?,?)
		ON DUPLICATE KEY UPDATE user_id=VALUES(user_id), platform=VALUES(platform)`, uid, token, platform, utc(now))
	return err
}

// ---- wallet ----

func (t *sqlTx) walletRow(uid int64, lock bool) (id int64, bal float64, err error) {
	for attempt := 0; attempt < 2; attempt++ {
		err = t.row(`SELECT id, balance FROM wallet_accounts WHERE user_id=?`+lockClause(lock), uid).Scan(&id, &bal)
		if !errors.Is(err, sql.ErrNoRows) {
			return
		}
		if _, e := t.exec(`INSERT IGNORE INTO wallet_accounts (user_id) VALUES (?)`, uid); e != nil {
			return 0, 0, e
		}
	}
	return
}

func (t *sqlTx) WalletBalance(uid int64, lock bool) (float64, error) {
	_, bal, err := t.walletRow(uid, lock)
	return bal, err
}

func (t *sqlTx) WalletAdjust(uid int64, delta float64, label string, now time.Time) (float64, error) {
	wid, bal, err := t.walletRow(uid, true)
	if err != nil {
		return 0, err
	}
	nb := round2(bal + delta)
	if nb < 0 {
		return bal, ErrInsufficient
	}
	dir, amt := "in", delta
	if delta < 0 {
		dir, amt = "out", -delta
	}
	if _, err := t.exec(`UPDATE wallet_accounts SET balance=? WHERE id=?`, nb, wid); err != nil {
		return 0, err
	}
	_, err = t.exec(`INSERT INTO wallet_transactions (wallet_id, direction, amount, label, created_at) VALUES (?,?,?,?,?)`,
		wid, dir, round2(math.Abs(amt)), label, utc(now))
	return nb, err
}

func (t *sqlTx) WalletTxs(uid int64) ([]WalletTx, error) {
	rows, err := t.query(`SELECT x.id, x.direction, x.label, x.amount, x.created_at FROM wallet_transactions x
		JOIN wallet_accounts a ON a.id = x.wallet_id WHERE a.user_id=? ORDER BY x.created_at DESC, x.id DESC`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []WalletTx{}
	for rows.Next() {
		var w WalletTx
		if err := rows.Scan(&w.ID, &w.Direction, &w.Label, &w.Amount, &w.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// ---- bookings & payments ----

const bookingSelect = `SELECT id, user_id, booking_number, item_type, title, subtitle, total, COALESCE(payment_method_label,''), booked_at, status,
  DATE_FORMAT(start_date,'%Y-%m-%d'), DATE_FORMAT(end_date,'%Y-%m-%d'), guests, subtotal, service_fee, cancel_reason, refund_amount FROM bookings`

func scanBooking(row interface{ Scan(...any) error }) (*bookingRec, error) {
	var b bookingRec
	var sub, start, end, reason sql.NullString
	var refund sql.NullFloat64
	if err := row.Scan(&b.ID, &b.UserID, &b.BookingNumber, &b.Type, &b.Title, &sub, &b.Amount, &b.Method, &b.Date, &b.Status,
		&start, &end, &b.Guests, &b.Subtotal, &b.ServiceFee, &reason, &refund); err != nil {
		return nil, notFound(err)
	}
	b.Subtitle, b.StartDate, b.EndDate, b.CancelReason = sub.String, start.String, end.String, reason.String
	if refund.Valid {
		v := refund.Float64
		b.RefundAmount = &v
	}
	return &b, nil
}

func (t *sqlTx) CreateBooking(b *bookingRec) error {
	res, err := t.exec(`INSERT INTO bookings (booking_number, user_id, status, item_type, title, subtitle, start_date, end_date, guests, subtotal, service_fee, total, booked_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`, "tmp-"+randID(""), b.UserID, b.Status, b.Type, b.Title, nullStr(b.Subtitle), nullStr(b.StartDate), nullStr(b.EndDate),
		b.Guests, b.Subtotal, b.ServiceFee, b.Amount, utc(b.Date))
	if err != nil {
		return err
	}
	b.ID, _ = res.LastInsertId()
	b.BookingNumber = bookingNumber(b.ID, b.Date)
	if _, err := t.exec(`UPDATE bookings SET booking_number=? WHERE id=?`, b.BookingNumber, b.ID); err != nil {
		return err
	}
	for _, l := range b.Lines {
		var lid any
		if l.ListingID != 0 {
			lid = l.ListingID
		}
		if _, err := t.exec(`INSERT INTO booking_line_items (booking_id, listing_id, description, quantity, unit_price) VALUES (?,?,?,?,?)`,
			b.ID, lid, l.Description, l.Quantity, l.UnitPrice); err != nil {
			return err
		}
	}
	return nil
}

func (t *sqlTx) GetBooking(id int64, lock bool) (*bookingRec, error) {
	return scanBooking(t.row(bookingSelect+` WHERE id=?`+lockClause(lock), id))
}

func (t *sqlTx) SaveBooking(b *bookingRec) error {
	var reason, refund any
	if b.CancelReason != "" {
		reason = b.CancelReason
	}
	if b.RefundAmount != nil {
		refund = *b.RefundAmount
	}
	res, err := t.exec(`UPDATE bookings SET status=?, payment_method_label=?, cancel_reason=?, refund_amount=?,
		cancelled_at = IF(?='cancelled', COALESCE(cancelled_at, UTC_TIMESTAMP()), NULL) WHERE id=?`,
		b.Status, nullStr(b.Method), reason, refund, b.Status, b.ID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return t.exists(`SELECT 1 FROM bookings WHERE id=?`, b.ID)
	}
	return nil
}

// exists returns ErrNotFound when the query yields no row.
func (t *sqlTx) exists(q string, args ...any) error {
	var one int
	return notFound(t.row(q, args...).Scan(&one))
}

func (t *sqlTx) ListBookings(uid int64, status string) ([]BookingSummary, error) {
	q, args := bookingSelect+` WHERE user_id=?`, []any{uid}
	if status != "" && status != "all" {
		q += ` AND status=?`
		args = append(args, status)
	}
	rows, err := t.query(q+` ORDER BY booked_at DESC, id DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []BookingSummary{}
	for rows.Next() {
		b, err := scanBooking(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b.BookingSummary)
	}
	return out, rows.Err()
}

const intentSelect = `SELECT id, user_id, COALESCE(booking_id,0), method, status, amount, COALESCE(promptpay_qr_url,''), COALESCE(client_secret,''), created_at FROM payment_intents`

func scanIntent(row interface{ Scan(...any) error }) (*intentRec, error) {
	var pi intentRec
	if err := row.Scan(&pi.ID, &pi.UserID, &pi.BookingID, &pi.Method, &pi.Status, &pi.Amount, &pi.PromptPayQrURL, &pi.ClientSecret, &pi.Created); err != nil {
		return nil, notFound(err)
	}
	return &pi, nil
}

func (t *sqlTx) CreateIntent(pi *intentRec) error {
	purpose := "booking"
	var bid any = pi.BookingID
	if pi.BookingID == 0 {
		purpose, bid = "wallet_topup", nil
	}
	_, err := t.exec(`INSERT INTO payment_intents (id, user_id, booking_id, purpose, method, status, amount, promptpay_qr_url, client_secret, created_at)
		VALUES (?,?,?,?,?,?,?,?,?,?)`, pi.ID, pi.UserID, bid, purpose, pi.Method, pi.Status, pi.Amount, nullStr(pi.PromptPayQrURL), nullStr(pi.ClientSecret), utc(pi.Created))
	if isDup(err) {
		return ErrDuplicate
	}
	return err
}

func (t *sqlTx) GetIntent(id string, lock bool) (*intentRec, error) {
	return scanIntent(t.row(intentSelect+` WHERE id=?`+lockClause(lock), id))
}

func (t *sqlTx) SaveIntent(pi *intentRec) error {
	_, err := t.exec(`UPDATE payment_intents SET status=? WHERE id=?`, pi.Status, pi.ID)
	return err
}

func (t *sqlTx) PendingIntents(uid int64) ([]*intentRec, error) {
	rows, err := t.query(intentSelect+` WHERE user_id=? AND status='pending' ORDER BY id`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*intentRec
	for rows.Next() {
		pi, err := scanIntent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, pi)
	}
	return out, rows.Err()
}

func (t *sqlTx) RecordPayment(bookingID int64, intentID, method string, amount float64, status string, now time.Time) error {
	_, err := t.exec(`INSERT INTO payments (booking_id, intent_id, method, status, amount, confirmed_at) VALUES (?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE intent_id=VALUES(intent_id), method=VALUES(method), status=VALUES(status), amount=VALUES(amount), confirmed_at=VALUES(confirmed_at)`,
		bookingID, intentID, method, status, amount, utc(now))
	return err
}

func (t *sqlTx) MarkPaymentRefunded(bookingID int64) error {
	_, err := t.exec(`UPDATE payments SET status='refunded' WHERE booking_id=?`, bookingID)
	return err
}

func (t *sqlTx) GetIdem(uid int64, endpoint, key string) (*idemRec, error) {
	var rec idemRec
	var body string
	err := t.row(`SELECT response_status, response_body FROM idempotency_keys WHERE user_id=? AND idem_key=? AND endpoint=?`, uid, key, endpoint).Scan(&rec.Status, &body)
	if err != nil {
		return nil, notFound(err)
	}
	rec.Body = []byte(body)
	return &rec, nil
}

func (t *sqlTx) PutIdem(uid int64, endpoint, key string, rec idemRec, now time.Time) error {
	_, err := t.exec(`INSERT INTO idempotency_keys (user_id, idem_key, endpoint, response_status, response_body, created_at) VALUES (?,?,?,?,?,?)`,
		uid, key, endpoint, rec.Status, string(rec.Body), utc(now))
	if isDup(err) {
		return ErrDuplicate
	}
	return err
}

func (t *sqlTx) AddWebhookEvent(provider, eventID string, now time.Time) error {
	_, err := t.exec(`INSERT INTO webhook_events (provider, event_id, received_at) VALUES (?,?,?)`, provider, eventID, utc(now))
	if isDup(err) {
		return ErrDuplicate
	}
	return err
}
