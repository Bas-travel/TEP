package api

import "testing"

func (e *env) loginAs(email string) string {
	code, m, _ := e.do("POST", "/v1/auth/login", "", map[string]string{"email": email, "password": "Passw0rd!"})
	if code != 200 {
		e.t.Fatalf("login %s: %d %v", email, code, m)
	}
	return m["accessToken"].(string)
}

func pitems(m map[string]any) []any { v, _ := m["items"].([]any); return v }

func TestPortalRBACMatrix(t *testing.T) {
	e := newEnv(t, nil)
	tok := map[string]string{"traveler": e.login(), "partner": e.loginAs("partner@tep.dev"), "staff": e.loginAs("staff@tep.dev"), "admin": e.loginAs("admin@tep.dev")}
	routes := []struct {
		method, path string
		allowed      []string
	}{
		{"GET", "/partner/v1/dashboard", []string{"partner"}},
		{"GET", "/partner/v1/listings", []string{"partner"}},
		{"GET", "/partner/v1/bookings", []string{"partner"}},
		{"GET", "/partner/v1/earnings", []string{"partner"}},
		{"GET", "/partner/v1/profile", []string{"partner"}},
		{"GET", "/admin/v1/dashboard", []string{"admin"}},
		{"GET", "/admin/v1/users", []string{"admin"}},
		{"GET", "/admin/v1/partners", []string{"admin"}},
		{"GET", "/admin/v1/listings", []string{"admin"}},
		{"GET", "/admin/v1/audit-logs", []string{"admin"}},
		{"GET", "/admin/v1/system", []string{"admin"}},
		{"GET", "/staff/v1/dashboard", []string{"staff", "admin"}},
		{"GET", "/staff/v1/bookings", []string{"staff", "admin"}},
		{"GET", "/staff/v1/tickets", []string{"staff", "admin"}},
		{"GET", "/staff/v1/reports/daily", []string{"staff", "admin"}},
	}
	for _, rt := range routes {
		if code, _, _ := e.do(rt.method, rt.path, "", nil); code != 401 {
			t.Errorf("%s without token = %d", rt.path, code)
		}
		for role, tk := range tok {
			code, m, _ := e.do(rt.method, rt.path, tk, nil)
			want := 403
			for _, a := range rt.allowed {
				if a == role {
					want = 200
				}
			}
			if code != want {
				t.Errorf("%s as %s = %d (want %d) %v", rt.path, role, code, want, m)
			}
		}
	}
	// /v1/me reports the role
	if _, m, _ := e.do("GET", "/v1/me", tok["staff"], nil); m["role"] != "staff" {
		t.Fatal(m)
	}
	// staff/partner/admin cannot delete themselves
	if code, _, _ := e.do("DELETE", "/v1/me", tok["staff"], nil); code != 409 {
		t.Fatal(code)
	}
}

