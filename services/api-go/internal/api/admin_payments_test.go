package api

import (
	"strings"
	"testing"
	"time"
)

func TestAdminPaymentsConsole(t *testing.T) {
	e := newEnv(t, nil)
	ad, tr := e.loginAs("admin@tep.dev"), e.login()
	for _, p := range []string{"/admin/v1/payments", "/admin/v1/payments/summary", "/admin/v1/payments/reconciliation"} {
		if code, _, _ := e.do("GET", p, tr, nil); code != 403 {
			t.Errorf("traveler GET %s = %d, want 403", p, code)
		}
	}
	for _, q := range []string{"?status=bogus", "?method=bitcoin", "?limit=0"} {
		if code, _, _ := e.do("GET", "/admin/v1/payments"+q, ad, nil); code != 422 {
			t.Errorf("GET payments%s = %d, want 422", q, code)
		}
	}
	base := e.itemsTotal(ad, "/admin/v1/payments")

	// a paid (PromptPay) booking
	_, b, _ := e.do("POST", "/v1/bookings", tr, map[string]any{"lines": []any{map[string]any{"listingId": 1, "quantity": 1}}})
	bid := itoa(b["id"].(float64))
	_, pi, _ := e.do("POST", "/v1/payments/intents", tr, map[string]any{"bookingId": b["id"], "method": "promptpay"})
	e.now = e.now.Add(4 * time.Second)
	e.do("GET", "/v1/payments/intents/"+pi["id"].(string), tr, nil)

	if got := e.itemsTotal(ad, "/admin/v1/payments"); got != base+1 {
		t.Fatalf("payments total %d, want %d", got, base+1)
	}
	code, list, _ := e.do("GET", "/admin/v1/payments?status=succeeded&method=promptpay&q="+b["bookingNumber"].(string), ad, nil)
	rows := pitems(list)
	if code != 200 || len(rows) != 1 {
		t.Fatalf("filtered list = %d %v", code, list)
	}
	row := rows[0].(map[string]any)
	if row["amount"].(float64) != 3296 || row["status"] != "succeeded" || row["bookingStatus"] != "confirmed" || !strings.Contains(row["traveler"].(string), "@") {
		t.Errorf("row = %v", row)
	}
	_, sum, _ := e.do("GET", "/admin/v1/payments/summary", ad, nil)
	if sum["netCollected"].(float64) < 3296 {
		t.Errorf("summary = %v", sum)
	}
	if _, rec, _ := e.do("GET", "/admin/v1/payments/reconciliation", ad, nil); rec["balanced"] != true {
		t.Fatalf("books should balance: %v", rec)
	}

	// refund guards
	if code, _, _ := e.do("POST", "/admin/v1/payments/"+bid+"/refund", tr, map[string]any{"reason": "ลูกค้าขอ"}); code != 403 {
		t.Errorf("traveler refund = %d, want 403", code)
	}
	if code, _, _ := e.do("POST", "/admin/v1/payments/"+bid+"/refund", ad, map[string]any{"reason": "x"}); code != 422 {
		t.Errorf("short reason = %d, want 422", code)
	}
	if code, _, _ := e.do("POST", "/admin/v1/payments/999999/refund", ad, map[string]any{"reason": "ไม่มีรายการ"}); code != 404 {
		t.Errorf("missing booking = %d, want 404", code)
	}
	_, w0, _ := e.do("GET", "/v1/wallet", tr, nil)
	code, out, _ := e.do("POST", "/admin/v1/payments/"+bid+"/refund", ad, map[string]any{"reason": "บริการไม่ตรงตามที่ระบุ"})
	if code != 200 || out["status"] != "cancelled" || out["refundAmount"].(float64) != 3296 {
		t.Fatalf("refund = %d %v", code, out)
	}
	if code, _, _ := e.do("POST", "/admin/v1/payments/"+bid+"/refund", ad, map[string]any{"reason": "ซ้ำ"}); code != 409 {
		t.Errorf("double refund = %d, want 409", code)
	}
	_, w1, _ := e.do("GET", "/v1/wallet", tr, nil)
	if w1["balance"].(float64) != w0["balance"].(float64)+3296 {
		t.Errorf("wallet %v -> %v", w0["balance"], w1["balance"])
	}
	_, after, _ := e.do("GET", "/admin/v1/payments?status=refunded&q="+b["bookingNumber"].(string), ad, nil)
	if len(pitems(after)) != 1 {
		t.Errorf("refunded payment not listed: %v", after)
	}
	if _, rec, _ := e.do("GET", "/admin/v1/payments/reconciliation", ad, nil); rec["balanced"] != true {
		t.Errorf("books should still balance after a refund: %v", rec)
	}
	e.requireAudit(ad, "admin.payment.refund")
}

func TestReconcileClassifiesMismatches(t *testing.T) {
	f := func(v float64) *float64 { return &v }
	cases := []struct {
		name string
		in   reconcileInput
		want string
	}{
		{"ok paid", reconcileInput{Status: "confirmed", Total: 100, HasPayment: true, PayStatus: "succeeded", PayAmount: 100}, ""},
		{"ok refunded", reconcileInput{Status: "cancelled", Total: 100, Refund: f(100), HasPayment: true, PayStatus: "refunded", PayAmount: 100}, ""},
		{"ok cancelled unpaid", reconcileInput{Status: "cancelled", Total: 100, Refund: f(0)}, ""},
		{"missing", reconcileInput{Status: "confirmed", Total: 100}, "missing_payment"},
		{"not succeeded", reconcileInput{Status: "completed", Total: 100, HasPayment: true, PayStatus: "failed", PayAmount: 100}, "payment_not_succeeded"},
		{"amount", reconcileInput{Status: "confirmed", Total: 100, HasPayment: true, PayStatus: "succeeded", PayAmount: 90}, "amount_mismatch"},
		{"paid but cancelled without refund", reconcileInput{Status: "cancelled", Total: 100, Refund: f(0), HasPayment: true, PayStatus: "succeeded", PayAmount: 100}, "refund_mismatch"},
		{"partial refund", reconcileInput{Status: "cancelled", Total: 100, Refund: f(40), HasPayment: true, PayStatus: "refunded", PayAmount: 100}, "refund_mismatch"},
		{"unexpected", reconcileInput{Status: "pending_payment", Total: 100, HasPayment: true, PayStatus: "pending", PayAmount: 100}, "unexpected_payment"},
	}
	for _, c := range cases {
		got := classifyReconcile(c.in)
		switch {
		case c.want == "" && len(got) != 0:
			t.Errorf("%s: unexpected %v", c.name, got)
		case c.want != "" && (len(got) != 1 || got[0].Kind != c.want):
			t.Errorf("%s: got %v, want %s", c.name, got, c.want)
		}
	}
}
