package api

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Bas-travel/tep/services/api-go/internal/httpx"
)

var methodLabel = map[string]string{"promptpay": "PromptPay", "card": "บัตรเครดิต/เดบิต", "wallet": "Wallet"}

func newIntentID() string {
	b := make([]byte, 9)
	_, _ = rand.Read(b)
	return "pi_" + hex.EncodeToString(b)
}

func (s *Server) listBookings(w http.ResponseWriter, r *http.Request, uid int64) {
	status := r.URL.Query().Get("status")
	switch status {
	case "", "all", "pending_payment", "confirmed", "completed", "cancelled":
	default:
		httpx.Invalid(w, r, fe("status", "invalid status"))
		return
	}
	var items []BookingSummary
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) { items, e = tx.ListBookings(uid, status); return })
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, map[string]any{"items": items})
}

// ownBooking loads a booking and hides other users' bookings behind 404.
func ownBooking(tx Tx, uid, id int64, lock bool) (*bookingRec, error) {
	b, err := tx.GetBooking(id, lock)
	if err != nil {
		return nil, err
	}
	if b.UserID != uid {
		return nil, ErrNotFound
	}
	return b, nil
}

func (s *Server) getBooking(w http.ResponseWriter, r *http.Request, uid int64) {
	id, ok := pathID(r, "bookingId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var d BookingDetail
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		b, e := ownBooking(tx, uid, id, false)
		if e == nil {
			d = b.BookingDetail
		}
		return e
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, d)
}

func (s *Server) createBooking(w http.ResponseWriter, r *http.Request, uid int64) {
	var req CreateBookingRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	var errs []httpx.FieldError
	var subtotal float64
	var title, subtitle, typ string
	var lines []bookingLine
	for i, ln := range req.Lines {
		l, ok := findListing(ln.ListingID)
		if !ok {
			errs = append(errs, fe(fmt.Sprintf("lines[%d].listingId", i), "unknown listing"))
			continue
		}
		q := ln.Quantity
		if q == 0 {
			q = 1
		}
		if q < 1 || q > 20 {
			errs = append(errs, fe(fmt.Sprintf("lines[%d].quantity", i), "must be 1..20"))
			continue
		}
		subtotal += l.PriceFrom * float64(q)
		lines = append(lines, bookingLine{ListingID: l.ID, Description: l.Name, Quantity: q, UnitPrice: l.PriceFrom})
		if title == "" {
			title, subtitle, typ = l.Name, l.Location, l.Type
		}
	}
	if len(req.Lines) == 0 && req.TripPackage == nil {
		errs = append(errs, fe("lines", "at least one line or tripPackage required"))
	}
	if req.TripPackage != nil && len(req.Lines) == 0 {
		if req.TripPackage.Price <= 0 || strings.TrimSpace(req.TripPackage.Title) == "" || len([]rune(req.TripPackage.Title)) > 190 {
			errs = append(errs, fe("tripPackage", "title (max 190) and positive price required"))
		} else {
			subtotal, title, subtitle, typ = req.TripPackage.Price, req.TripPackage.Title, req.TripPackage.Subtitle, "trip"
			lines = append(lines, bookingLine{Description: title, Quantity: 1, UnitPrice: subtotal})
		}
	}
	start, end, guests := s.validateTrip(req, &errs)
	var discount float64
	if c := strings.ToUpper(strings.TrimSpace(req.CouponCode)); c != "" {
		if c != "WELCOME10" || subtotal < 1000 {
			errs = append(errs, fe("couponCode", "coupon invalid or minimum spend not met"))
		} else {
			discount = min(round2(subtotal*0.10), 500)
		}
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	fee := round2(subtotal * 0.03)
	total := round2(subtotal + fee - discount)
	now := s.now()
	s.idempotent(w, r, uid, "bookings", func(tx Tx) (int, any, error) {
		b := &bookingRec{UserID: uid, Lines: lines, BookingDetail: BookingDetail{
			BookingSummary: BookingSummary{Type: typ, Title: title, Subtitle: subtitle, Amount: total, Date: now, Status: "pending_payment",
				StartDate: start, EndDate: end, Guests: guests},
			Subtotal: round2(subtotal), ServiceFee: fee}}
		if err := tx.CreateBooking(b); err != nil {
			return 0, nil, err
		}
		return 201, b.BookingDetail, nil
	})
}

func (s *Server) cancelBooking(w http.ResponseWriter, r *http.Request, uid int64) {
	id, ok := pathID(r, "bookingId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if err := httpx.Decode(r, &req); err != nil || strings.TrimSpace(req.Reason) == "" || len(req.Reason) > 500 {
		httpx.Invalid(w, r, fe("reason", "required, max 500 chars"))
		return
	}
	var d BookingDetail
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		b, e := ownBooking(tx, uid, id, true)
		if e != nil {
			return e
		}
		if b.Status != "pending_payment" && b.Status != "confirmed" {
			return invalid(fe("status", "booking cannot be cancelled in state "+b.Status))
		}
		refund := 0.0
		if b.Status == "confirmed" {
			refund = b.Amount
			if _, e := tx.WalletAdjust(uid, refund, "คืนเงินการจอง "+b.BookingNumber, s.now()); e != nil {
				return e
			}
			if e := tx.MarkPaymentRefunded(b.ID); e != nil {
				return e
			}
		}
		b.Status, b.CancelReason, b.RefundAmount = "cancelled", req.Reason, &refund
		if e := tx.SaveBooking(b); e != nil {
			return e
		}
		d = b.BookingDetail
		return nil
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, d)
}