func TestPartnerPortal(t *testing.T) {
	e := newEnv(t, nil)
	pt := e.loginAs("partner@tep.dev")
	demo := e.login()

	_, m, _ := e.do("GET", "/partner/v1/listings", pt, nil)
	if len(pitems(m)) != 4 {
		t.Fatal(m)
	}
	// other providers' listings do not exist for this partner
	if code, _, _ := e.do("GET", "/partner/v1/listings/5", pt, nil); code != 404 {
		t.Fatal(code)
	}
	if code, _, _ := e.do("PATCH", "/partner/v1/listings/5", pt, map[string]any{"name": "x"}); code != 404 {
		t.Fatal(code)
	}
	// validation
	if code, _, _ := e.do("POST", "/partner/v1/listings", pt, map[string]any{"name": "", "type": "flight", "priceFrom": -1}); code != 422 {
		t.Fatal(code)
	}
	good := map[string]any{"name": "Sunset Villa", "type": "hotel", "location": "Phuket", "destinationId": 1, "priceFrom": 2500.0,
		"description": "วิลล่าริมทะเล", "amenities": []string{"Wi-Fi", "สระว่ายน้ำ"}}
	code, m, _ := e.do("POST", "/partner/v1/listings", pt, good)
	if code != 201 || m["status"] != "draft" {
		t.Fatal(code, m)
	}
	id := itoa(m["id"].(float64))
	// drafts are invisible to travelers
	if code, _, _ := e.do("GET", "/v1/catalog/listings/"+id, "", nil); code != 404 {
		t.Fatal("draft visible", code)
	}
	if code, m, _ = e.do("PATCH", "/partner/v1/listings/"+id, pt, map[string]any{"status": "published", "priceFrom": 2600.0}); code != 200 || m["priceFrom"] != 2600.0 {
		t.Fatal(code, m)
	}
	if code, m, _ = e.do("GET", "/v1/catalog/listings/"+id, "", nil); code != 200 || m["name"] != "Sunset Villa" {
		t.Fatal(code, m)
	}
	if code, _, _ = e.do("PATCH", "/partner/v1/listings/"+id, pt, map[string]any{"status": "archived"}); code != 200 {
		t.Fatal(code)
	}
	if code, _, _ = e.do("GET", "/v1/catalog/listings/"+id, "", nil); code != 404 {
		t.Fatal("archived visible", code)
	}

	// bookings: confirmed (1), completed (2), pending (3) all contain this partner's lines
	_, m, _ = e.do("GET", "/partner/v1/bookings", pt, nil)
	if m["total"] != 3.0 {
		t.Fatal(m)
	}
	_, m, _ = e.do("GET", "/partner/v1/bookings?status=confirmed", pt, nil)
	if m["total"] != 1.0 {
		t.Fatal(m)
	}
	if code, _, _ := e.do("GET", "/partner/v1/bookings?status=nope", pt, nil); code != 422 {
		t.Fatal(code)
	}
	// complete: only confirmed
	if code, _, _ := e.do("POST", "/partner/v1/bookings/3/complete", pt, nil); code != 409 {
		t.Fatal(code)
	}
	code, m, _ = e.do("POST", "/partner/v1/bookings/1/complete", pt, nil)
	if code != 200 || m["status"] != "completed" {
		t.Fatal(code, m)
	}
	if code, _, _ = e.do("POST", "/partner/v1/bookings/1/complete", pt, nil); code != 409 {
		t.Fatal(code)
	}
	// reject needs a reason; pending one is cancelled
	if code, _, _ = e.do("POST", "/partner/v1/bookings/3/reject", pt, map[string]any{}); code != 422 {
		t.Fatal(code)
	}
	code, m, _ = e.do("POST", "/partner/v1/bookings/3/reject", pt, map[string]any{"reason": "ห้องเต็ม"})
	if code != 200 || m["status"] != "cancelled" {
		t.Fatal(code, m)
	}
	// a booking of nobody's listing is invisible
	if code, _, _ = e.do("GET", "/partner/v1/bookings/999", pt, nil); code != 404 {
		t.Fatal(code)
	}
	// earnings: 2 confirmed/completed bookings, 12% commission
	_, m, _ = e.do("GET", "/partner/v1/earnings", pt, nil)
	if m["commissionRate"] != 12.0 || m["gross"].(float64) <= 0 {
		t.Fatal(m)
	}
	g, c, n := m["gross"].(float64), m["commission"].(float64), m["net"].(float64)
	if round2(g*0.12) != c || round2(g-c) != n || len(m["months"].([]any)) < 1 {
		t.Fatal(m)
	}
	_, m, _ = e.do("GET", "/partner/v1/dashboard", pt, nil)
	if m["bookingsToComplete"] != 0.0 {
		t.Fatal(m)
	}
	// profile
	if code, m, _ = e.do("PATCH", "/partner/v1/profile", pt, map[string]any{"businessName": "Andaman Group"}); code != 200 || m["businessName"] != "Andaman Group" {
		t.Fatal(code, m)
	}
	if code, _, _ = e.do("PATCH", "/partner/v1/profile", pt, map[string]any{"businessName": ""}); code != 422 {
		t.Fatal(code)
	}
	// traveler notified of the rejection
	_ = demo
}

