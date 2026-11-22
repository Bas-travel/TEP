package api

import (
	"fmt"
	"testing"
)

func TestAdminCoupons(t *testing.T) {
	e := newEnv(t, nil)
	ad, tr := e.loginAs("admin@tep.dev"), e.login()
	base := e.itemsTotal(ad, "/admin/v1/coupons")

	for name, body := range map[string]map[string]any{
		"bad code":         {"code": "x", "title": "ทดสอบ", "percentOff": 10},
		"percent too high": {"code": "TESTPCT", "title": "ทดสอบ", "percentOff": 101},
		"fixed no max":     {"code": "TESTFIX", "title": "ทดสอบ", "percentOff": 0},
		"bad date":         {"code": "TESTDATE", "title": "ทดสอบ", "percentOff": 5, "expiresOn": "31/12/2026"},
	} {
		if code, _, _ := e.do("POST", "/admin/v1/coupons", ad, body); code != 422 {
			t.Errorf("%s: want 422, got %d", name, code)
		}
	}
	if code, _, _ := e.do("POST", "/admin/v1/coupons", tr, map[string]any{"code": "NOPE10", "title": "x1", "percentOff": 10}); code != 403 {
		t.Errorf("traveler create coupon = %d, want 403", code)
	}

	code, c, _ := e.do("POST", "/admin/v1/coupons", ad, map[string]any{"code": "qa10", "title": "ลด 10% ทดสอบ", "percentOff": 10, "minSpend": 500, "maxDiscount": 200, "expiresOn": "2099-01-01"})
	if code != 201 || c["code"] != "QA10" || c["granted"].(float64) != 0 {
		t.Fatalf("create = %d %v", code, c)
	}
	if code, _, _ := e.do("POST", "/admin/v1/coupons", ad, map[string]any{"code": "QA10", "title": "ซ้ำ", "percentOff": 10}); code != 409 {
		t.Errorf("duplicate code = %d, want 409", code)
	}
	if got := e.itemsTotal(ad, "/admin/v1/coupons"); got != base+1 {
		t.Errorf("total %d, want %d", got, base+1)
	}
	if code, m, _ := e.do("PUT", "/admin/v1/coupons/QA10", ad, map[string]any{"title": "ลด 12% ทดสอบ", "percentOff": 12, "maxDiscount": 300}); code != 200 || m["percentOff"].(float64) != 12 {
		t.Errorf("update = %d %v", code, m)
	}
	if code, _, _ := e.do("PUT", "/admin/v1/coupons/NOSUCH", ad, map[string]any{"title": "ไม่มี", "percentOff": 12}); code != 404 {
		t.Errorf("update missing = %d, want 404", code)
	}
	// search
	if code, m, _ := e.do("GET", "/admin/v1/coupons?q=qa10", ad, nil); code != 200 || len(pitems(m)) != 1 {
		t.Errorf("search = %d %v", code, m)
	}
	// a coupon held by a traveler cannot be deleted (WELCOME10 is granted to the demo traveler)
	if code, _, _ := e.do("DELETE", "/admin/v1/coupons/WELCOME10", ad, nil); code != 409 {
		t.Errorf("delete held coupon = %d, want 409", code)
	}
	if code, _, _ := e.do("DELETE", "/admin/v1/coupons/QA10", ad, nil); code != 204 {
		t.Errorf("delete = %d, want 204", code)
	}
	if code, _, _ := e.do("DELETE", "/admin/v1/coupons/QA10", ad, nil); code != 404 {
		t.Errorf("delete again = %d, want 404", code)
	}
	e.requireAudit(ad, "admin.coupon.create", "admin.coupon.update", "admin.coupon.delete")
}

func TestAdminFaqsDriveThePublicList(t *testing.T) {
	e := newEnv(t, nil)
	ad := e.loginAs("admin@tep.dev")
	public := func() []any {
		code, m, _ := e.do("GET", "/v1/support/faqs", "", nil)
		if code != 200 {
			t.Fatalf("public faqs = %d", code)
		}
		return pitems(m)
	}
	before := len(public())
	if code, _, _ := e.do("POST", "/admin/v1/faqs", ad, map[string]any{"question": "?", "answer": ""}); code != 422 {
		t.Errorf("invalid faq = %d, want 422", code)
	}
	code, f, _ := e.do("POST", "/admin/v1/faqs", ad, map[string]any{"question": "ทดสอบคำถามใหม่?", "answer": "คำตอบทดสอบ"})
	if code != 201 {
		t.Fatalf("create = %d %v", code, f)
	}
	id := fmt.Sprint(int(f["id"].(float64)))
	pub := public()
	if len(pub) != before+1 || pub[len(pub)-1].(map[string]any)["question"] != "ทดสอบคำถามใหม่?" {
		t.Fatalf("new FAQ not appended to the public list (%d -> %d)", before, len(pub))
	}
	if code, m, _ := e.do("PUT", "/admin/v1/faqs/"+id, ad, map[string]any{"question": "แก้ไขคำถาม?", "answer": "แก้ไขคำตอบ", "sortNo": 0}); code != 200 || m["sortNo"].(float64) != 0 {
		t.Errorf("update = %d %v", code, m)
	}
	// Public order == admin order (sortNo, then id) and the edited entry carries its new position.
	_, adm, _ := e.do("GET", "/admin/v1/faqs?limit=100", ad, nil)
	pubNow, admNow := public(), pitems(adm)
	if len(pubNow) != len(admNow) {
		t.Fatalf("public %d vs admin %d", len(pubNow), len(admNow))
	}
	for i := range pubNow {
		if pubNow[i].(map[string]any)["id"] != admNow[i].(map[string]any)["id"] {
			t.Fatalf("public/admin order differs at %d", i)
		}
	}
	for i, it := range admNow {
		m := it.(map[string]any)
		if i > 0 && m["sortNo"].(float64) < admNow[i-1].(map[string]any)["sortNo"].(float64) {
			t.Fatalf("admin list not ordered by sortNo at %d", i)
		}
	}
	if code, _, _ := e.do("PUT", "/admin/v1/faqs/99999", ad, map[string]any{"question": "ไม่มี?", "answer": "ไม่มี"}); code != 404 {
		t.Errorf("update missing = %d, want 404", code)
	}
	if code, _, _ := e.do("DELETE", "/admin/v1/faqs/"+id, ad, nil); code != 204 {
		t.Errorf("delete = %d", code)
	}
	if got := len(public()); got != before {
		t.Errorf("public list %d, want %d after delete", got, before)
	}
	e.requireAudit(ad, "admin.faq.create", "admin.faq.update", "admin.faq.delete")
}

