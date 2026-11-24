package api

import (
	"math"
	"time"
)

// PaymentAdmin is one row of the admin payments console (a payment joined with its booking and traveler).
type PaymentAdmin struct {
	BookingID     int64      `json:"bookingId"`
	BookingNumber string     `json:"bookingNumber"`
	Traveler      string     `json:"traveler"` // traveler e-mail
	Method        string     `json:"method"`
	Status        string     `json:"status"` // pending | succeeded | failed | refunded
	Amount        float64    `json:"amount"`
	IntentID      string     `json:"intentId,omitempty"`
	ConfirmedAt   *time.Time `json:"confirmedAt,omitempty"`
	BookingStatus string     `json:"bookingStatus"`
	RefundAmount  *float64   `json:"refundAmount,omitempty"`
}

type PaymentFilter struct {
	Status, Method, Q string
	Offset, Limit     int
}

// PaymentTotal aggregates payments by status.
type PaymentTotal struct {
	Status string  `json:"status"`
	Count  int     `json:"count"`
	Amount float64 `json:"amount"`
}

// ReconcileIssue is a disagreement between a booking and its payment record.
type ReconcileIssue struct {
	BookingID     int64   `json:"bookingId"`
	BookingNumber string  `json:"bookingNumber"`
	Kind          string  `json:"kind"`
	Detail        string  `json:"detail"`
	Expected      float64 `json:"expected"`
	Actual        float64 `json:"actual"`
}

// reconcileInput is the storage-neutral view both repositories feed into classifyReconcile.
type reconcileInput struct {
	BookingID     int64
	BookingNumber string
	Status        string
	Total         float64
	Refund        *float64
	HasPayment    bool
	PayStatus     string
	PayAmount     float64
}

func classifyReconcile(in reconcileInput) []ReconcileIssue {
	issue := func(kind, detail string, exp, act float64) ReconcileIssue {
		return ReconcileIssue{BookingID: in.BookingID, BookingNumber: in.BookingNumber, Kind: kind, Detail: detail, Expected: exp, Actual: act}
	}
	diff := func(a, b float64) bool { return math.Abs(a-b) > 0.005 }
	switch in.Status {
	case "confirmed", "completed":
		switch {
		case !in.HasPayment:
			return []ReconcileIssue{issue("missing_payment", "paid booking has no payment record", in.Total, 0)}
		case in.PayStatus != "succeeded":
			return []ReconcileIssue{issue("payment_not_succeeded", "paid booking, payment is "+in.PayStatus, in.Total, 0)}
		case diff(in.PayAmount, in.Total):
			return []ReconcileIssue{issue("amount_mismatch", "payment amount differs from the booking total", in.Total, in.PayAmount)}
		}
	case "cancelled":
		if !in.HasPayment {
			return nil
		}
		refund := 0.0
		if in.Refund != nil {
			refund = *in.Refund
		}
		if in.PayStatus == "succeeded" {
			return []ReconcileIssue{issue("refund_mismatch", "cancelled booking still holds a succeeded payment (money not returned)", in.PayAmount, refund)}
		}
		if (in.PayStatus == "refunded") != (refund > 0) {
			return []ReconcileIssue{issue("refund_mismatch", "cancelled booking: refund and payment status disagree ("+in.PayStatus+")", in.PayAmount, refund)}
		}
		if in.PayStatus == "refunded" && diff(refund, in.PayAmount) {
			return []ReconcileIssue{issue("refund_mismatch", "refunded amount differs from the payment amount", in.PayAmount, refund)}
		}
	case "pending_payment":
		if in.HasPayment {
			return []ReconcileIssue{issue("unexpected_payment", "booking awaits payment but a payment record exists", 0, in.PayAmount)}
		}
	}
	return nil
}