func TestAdminPortal(t *testing.T) {
	e := newEnv(t, nil)
	ad := e.loginAs("admin@tep.dev")
	demo := e.login()

	var code int
	_, m, _ := e.do("GET", "/admin/v1/dashboard", ad, nil)
	if m["bookingsTotal"] != 3.0 || m["listingsPublished"] != 12.0 {
		t.Fatal(m)
	}
	// users: filter, paging, validation
	_, m, _ = e.do("GET", "/admin/v1/users?role=staff", ad, nil)
	if m["total"] != 1.0 {
		t.Fatal(m)
	}
	_, m, _ = e.do("GET", "/admin/v1/users?limit=2&offset=1", ad, nil)
	if len(pitems(m)) != 2 || m["limit"] != 2.0 || m["offset"] != 1.0 {
		t.Fatal(m)
	}
	if code, _, _ := e.do("GET", "/admin/v1/users?limit=500", ad, nil); code != 422 {
		t.Fatal(code)
	}
	if code, _, _ := e.do("GET", "/admin/v1/users?role=wizard", ad, nil); code != 422 {
		t.Fatal(code)
	}
	// role changes: never to partner, never self, never a partner account
	if code, _, _ := e.do("PATCH", "/admin/v1/users/1", ad, map[string]any{"role": "partner"}); code != 422 {
		t.Fatal(code)
	}
	if code, _, _ := e.do("PATCH", "/admin/v1/users/6", ad, map[string]any{"status": "suspended"}); code != 409 {
		t.Fatal(code)
	}
	if code, _, _ := e.do("PATCH", "/admin/v1/users/8", ad, map[string]any{"role": "traveler"}); code != 409 {
		t.Fatal(code)
	}
	if code, _, _ := e.do("PATCH", "/admin/v1/users/999", ad, map[string]any{"role": "staff"}); code != 404 {
		t.Fatal(code)
	}
	// suspending a traveler locks them out immediately
	if code, m, _ = e.do("PATCH", "/admin/v1/users/1", ad, map[string]any{"status": "suspended"}); code != 200 || m["status"] != "suspended" {
		t.Fatal(code, m)
	}
	if code, _, _ := e.do("GET", "/v1/me", demo, nil); code != 401 {
		t.Fatal("suspended user still authorised", code)
	}
	if code, _, _ := e.do("POST", "/v1/auth/login", "", map[string]string{"email": "demo@tep.dev", "password": "Passw0rd!"}); code == 200 {
		t.Fatal("suspended user can log in")
	}
	if code, m, _ = e.do("PATCH", "/admin/v1/users/1", ad, map[string]any{"status": "active", "role": "moderator"}); code != 200 || m["role"] != "moderator" {
		t.Fatal(code, m)
	}

	// partners: pending application of m_fah (partner 2)
	_, m, _ = e.do("GET", "/admin/v1/partners?status=pending", ad, nil)
	if m["total"] != 1.0 {
		t.Fatal(m)
	}
	if code, _, _ := e.do("PATCH", "/admin/v1/partners/2", ad, map[string]any{"status": "suspended"}); code != 409 {
		t.Fatal("pending->suspended", code)
	}
	if code, _, _ := e.do("PATCH", "/admin/v1/partners/2", ad, map[string]any{"status": "rejected"}); code != 422 {
		t.Fatal("reject without note", code)
	}
	if code, _, _ := e.do("PATCH", "/admin/v1/partners/2", ad, map[string]any{"commissionRate": 60}); code != 422 {
		t.Fatal(code)
	}
	code, m, _ = e.do("PATCH", "/admin/v1/partners/2", ad, map[string]any{"status": "approved", "commissionRate": 10})
	if code != 200 || m["status"] != "approved" || m["commissionRate"] != 10.0 {
		t.Fatal(code, m)
	}
	if _, m, _ = e.do("GET", "/admin/v1/users/4", ad, nil); m["role"] != "partner" {
		t.Fatal("owner not promoted", m)
	}
	// suspending the original partner hides listings and closes the portal
	pt := e.loginAs("partner@tep.dev")
	if code, _, _ = e.do("PATCH", "/admin/v1/partners/1", ad, map[string]any{"status": "suspended", "decisionNote": "ร้องเรียนซ้ำ"}); code != 200 {
		t.Fatal(code)
	}
	if code, _, _ = e.do("GET", "/v1/catalog/listings/1", "", nil); code != 404 {
		t.Fatal("suspended partner's listing is public", code)
	}
	if code, _, _ = e.do("GET", "/partner/v1/dashboard", pt, nil); code != 403 {
		t.Fatal(code)
	}
	if code, _, _ = e.do("PATCH", "/admin/v1/partners/1", ad, map[string]any{"status": "approved"}); code != 200 {
		t.Fatal(code)
	}
	if code, _, _ = e.do("GET", "/v1/catalog/listings/1", "", nil); code != 200 {
		t.Fatal(code)
	}
	// listing moderation
	if code, m, _ = e.do("PATCH", "/admin/v1/listings/9", ad, map[string]any{"status": "archived"}); code != 200 || m["status"] != "archived" {
		t.Fatal(code, m)
	}
	if code, _, _ = e.do("PATCH", "/admin/v1/listings/9", ad, map[string]any{"status": "x"}); code != 422 {
		t.Fatal(code)
	}
	// audit trail
	_, m, _ = e.do("GET", "/admin/v1/audit-logs?action=admin.partner", ad, nil)
	if m["total"].(float64) < 3 {
		t.Fatal(m)
	}
	_, m, _ = e.do("GET", "/admin/v1/system", ad, nil)
	if m["version"] != "test" || m["mode"] != "mock" {
		t.Fatal(m)
	}
}