// confirmBooking marks a pending booking as paid. The caller holds the booking row lock.
func (s *Server) confirmBooking(tx Tx, b *bookingRec, pi *intentRec) error {
	b.Status, b.Method = "confirmed", methodLabel[pi.Method]
	if err := tx.SaveBooking(b); err != nil {
		return err
	}
	return tx.RecordPayment(b.ID, pi.ID, pi.Method, pi.Amount, "succeeded", s.now())
}

func (s *Server) createIntent(w http.ResponseWriter, r *http.Request, uid int64) {
	var req struct {
		BookingID int64  `json:"bookingId"`
		Method    string `json:"method"`
	}
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Invalid(w, r, fe("body", err.Error()))
		return
	}
	if _, ok := methodLabel[req.Method]; !ok || req.BookingID <= 0 {
		httpx.Invalid(w, r, fe("method/bookingId", "method must be promptpay|card|wallet and bookingId required"))
		return
	}
	s.idempotent(w, r, uid, "intents", func(tx Tx) (int, any, error) {
		b, err := ownBooking(tx, uid, req.BookingID, true)
		if err != nil {
			return 0, nil, err
		}
		if b.Status != "pending_payment" {
			return 0, nil, invalid(fe("bookingId", "booking is not awaiting payment"))
		}
		pi := &intentRec{UserID: uid, BookingID: b.ID, Created: s.now()}
		pi.ID, pi.Method, pi.Amount = newIntentID(), req.Method, b.Amount
		if req.Method == "wallet" {
			if _, err := tx.WalletAdjust(uid, -b.Amount, "ชำระการจอง "+b.BookingNumber, s.now()); err != nil {
				return 0, nil, err // ErrInsufficient -> 402
			}
			pi.Status = "succeeded"
			if err := s.confirmBooking(tx, b, pi); err != nil {
				return 0, nil, err
			}
		} else {
			pi.Status = "pending"
			if req.Method == "promptpay" {
				pi.PromptPayQrURL = "https://mock.tep.local/qr/" + pi.ID + ".png"
			} else {
				pi.ClientSecret = pi.ID + "_secret_mock"
			}
		}
		if err := tx.CreateIntent(pi); err != nil {
			return 0, nil, err
		}
		return 201, pi.PaymentIntent, nil
	})
}

// finishIntent applies the PSP outcome ("succeeded" | "failed") to a pending intent. Caller holds the intent lock.
func (s *Server) finishIntent(tx Tx, pi *intentRec, status string) error {
	if pi.Status != "pending" {
		return nil
	}
	pi.Status = status
	if err := tx.SaveIntent(pi); err != nil {
		return err
	}
	if status != "succeeded" {
		return nil
	}
	if pi.BookingID != 0 {
		b, err := tx.GetBooking(pi.BookingID, true)
		if err != nil {
			return err
		}
		if b.Status == "pending_payment" {
			return s.confirmBooking(tx, b, pi)
		}
		return nil
	}
	_, err := tx.WalletAdjust(pi.UserID, pi.Amount, "เติมเงิน PromptPay", s.now())
	return err
}

// settleIfDue is the mock PSP (MOCK_PAYMENT=true): a pending intent succeeds after settleAfter.
func (s *Server) settleIfDue(tx Tx, pi *intentRec) error {
	if pi.Status == "pending" && s.cfg.MockPayment && s.now().Sub(pi.Created) >= s.settleAfter {
		return s.finishIntent(tx, pi, "succeeded")
	}
	return nil
}

func (s *Server) getIntent(w http.ResponseWriter, r *http.Request, uid int64) {
	var out PaymentIntent
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		pi, e := tx.GetIntent(r.PathValue("intentId"), true)
		if e != nil {
			return e
		}
		if pi.UserID != uid {
			return ErrNotFound
		}
		if e := s.settleIfDue(tx, pi); e != nil {
			return e
		}
		out = pi.PaymentIntent
		return nil
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}