func TestAdminModerationQueue(t *testing.T) {
	e := newEnv(t, nil)
	ad, tr, other := e.loginAs("admin@tep.dev"), e.login(), e.loginAs("partner@tep.dev")
	mk := func(text string) string {
		code, p, _ := e.do("POST", "/v1/community/posts", tr, map[string]any{"type": "tip", "destination": "ภูเก็ต", "text": text})
		if code != 201 {
			t.Fatalf("post = %d %v", code, p)
		}
		return fmt.Sprint(int(p["id"].(float64)))
	}
	spam, mild := mk("โฆษณาขายของ ติดต่อไลน์ด่วนที่นี่เลย"), mk("ข้อมูลที่อาจไม่ถูกต้องนัก")
	base := e.itemsTotal(ad, "/admin/v1/reports")
	for _, c := range []struct{ tok, post, reason string }{{other, spam, "spam"}, {ad, spam, "spam"}, {other, mild, "misinformation"}} {
		if code, _, _ := e.do("POST", "/v1/community/posts/"+c.post+"/report", c.tok, map[string]any{"reason": c.reason}); code != 204 {
			t.Fatalf("report %s = %d", c.post, code)
		}
	}
	code, q, _ := e.do("GET", "/admin/v1/reports", ad, nil)
	rows := pitems(q)
	if code != 200 || int(q["total"].(float64)) != base+2 || len(rows) < 2 {
		t.Fatalf("queue = %d %v", code, q)
	}
	top := rows[0].(map[string]any) // most-reported first
	if fmt.Sprint(int(top["postId"].(float64))) != spam || top["reports"].(float64) != 2 {
		t.Fatalf("queue order/grouping wrong: %v", top)
	}
	if code, _, _ := e.do("POST", "/admin/v1/reports/"+mild+"/dismiss", tr, nil); code != 403 {
		t.Errorf("traveler dismiss = %d, want 403", code)
	}
	if code, m, _ := e.do("POST", "/admin/v1/reports/"+mild+"/dismiss", ad, nil); code != 200 || m["dismissed"].(float64) != 1 {
		t.Errorf("dismiss = %d %v", code, m)
	}
	if code, _, _ := e.do("POST", "/admin/v1/reports/"+mild+"/dismiss", ad, nil); code != 404 {
		t.Errorf("dismiss again = %d, want 404", code)
	}
	if code, _, _ := e.do("GET", "/v1/community/posts/"+mild, tr, nil); code != 200 {
		t.Errorf("dismissed post must stay visible, got %d", code)
	}
	if code, _, _ := e.do("POST", "/admin/v1/reports/"+spam+"/remove", ad, nil); code != 204 {
		t.Errorf("remove = %d, want 204", code)
	}
	if code, _, _ := e.do("GET", "/v1/community/posts/"+spam, tr, nil); code != 404 {
		t.Errorf("removed post must be gone, got %d", code)
	}
	if got := e.itemsTotal(ad, "/admin/v1/reports"); got != base {
		t.Errorf("queue total %d, want %d", got, base)
	}
	e.requireAudit(ad, "admin.report.dismiss", "admin.post.remove")
}

// itemsTotal returns the "total" of a paginated admin list.
func (e *env) itemsTotal(tok, path string) int {
	e.t.Helper()
	code, m, _ := e.do("GET", path, tok, nil)
	if code != 200 {
		e.t.Fatalf("%s = %d %v", path, code, m)
	}
	return int(m["total"].(float64))
}

// requireAudit fails unless every action appears in the audit log.
func (e *env) requireAudit(tok string, actions ...string) {
	e.t.Helper()
	code, m, _ := e.do("GET", "/admin/v1/audit-logs?limit=100", tok, nil)
	if code != 200 {
		e.t.Fatalf("audit-logs = %d", code)
	}
	seen := map[string]bool{}
	for _, it := range pitems(m) {
		seen[it.(map[string]any)["action"].(string)] = true
	}
	for _, a := range actions {
		if !seen[a] {
			e.t.Errorf("audit log is missing %q", a)
		}
	}
}