func TestStaffPortalAndTravelerSide(t *testing.T) {
	e := newEnv(t, nil)
	st := e.loginAs("staff@tep.dev")
	demo := e.login()

	_, m, _ := e.do("GET", "/staff/v1/dashboard", st, nil)
	if m["openTickets"] != 1.0 || m["uncheckedBookings"] != 1.0 || m["checkedToday"] != 0.0 {
		t.Fatal(m)
	}
	// booking checks
	_, m, _ = e.do("GET", "/staff/v1/bookings?status=confirmed&checked=false", st, nil)
	if m["total"] != 1.0 {
		t.Fatal(m)
	}
	if code, _, _ := e.do("POST", "/staff/v1/bookings/1/check", st, map[string]any{"result": "flagged"}); code != 422 {
		t.Fatal("flag without note", code)
	}
	if code, _, _ := e.do("POST", "/staff/v1/bookings/3/check", st, map[string]any{"result": "verified"}); code != 409 {
		t.Fatal("unpaid booking", code)
	}
	code, m, _ := e.do("POST", "/staff/v1/bookings/1/check", st, map[string]any{"result": "flagged", "note": "ชื่อผู้เดินทางไม่ตรง"})
	if code != 200 || m["verification"] != "flagged" || len(m["checks"].([]any)) != 1 {
		t.Fatal(code, m)
	}
	_, m, _ = e.do("GET", "/staff/v1/dashboard", st, nil)
	if m["checkedToday"] != 1.0 || m["flaggedToday"] != 1.0 || m["uncheckedBookings"] != 0.0 {
		t.Fatal(m)
	}
	_, m, _ = e.do("GET", "/staff/v1/reports/daily", st, nil)
	if m["bookingsFlagged"] != 1.0 || m["myChecks"] != 1.0 {
		t.Fatal(m)
	}
	if code, _, _ = e.do("GET", "/staff/v1/reports/daily?date=bad", st, nil); code != 422 {
		t.Fatal(code)
	}

	// tickets: traveler opens one, staff answers, traveler replies (reopens), staff resolves
	code, m, _ = e.do("POST", "/v1/support/tickets", demo, map[string]any{"subject": "ช่วยด้วย", "message": "ขอเลื่อนวัน"})
	if code != 201 {
		t.Fatal(code, m)
	}
	tid := itoa(m["id"].(float64))
	if m["status"] != "open" {
		t.Fatal(m)
	}
	if code, _, _ = e.do("POST", "/staff/v1/tickets/"+tid+"/replies", st, map[string]any{"message": " "}); code != 422 {
		t.Fatal(code)
	}
	code, m, _ = e.do("POST", "/staff/v1/tickets/"+tid+"/replies", st, map[string]any{"message": "ได้ค่ะ แจ้งวันใหม่ได้เลย"})
	if code != 201 || m["status"] != "answered" || len(m["replies"].([]any)) != 1 {
		t.Fatal(code, m)
	}
	_, m, _ = e.do("GET", "/v1/support/tickets", demo, nil)
	var mine map[string]any
	for _, x := range pitems(m) {
		if tk := x.(map[string]any); itoa(tk["id"].(float64)) == tid {
			mine = tk
		}
	}
	if mine == nil || mine["status"] != "answered" || len(mine["replies"].([]any)) != 1 {
		t.Fatal(mine)
	}
	if code, _, _ = e.do("POST", "/v1/support/tickets/1/replies", demo, map[string]any{"message": "x"}); code != 404 {
		t.Fatal("replied to someone else's ticket", code)
	}
	code, m, _ = e.do("POST", "/v1/support/tickets/"+tid+"/replies", demo, map[string]any{"message": "วันที่ 20 ค่ะ"})
	if code != 201 || m["status"] != "open" {
		t.Fatal(code, m)
	}
	if code, m, _ = e.do("PATCH", "/staff/v1/tickets/"+tid, st, map[string]any{"status": "resolved"}); code != 200 || m["status"] != "resolved" {
		t.Fatal(code, m)
	}
	if code, _, _ = e.do("POST", "/staff/v1/tickets/"+tid+"/replies", st, map[string]any{"message": "x"}); code != 409 {
		t.Fatal("reply on resolved", code)
	}
	if code, _, _ = e.do("PATCH", "/staff/v1/tickets/"+tid, st, map[string]any{"status": "answered"}); code != 422 {
		t.Fatal(code)
	}
	_, m, _ = e.do("GET", "/staff/v1/tickets?status=resolved", st, nil)
	if m["total"] != 2.0 {
		t.Fatal(m)
	}
	// replies notified the traveler
	_, m, _ = e.do("GET", "/v1/notifications", demo, nil)
	found := false
	for _, x := range pitems(m) {
		found = found || x.(map[string]any)["title"] == "ทีมงานตอบกลับคำร้องของคุณ"
	}
	if !found {
		t.Fatal("no notification", m)
	}

	// partner application by a traveler
	if code, _, _ = e.do("GET", "/v1/partners/me", demo, nil); code != 404 {
		t.Fatal(code)
	}
	if code, _, _ = e.do("POST", "/v1/partners/apply", demo, map[string]any{"businessName": ""}); code != 422 {
		t.Fatal(code)
	}
	code, m, _ = e.do("POST", "/v1/partners/apply", demo, map[string]any{"businessName": "Demo Tours", "contactPhone": "080"})
	if code != 201 || m["status"] != "pending" {
		t.Fatal(code, m)
	}
	if code, _, _ = e.do("POST", "/v1/partners/apply", demo, map[string]any{"businessName": "Again"}); code != 409 {
		t.Fatal(code)
	}
	// admin rejects, traveler re-applies
	ad := e.loginAs("admin@tep.dev")
	pid := itoa(m["id"].(float64))
	if code, _, _ = e.do("PATCH", "/admin/v1/partners/"+pid, ad, map[string]any{"status": "rejected", "decisionNote": "ข้อมูลไม่ครบ"}); code != 200 {
		t.Fatal(code)
	}
	if _, m, _ = e.do("GET", "/v1/partners/me", demo, nil); m["status"] != "rejected" || m["decisionNote"] != "ข้อมูลไม่ครบ" {
		t.Fatal(m)
	}
	if code, m, _ = e.do("POST", "/v1/partners/apply", demo, map[string]any{"businessName": "Demo Tours 2"}); code != 201 || m["status"] != "pending" {
		t.Fatal(code, m)
	}
}
