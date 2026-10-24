package api

import (
	"fmt"
	"sync"
	"time"
)

type userRec struct {
	ID       int64
	Role     string
	PassHash string
	Profile  UserProfile
	Wallet   float64
	Txs      []WalletTx
}

type bookingRec struct {
	BookingDetail
	UserID int64
}

type intentRec struct {
	PaymentIntent
	UserID    int64
	BookingID int64 // 0 for wallet top-up
	Created   time.Time
	Settled   bool
}

type idemRec struct {
	Status int
	Body   []byte
}

// store is the in-memory DEV/mock backing store (API_MODE=mock). Not for production.
type store struct {
	mu                                        sync.Mutex
	users                                     map[int64]*userRec
	byEmail                                   map[string]int64
	refresh                                   map[string]int64 // active refresh token -> user id
	used                                      map[string]int64 // rotated (spent) refresh token -> user id, for reuse detection
	bookings                                  map[int64]*bookingRec
	intents                                   map[string]*intentRec
	idem                                      map[string]idemRec
	nextUser, nextBooking, nextTx, nextIntent int64
}

func newStore() *store {
	return &store{
		users: map[int64]*userRec{}, byEmail: map[string]int64{},
		refresh: map[string]int64{}, used: map[string]int64{},
		bookings: map[int64]*bookingRec{}, intents: map[string]*intentRec{}, idem: map[string]idemRec{},
		nextUser: 1, nextBooking: 1, nextTx: 1, nextIntent: 1,
	}
}

func (s *store) bookingNumber(id int64, t time.Time) string {
	return fmt.Sprintf("TEP-%s-%05d", t.Format("060102"), id)
}
