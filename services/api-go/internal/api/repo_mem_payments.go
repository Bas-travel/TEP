package api

import (
	"sort"
	"strings"
)

func (m *memRepo) AdminPayments(f PaymentFilter) ([]PaymentAdmin, int, error) {
	q := strings.ToLower(strings.TrimSpace(f.Q))
	var all []PaymentAdmin
	for id, p := range m.pays {
		b := m.bookings[id]
		if b == nil || (f.Status != "" && p.status != f.Status) || (f.Method != "" && p.method != f.Method) {
			continue
		}
		email := ""
		if u := m.users[b.UserID]; u != nil {
			email = u.Profile.Email
		}
		if q != "" && !strings.Contains(strings.ToLower(b.BookingNumber), q) && !strings.Contains(strings.ToLower(email), q) {
			continue
		}
		row := PaymentAdmin{BookingID: id, BookingNumber: b.BookingNumber, Traveler: email, Method: p.method, Status: p.status,
			Amount: p.amount, IntentID: p.intent, BookingStatus: b.Status, RefundAmount: b.RefundAmount}
		if !p.at.IsZero() {
			at := p.at
			row.ConfirmedAt = &at
		}
		all = append(all, row)
	}
	sort.Slice(all, func(i, j int) bool {
		ai, aj := all[i].ConfirmedAt, all[j].ConfirmedAt
		if ai != nil && aj != nil && !ai.Equal(*aj) {
			return ai.After(*aj)
		}
		return all[i].BookingID > all[j].BookingID
	})
	total := len(all)
	if f.Offset >= total {
		return []PaymentAdmin{}, total, nil
	}
	end := min(total, f.Offset+f.Limit)
	return all[f.Offset:end], total, nil
}

func (m *memRepo) PaymentTotals() ([]PaymentTotal, error) {
	by := map[string]*PaymentTotal{}
	for _, p := range m.pays {
		t := by[p.status]
		if t == nil {
			t = &PaymentTotal{Status: p.status}
			by[p.status] = t
		}
		t.Count++
		t.Amount += p.amount
	}
	out := []PaymentTotal{}
	for _, t := range by {
		out = append(out, *t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Status < out[j].Status })
	return out, nil
}

func (m *memRepo) ReconcilePayments() ([]ReconcileIssue, error) {
	ids := make([]int64, 0, len(m.bookings))
	for id := range m.bookings {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	out := []ReconcileIssue{}
	for _, id := range ids {
		b := m.bookings[id]
		in := reconcileInput{BookingID: id, BookingNumber: b.BookingNumber, Status: b.Status, Total: b.Amount, Refund: b.RefundAmount}
		if p := m.pays[id]; p != nil {
			in.HasPayment, in.PayStatus, in.PayAmount = true, p.status, p.amount
		}
		out = append(out, classifyReconcile(in)...)
	}
	return out, nil
}
