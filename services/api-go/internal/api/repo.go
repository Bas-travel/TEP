package api

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Storage layer.
//
// Business rules live once, in the handlers. A handler talks to storage only through Repo.Tx, which hands it a Tx
// (a small set of primitives). Two implementations exist and run the same test-suite:
//
//   - memRepo (repo_mem.go): in-memory, API_MODE=mock / no MYSQL_DSN. Everything in a Tx runs under one mutex.
//     It has no rollback, so handlers must validate before they mutate and return business errors before writing.
//   - sqlRepo (repo_sql.go): MySQL 8 with real transactions and row locks (SELECT … FOR UPDATE on wallet/booking/intent).
//
// Convention: a Tx function returns a non-nil error only when the whole unit of work must be undone (SQL rolls back).
// Expected business outcomes (wrong password, reused refresh token that must still revoke the family…) are reported
// through variables captured by the closure and the function returns nil, so their writes are committed.

var (
	ErrNotFound     = errors.New("not found")
	ErrDuplicate    = errors.New("duplicate")
	ErrInsufficient = errors.New("insufficient wallet balance")
	ErrForbidden    = errors.New("forbidden")
)

// Repo is the storage entry point.
type Repo interface {
	// Tx runs fn atomically. fn may be retried by the caller on ErrDuplicate (idempotency races).
	Tx(ctx context.Context, fn func(Tx) error) error
	Close() error
}

// RefreshState is the result of presenting a refresh token.
type RefreshState int

const (
	RefreshUnknown RefreshState = iota
	RefreshOK                   // active; it is now spent (rotated)
	RefreshReused               // already spent or revoked: caller must revoke the whole family
	RefreshExpired
)

type userRec struct {
	ID       int64
	Role     string
	Status   string // active | deleted
	PassHash string
	Bio      string
	Handle   string // public member id used by the community API (e.g. "m_nok", "m_12")
	Profile  UserProfile
}

type bookingLine struct {
	ListingID   int64 // 0 for a trip package
	Description string
	Quantity    int
	UnitPrice   float64
}

type bookingRec struct {
	BookingDetail
	UserID int64
	Lines  []bookingLine
}

type intentRec struct {
	PaymentIntent
	UserID    int64
	BookingID int64 // 0 for wallet top-up
	Created   time.Time
}

type idemRec struct {
	Status int
	Body   []byte
}

type reviewRec struct {
	ID        int64
	ListingID int64
	Rating    int
	Comment   string
	CreatedAt time.Time
}

type couponDef struct {
	Code        string
	Title       string
	Detail      string
	PercentOff  int // 0 => fixed amount (MaxDiscount)
	MinSpend    float64
	MaxDiscount float64
	ExpiresOn   string // YYYY-MM-DD
}

type jobRec struct {
	ItineraryJob
	Input any
}

// Tx is the set of storage primitives. All methods are scoped to the current transaction.
type Tx interface {
	// ---- users & sessions ----
	CreateUser(u *userRec, now time.Time) error // sets u.ID (and Profile.ID); ErrDuplicate on email/handle
	UserByID(id int64, lock bool) (*userRec, error)
	UserByEmail(email string) (*userRec, error)
	SaveUser(u *userRec) error
	DeleteUser(id int64, now time.Time) error // anonymise + status=deleted + revoke sessions + drop push tokens
	AddRefresh(hash string, uid int64, expires time.Time, now time.Time) error
	UseRefresh(hash string, now time.Time) (uid int64, st RefreshState, err error)
	RevokeRefresh(hash string, uid int64) error
	RevokeAllRefresh(uid int64) error
	UpsertPushToken(uid int64, token, platform string, now time.Time) error

	// ---- wallet ----
	WalletBalance(uid int64, lock bool) (float64, error)
	WalletAdjust(uid int64, delta float64, label string, now time.Time) (float64, error) // ErrInsufficient if it would go below 0
	WalletTxs(uid int64) ([]WalletTx, error)                                             // newest first

	// ---- bookings & payments ----
	CreateBooking(b *bookingRec) error // sets ID and BookingNumber
	GetBooking(id int64, lock bool) (*bookingRec, error)
	SaveBooking(b *bookingRec) error // status, method, cancel fields
	ListBookings(uid int64, status string) ([]BookingSummary, error)
	CreateIntent(pi *intentRec) error
	GetIntent(id string, lock bool) (*intentRec, error)
	SaveIntent(pi *intentRec) error
	PendingIntents(uid int64) ([]*intentRec, error)
	RecordPayment(bookingID int64, intentID, method string, amount float64, status string, now time.Time) error
	MarkPaymentRefunded(bookingID int64) error
	GetIdem(uid int64, endpoint, key string) (*idemRec, error)
	PutIdem(uid int64, endpoint, key string, rec idemRec, now time.Time) error // ErrDuplicate if the key already exists
	AddWebhookEvent(provider, eventID string, now time.Time) error             // ErrDuplicate => already processed

	// ---- engagement ----
	Settings(uid int64) (Settings, error)
	SaveSettings(uid int64, s Settings) error
	UpsertCoupon(c couponDef) error
	GrantCoupon(uid int64, code string) error
	UserCoupons(uid int64, today string) ([]UserCoupon, error)
	AddNotification(uid int64, n Notification) (Notification, error)
	Notifications(uid int64) ([]Notification, error) // newest first
	MarkNotificationRead(uid, id int64) error
	MarkAllNotificationsRead(uid int64) error
	AddTrip(uid int64, in SavedTripInput, now time.Time) (SavedTrip, error)
	Trips(uid int64) ([]SavedTrip, error)
	Trip(uid, id int64) (*SavedTrip, error)
	AddReview(uid, listingID int64, rating int, comment string, now time.Time) (reviewRec, error)
	Reviews(uid int64) ([]reviewRec, error)
	AddTicket(uid int64, subject, message string, now time.Time) (SupportTicket, error)
	Tickets(uid int64) ([]SupportTicket, error)
	AddJob(uid int64, j jobRec, now time.Time) error
	Job(uid int64, id string) (*ItineraryJob, error)

	// ---- community ----
	UpsertGroup(g GroupDetail) error
	Groups(uid int64) ([]Group, error)
	GroupByID(uid, id int64) (*GroupDetail, error)
	ToggleGroup(uid, id int64) (bool, error)
	Feed(viewer int64, mode, destination string) ([]Post, error)
	PostByID(viewer, id int64) (*Post, error)
	Comments(postID int64) ([]Comment, error) // oldest first
	AddPost(authorUID int64, p Post, imageKey string, now time.Time) (*Post, error)
	DeletePost(uid, id int64) error // ErrNotFound / ErrForbidden
	TogglePostLike(uid, id int64) (liked bool, count int, err error)
	TogglePostSave(uid, id int64) (bool, error)
	HidePost(uid, id int64) error
	ReportPost(uid, id int64, reason string, now time.Time) error
	AddComment(uid, postID int64, text string, now time.Time) (Comment, error)
	Member(viewer int64, handle string) (*MemberProfile, error)
	ToggleFollow(viewer int64, handle string) (bool, error)
}

func bookingNumber(id int64, t time.Time) string {
	return fmt.Sprintf("TEP-%s-%05d", t.Format("060102"), id)
}

// memberName is the display name used for community authors.
func memberName(p UserProfile) string {
	n := p.FirstName
	if r := []rune(p.LastName); len(r) > 0 {
		n += " " + string(r[0]) + "."
	}
	return n
}
