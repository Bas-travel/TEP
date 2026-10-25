package api

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Bas-travel/tep/services/api-go/internal/httpx"
)

var methodLabel = map[string]string{"promptpay": "PromptPay", "card": "บัตรเครดิต/เดบิต", "wallet": "Wallet"}

func (s *Server) listBookings(w http.ResponseWriter, r *http.Request, uid int64) {
	status := r.URL.Query().Get("status")
	switch status {
	case "", "all", "pending_payment", "confirmed", "completed", "cancelled":
	default:
		httpx.Invalid(w, r, httpx.FieldError{Field: "status", Message: "invalid status"})
		return
	}
	s.st.mu.Lock()
	items := []BookingSummary{}
	for _, b := range s.st.bookings {
		if b.UserID == uid && (status == "" || status == "all" || b.Status == status) {
			items = append(items, b.BookingSummary)
		}
	}
	s.st.mu.Unlock()
	sort.Slice(items, func(i, j int) bool { return items[i].Date.After(items[j].Date) })
	httpx.WriteJSON(w, 200, map[string]any{"items": items})
}

func (s *Server) getBooking(w http.ResponseWriter, r *http.Request, uid int64) {
	id, ok := pathID(r, "bookingId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	s.st.mu.Lock()
	b, found := s.st.bookings[id]
	var d BookingDetail
	if found && b.UserID == uid {
		d = b.BookingDetail
	} else {
		found = false
	}
	s.st.mu.Unlock()
	if !found {
		httpx.NotFound(w, r)
		return
	}
	httpx.WriteJSON(w, 200, d)
}

func (s *Server) createBooking(w http.ResponseWriter, r *http.Request, uid int64) {
	key := s.idemKey(r, uid, "bookings")
	if s.replay(w, key) {
		return
	}
	var req CreateBookingRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Invalid(w, r, httpx.FieldError{Field: "body", Message: err.Error()})
		return
	}
	var errs []httpx.FieldError
	var subtotal float64
	var title, subtitle, typ string
	for i, ln := range req.Lines {
		l, ok := findListing(ln.ListingID)
		if !ok {
			errs = append(errs, httpx.FieldError{Field: fmt.Sprintf("lines[%d].listingId", i), Message: "unknown listing"})
			continue
		}
		q := ln.Quantity
		if q == 0 {
			q = 1
		}
		if q < 1 || q > 20 {
			errs = append(errs, httpx.FieldError{Field: fmt.Sprintf("lines[%d].quantity", i), Message: "must be 1..20"})
			continue
		}
		subtotal += l.PriceFrom * float64(q)
		if title == "" {
			title, subtitle, typ = l.Name, l.Location, l.Type
		}
	}
	if len(req.Lines) == 0 && req.TripPackage == nil {
		errs = append(errs, httpx.FieldError{Field: "lines", Message: "at least one line or tripPackage required"})
	}
	if req.TripPackage != nil && len(req.Lines) == 0 {
		if req.TripPackage.Price <= 0 || strings.TrimSpace(req.TripPackage.Title) == "" {
			errs = append(errs, httpx.FieldError{Field: "tripPackage", Message: "title and positive price required"})
		} else {
			subtotal, title, subtitle, typ = req.TripPackage.Price, req.TripPackage.Title, req.TripPackage.Subtitle, "trip"
		}
	}
	start, end, guests := s.validateTrip(req, &errs)
	var discount float64
	if c := strings.ToUpper(strings.TrimSpace(req.CouponCode)); c != "" {
		if c != "WELCOME10" || subtotal < 1000 {
			errs = append(errs, httpx.FieldError{Field: "couponCode", Message: "coupon invalid or minimum spend not met"})
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
	s.st.mu.Lock()
	id := s.st.nextBooking
	s.st.nextBooking++
	b := &bookingRec{UserID: uid, BookingDetail: BookingDetail{
		BookingSummary: BookingSummary{ID: id, BookingNumber: s.st.bookingNumber(id, now), Type: typ, Title: title, Subtitle: subtitle,
			Amount: total, Date: now, Status: "pending_payment", StartDate: start, EndDate: end, Guests: guests},
		Subtotal: round2(subtotal), ServiceFee: fee}}
	s.st.bookings[id] = b
	d := b.BookingDetail
	s.st.mu.Unlock()
	s.respondIdem(w, key, 201, d)
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
		httpx.Invalid(w, r, httpx.FieldError{Field: "reason", Message: "required, max 500 chars"})
		return
	}
	s.st.mu.Lock()
	b, found := s.st.bookings[id]
	if !found || b.UserID != uid {
		s.st.mu.Unlock()
		httpx.NotFound(w, r)
		return
	}
	if b.Status != "pending_payment" && b.Status != "confirmed" {
		s.st.mu.Unlock()
		httpx.Invalid(w, r, httpx.FieldError{Field: "status", Message: "booking cannot be cancelled in state " + b.Status})
		return
	}
	refund := 0.0
	if b.Status == "confirmed" {
		refund = b.Amount
		u := s.st.users[uid]
		u.Wallet = round2(u.Wallet + refund)
		u.Txs = append(u.Txs, WalletTx{ID: s.st.nextTx, Direction: "in", Label: "คืนเงินการจอง " + b.BookingNumber, Amount: refund, CreatedAt: s.now()})
		s.st.nextTx++
	}
	b.Status, b.CancelReason, b.RefundAmount = "cancelled", req.Reason, &refund
	d := b.BookingDetail
	s.st.mu.Unlock()
	httpx.WriteJSON(w, 200, d)
}

func (s *Server) createIntent(w http.ResponseWriter, r *http.Request, uid int64) {
	key := s.idemKey(r, uid, "intents")
	if s.replay(w, key) {
		return
	}
	var req struct {
		BookingID int64  `json:"bookingId"`
		Method    string `json:"method"`
	}
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Invalid(w, r, httpx.FieldError{Field: "body", Message: err.Error()})
		return
	}
	if _, ok := methodLabel[req.Method]; !ok || req.BookingID <= 0 {
		httpx.Invalid(w, r, httpx.FieldError{Field: "method/bookingId", Message: "method must be promptpay|card|wallet and bookingId required"})
		return
	}
	s.st.mu.Lock()
	b, found := s.st.bookings[req.BookingID]
	if !found || b.UserID != uid {
		s.st.mu.Unlock()
		httpx.NotFound(w, r)
		return
	}
	if b.Status != "pending_payment" {
		s.st.mu.Unlock()
		httpx.Invalid(w, r, httpx.FieldError{Field: "bookingId", Message: "booking is not awaiting payment"})
		return
	}
	pi := &intentRec{UserID: uid, BookingID: b.ID, Created: s.now()}
	pi.ID = fmt.Sprintf("pi_%06d", s.st.nextIntent)
	s.st.nextIntent++
	pi.Method, pi.Amount = req.Method, b.Amount
	if req.Method == "wallet" {
		u := s.st.users[uid]
		if u.Wallet < b.Amount {
			s.st.mu.Unlock()
			httpx.WriteProblem(w, r, 402, httpx.CodeWalletShort, "Payment Required", "wallet balance is insufficient")
			return
		}
		u.Wallet = round2(u.Wallet - b.Amount)
		u.Txs = append(u.Txs, WalletTx{ID: s.st.nextTx, Direction: "out", Label: "ชำระการจอง " + b.BookingNumber, Amount: b.Amount, CreatedAt: s.now()})
		s.st.nextTx++
		pi.Status, pi.Settled = "succeeded", true
		b.Status, b.Method = "confirmed", methodLabel["wallet"]
	} else {
		pi.Status = "pending"
		if req.Method == "promptpay" {
			pi.PromptPayQrURL = "https://mock.tep.local/qr/" + pi.ID + ".png"
		} else {
			pi.ClientSecret = pi.ID + "_secret_mock"
		}
	}
	s.st.intents[pi.ID] = pi
	out := pi.PaymentIntent
	s.st.mu.Unlock()
	s.respondIdem(w, key, 201, out)
}