func (s *Server) getWallet(w http.ResponseWriter, r *http.Request, uid int64) {
	var out Wallet
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		pend, e := tx.PendingIntents(uid) // settle due top-ups so the balance is fresh
		if e != nil {
			return e
		}
		for _, p := range pend {
			pi, e := tx.GetIntent(p.ID, true)
			if e != nil {
				return e
			}
			if e := s.settleIfDue(tx, pi); e != nil {
				return e
			}
		}
		bal, e := tx.WalletBalance(uid, false)
		if e != nil {
			return e
		}
		txs, e := tx.WalletTxs(uid)
		if e != nil {
			return e
		}
		out = Wallet{Balance: bal, Currency: "THB", Transactions: txs}
		return nil
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}

func (s *Server) topUp(w http.ResponseWriter, r *http.Request, uid int64) {
	var req struct {
		Amount float64 `json:"amount"`
		Method string  `json:"method"`
	}
	if err := httpx.Decode(r, &req); err != nil || req.Amount < 1 || req.Amount > 100000 || (req.Method != "" && req.Method != "promptpay") {
		httpx.Invalid(w, r, fe("amount", "must be between 1 and 100000 (method promptpay)"))
		return
	}
	var out PaymentIntent
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		pi := &intentRec{UserID: uid, Created: s.now()}
		pi.ID, pi.Method, pi.Amount, pi.Status = newIntentID(), "promptpay", round2(req.Amount), "pending"
		pi.PromptPayQrURL = "https://mock.tep.local/qr/" + pi.ID + ".png"
		out = pi.PaymentIntent
		return tx.CreateIntent(pi)
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 202, out)
}

// paymentWebhook receives PSP callbacks. The raw body is authenticated with HMAC-SHA256 (hex, header X-Signature) using
// WEBHOOK_SECRET; each (provider, eventId) is processed once.
func (s *Server) paymentWebhook(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	if s.cfg.WebhookSecret == "" || (provider != "promptpay" && provider != "card" && provider != "mock") {
		httpx.NotFound(w, r)
		return
	}
	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		httpx.Invalid(w, r, fe("body", "unreadable"))
		return
	}
	mac := hmac.New(sha256.New, []byte(s.cfg.WebhookSecret))
	mac.Write(raw)
	want := mac.Sum(nil)
	got, derr := hex.DecodeString(strings.TrimSpace(r.Header.Get("X-Signature")))
	if derr != nil || !hmac.Equal(got, want) {
		httpx.Unauthorized(w, r, "invalid signature")
		return
	}
	var ev struct {
		EventID  string `json:"eventId"`
		IntentID string `json:"intentId"`
		Status   string `json:"status"`
	}
	if err := json.Unmarshal(raw, &ev); err != nil || ev.EventID == "" || len(ev.EventID) > 120 || ev.IntentID == "" || (ev.Status != "succeeded" && ev.Status != "failed") {
		httpx.Invalid(w, r, fe("body", "eventId, intentId and status (succeeded|failed) required"))
		return
	}
	duplicate := false
	err = s.repo.Tx(r.Context(), func(tx Tx) error {
		if e := tx.AddWebhookEvent(provider, ev.EventID, s.now()); e != nil {
			if e == ErrDuplicate {
				duplicate = true
				return nil
			}
			return e
		}
		pi, e := tx.GetIntent(ev.IntentID, true)
		if e != nil {
			return e
		}
		return s.finishIntent(tx, pi, ev.Status)
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, map[string]any{"received": true, "duplicate": duplicate})
}

// validateTrip checks the optional stay/visit dates and guest count of a booking request (finding F-1).
// Dates are calendar dates (YYYY-MM-DD) interpreted in the server's local zone; the start may not be in the past.
func (s *Server) validateTrip(req CreateBookingRequest, errs *[]httpx.FieldError) (start, end string, guests int) {
	const layout = "2006-01-02"
	guests = req.Guests
	if guests == 0 {
		guests = 1
	}
	if guests < 1 || guests > 20 {
		*errs = append(*errs, httpx.FieldError{Field: "guests", Message: "must be 1..20"})
	}
	today := s.now().Format(layout)
	var sd, ed time.Time
	var err error
	if req.StartDate != "" {
		if sd, err = time.Parse(layout, req.StartDate); err != nil {
			*errs = append(*errs, httpx.FieldError{Field: "startDate", Message: "must be a date YYYY-MM-DD"})
		} else if req.StartDate < today {
			*errs = append(*errs, httpx.FieldError{Field: "startDate", Message: "must not be in the past"})
		} else {
			start = req.StartDate
		}
	}
	if req.EndDate != "" {
		switch {
		case req.StartDate == "":
			*errs = append(*errs, httpx.FieldError{Field: "endDate", Message: "startDate is required when endDate is given"})
		default:
			if ed, err = time.Parse(layout, req.EndDate); err != nil {
				*errs = append(*errs, httpx.FieldError{Field: "endDate", Message: "must be a date YYYY-MM-DD"})
			} else if start != "" && ed.Before(sd) {
				*errs = append(*errs, httpx.FieldError{Field: "endDate", Message: "must not be before startDate"})
			} else {
				end = req.EndDate
			}
		}
	}
	return start, end, guests
}
