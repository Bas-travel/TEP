package api

import (
	"database/sql"
	"strings"
)

func (t *sqlTx) AdminPayments(f PaymentFilter) ([]PaymentAdmin, int, error) {
	var w whereBuilder
	if f.Status != "" {
		w.add(`p.status=?`, f.Status)
	}
	if f.Method != "" {
		w.add(`p.method=?`, f.Method)
	}
	if q := strings.TrimSpace(f.Q); q != "" {
		w.add(`(b.booking_number LIKE ? OR u.email LIKE ?)`, likeArg(q), likeArg(q))
	}
	from := ` FROM payments p JOIN bookings b ON b.id=p.booking_id JOIN users u ON u.id=b.user_id`
	var total int
	if err := t.row(`SELECT COUNT(*)`+from+w.sql(), w.args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args := append(append([]any{}, w.args...), f.Limit, f.Offset)
	rows, err := t.query(`SELECT b.id, b.booking_number, u.email, p.method, p.status, p.amount, COALESCE(p.intent_id,''), p.confirmed_at, b.status, b.refund_amount`+
		from+w.sql()+` ORDER BY p.confirmed_at DESC, p.id DESC LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []PaymentAdmin{}
	for rows.Next() {
		var r PaymentAdmin
		var at sql.NullTime
		var refund sql.NullFloat64
		if err := rows.Scan(&r.BookingID, &r.BookingNumber, &r.Traveler, &r.Method, &r.Status, &r.Amount, &r.IntentID, &at, &r.BookingStatus, &refund); err != nil {
			return nil, 0, err
		}
		if at.Valid {
			v := at.Time
			r.ConfirmedAt = &v
		}
		if refund.Valid {
			v := refund.Float64
			r.RefundAmount = &v
		}
		out = append(out, r)
	}
	return out, total, rows.Err()
}

func (t *sqlTx) PaymentTotals() ([]PaymentTotal, error) {
	rows, err := t.query(`SELECT status, COUNT(*), COALESCE(SUM(amount),0) FROM payments GROUP BY status ORDER BY status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []PaymentTotal{}
	for rows.Next() {
		var r PaymentTotal
		if err := rows.Scan(&r.Status, &r.Count, &r.Amount); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (t *sqlTx) ReconcilePayments() ([]ReconcileIssue, error) {
	rows, err := t.query(`SELECT b.id, b.booking_number, b.status, b.total, b.refund_amount, p.id IS NOT NULL, COALESCE(p.status,''), COALESCE(p.amount,0)
		FROM bookings b LEFT JOIN payments p ON p.booking_id=b.id ORDER BY b.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ReconcileIssue{}
	for rows.Next() {
		var in reconcileInput
		var refund sql.NullFloat64
		if err := rows.Scan(&in.BookingID, &in.BookingNumber, &in.Status, &in.Total, &refund, &in.HasPayment, &in.PayStatus, &in.PayAmount); err != nil {
			return nil, err
		}
		if refund.Valid {
			v := refund.Float64
			in.Refund = &v
		}
		out = append(out, classifyReconcile(in)...)
	}
	return out, rows.Err()
}