// settleLocked applies the mock PSP outcome once the settle delay has elapsed. Caller holds st.mu.
func (s *Server) settleLocked(pi *intentRec) {
	if pi.Settled || s.now().Sub(pi.Created) < s.settleAfter {
		return
	}
	pi.Settled, pi.Status = true, "succeeded"
	if pi.BookingID != 0 {
		if b := s.st.bookings[pi.BookingID]; b != nil && b.Status == "pending_payment" {
			b.Status, b.Method = "confirmed", methodLabel[pi.Method]
		}
		return
	}
	u := s.st.users[pi.UserID]
	u.Wallet = round2(u.Wallet + pi.Amount)
	u.Txs = append(u.Txs, WalletTx{ID: s.st.nextTx, Direction: "in", Label: "เติมเงิน PromptPay", Amount: pi.Amount, CreatedAt: s.now()})
	s.st.nextTx++
}

func (s *Server) getIntent(w http.ResponseWriter, r *http.Request, uid int64) {
	s.st.mu.Lock()
	pi, ok := s.st.intents[r.PathValue("intentId")]
	if !ok || pi.UserID != uid {
		s.st.mu.Unlock()
		httpx.NotFound(w, r)
		return
	}
	s.settleLocked(pi)
	out := pi.PaymentIntent
	s.st.mu.Unlock()
	httpx.WriteJSON(w, 200, out)
}

func (s *Server) getWallet(w http.ResponseWriter, r *http.Request, uid int64) {
	s.st.mu.Lock()
	for _, pi := range s.st.intents { // settle due top-ups so balance is fresh
		if pi.UserID == uid {
			s.settleLocked(pi)
		}
	}
	u := s.st.users[uid]
	txs := append([]WalletTx{}, u.Txs...)
	out := Wallet{Balance: u.Wallet, Currency: "THB", Transactions: txs}
	s.st.mu.Unlock()
	sort.Slice(out.Transactions, func(i, j int) bool { return out.Transactions[i].CreatedAt.After(out.Transactions[j].CreatedAt) })
	httpx.WriteJSON(w, 200, out)
}

func (s *Server) topUp(w http.ResponseWriter, r *http.Request, uid int64) {
	var req struct {
		Amount float64 `json:"amount"`
		Method string  `json:"method"`
	}
	if err := httpx.Decode(r, &req); err != nil || req.Amount < 1 || req.Amount > 100000 || (req.Method != "" && req.Method != "promptpay") {
		httpx.Invalid(w, r, httpx.FieldError{Field: "amount", Message: "must be between 1 and 100000 (method promptpay)"})
		return
	}
	s.st.mu.Lock()
	pi := &intentRec{UserID: uid, Created: time.Now()}
	pi.Created = s.now()
	pi.ID = fmt.Sprintf("pi_%06d", s.st.nextIntent)
	s.st.nextIntent++
	pi.Method, pi.Amount, pi.Status = "promptpay", round2(req.Amount), "pending"
	pi.PromptPayQrURL = "https://mock.tep.local/qr/" + pi.ID + ".png"
	s.st.intents[pi.ID] = pi
	out := pi.PaymentIntent
	s.st.mu.Unlock()
	httpx.WriteJSON(w, 202, out)
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
