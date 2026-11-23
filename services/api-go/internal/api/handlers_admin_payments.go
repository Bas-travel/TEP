package api

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/Bas-travel/tep/services/api-go/internal/httpx"
)

// adminPaymentRoutes: payments console, reconciliation and admin refunds (admin-only, audited).
func (s *Server) adminPaymentRoutes(m *http.ServeMux) {
	p := s.portal(RoleAdmin)
	m.HandleFunc("GET /admin/v1/payments", p(s.adminListPayments))
	m.HandleFunc("GET /admin/v1/payments/summary", p(s.adminPaymentSummary))
	m.HandleFunc("GET /admin/v1/payments/reconciliation", p(s.adminReconciliation))
	m.HandleFunc("POST /admin/v1/payments/{bookingId}/refund", p(s.adminRefundPayment))
}

func (s *Server) adminListPayments(w http.ResponseWriter, r *http.Request, a actor) {
	q := r.URL.Query()
	off, lim, errs := pageParams(r)
	if v := q.Get("status"); v != "" && !oneOf(v, "pending", "succeeded", "failed", "refunded") {
		errs = append(errs, fe("status", "must be pending, succeeded, failed or refunded"))
	}
	if v := q.Get("method"); v != "" && !oneOf(v, "promptpay", "card", "wallet") {
		errs = append(errs, fe("method", "must be promptpay, card or wallet"))
	}
	if len(errs) > 0 {
		httpx.Invalid(w, r, errs...)
		return
	}
	var items []PaymentAdmin
	var total int
	err := s.repo.Tx(r.Context(), func(tx Tx) (e error) {
		items, total, e = tx.AdminPayments(PaymentFilter{Status: q.Get("status"), Method: q.Get("method"), Q: q.Get("q"), Offset: off, Limit: lim})
		return
	})
	if s.fail(w, r, err) {
		return
	}
	pageJSON(w, items, total, off, lim)
}

func (s *Server) adminPaymentSummary(w http.ResponseWriter, r *http.Request, a actor) {
	var items []PaymentTotal
	if s.fail(w, r, s.repo.Tx(r.Context(), func(tx Tx) (e error) { items, e = tx.PaymentTotals(); return })) {
		return
	}
	var net float64
	for _, t := range items {
		if t.Status == "succeeded" {
			net += t.Amount
		}
	}
	httpx.WriteJSON(w, 200, map[string]any{"items": items, "netCollected": net})
}

func (s *Server) adminReconciliation(w http.ResponseWriter, r *http.Request, a actor) {
	var items []ReconcileIssue
	if s.fail(w, r, s.repo.Tx(r.Context(), func(tx Tx) (e error) { items, e = tx.ReconcilePayments(); return })) {
		return
	}
	httpx.WriteJSON(w, 200, map[string]any{"items": items, "total": len(items), "balanced": len(items) == 0})
}

// adminRefundPayment cancels a paid (confirmed) booking and returns the money to the traveler's wallet.
func (s *Server) adminRefundPayment(w http.ResponseWriter, r *http.Request, a actor) {
	id, ok := pathID(r, "bookingId")
	if !ok {
		httpx.NotFound(w, r)
		return
	}
	var in struct {
		Reason string `json:"reason"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Invalid(w, r, fe("body", "invalid JSON"))
		return
	}
	in.Reason = strings.TrimSpace(in.Reason)
	if n := utf8.RuneCountInString(in.Reason); n < 3 || n > 255 {
		httpx.Invalid(w, r, fe("reason", "3-255 characters"))
		return
	}
	var out map[string]any
	err := s.repo.Tx(r.Context(), func(tx Tx) error {
		b, e := tx.GetBooking(id, true)
		if e != nil {
			return e
		}
		if b.Status != "confirmed" {
			return conflict("only paid, not yet completed bookings can be refunded; this one is " + b.Status)
		}
		if e := s.refundBooking(tx, b, "admin refund: "+in.Reason); e != nil {
			return e
		}
		out = map[string]any{"bookingId": b.ID, "bookingNumber": b.BookingNumber, "status": b.Status, "refundAmount": *b.RefundAmount}
		return s.audit(tx, a, "admin.payment.refund", "booking", b.ID, map[string]any{"amount": *b.RefundAmount, "reason": in.Reason})
	})
	if s.fail(w, r, err) {
		return
	}
	httpx.WriteJSON(w, 200, out)
}
