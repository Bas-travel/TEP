package api

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestAdminAnalytics(t *testing.T) {
	e := newEnv(t, nil)
	ad, tr := e.loginAs("admin@tep.dev"), e.login()
	if code, _, _ := e.do("GET", "/admin/v1/analytics", tr, nil); code != 403 {
		t.Errorf("traveler analytics = %d, want 403", code)
	}
	for _, q := range []string{"?days=0", "?days=366", "?days=abc"} {
		if code, _, _ := e.do("GET", "/admin/v1/analytics"+q, ad, nil); code != 422 {
			t.Errorf("analytics%s = %d, want 422", q, code)
		}
	}
	_, before, _ := e.do("GET", "/admin/v1/analytics", ad, nil)
	b0 := before["totals"].(map[string]any)["bookings"].(float64)
	r0 := before["totals"].(map[string]any)["revenue"].(float64)

	_, b, _ := e.do("POST", "/v1/bookings", tr, map[string]any{"lines": []any{map[string]any{"listingId": 1, "quantity": 1}}})
	_, pi, _ := e.do("POST", "/v1/payments/intents", tr, map[string]any{"bookingId": b["id"], "method": "promptpay"})
	e.now = e.now.Add(4 * time.Second)
	e.do("GET", "/v1/payments/intents/"+pi["id"].(string), tr, nil)

	code, a, _ := e.do("GET", "/admin/v1/analytics?days=7", ad, nil)
	if code != 200 || len(a["series"].([]any)) != 7 || a["days"].(float64) != 7 {
		t.Fatalf("analytics = %d %v", code, a)
	}
	tot := a["totals"].(map[string]any)
	if tot["bookings"].(float64) != b0+1 || tot["revenue"].(float64) != r0+3296 {
		t.Errorf("totals %v, want bookings %v revenue %v", tot, b0+1, r0+3296)
	}
	last := a["series"].([]any)[6].(map[string]any)
	if last["date"] != startOfDay(e.now).Format("2006-01-02") || last["bookings"].(float64) < 1 {
		t.Errorf("today's bucket = %v", last)
	}
	var sum float64
	for _, p := range a["series"].([]any) {
		sum += p.(map[string]any)["revenue"].(float64)
	}
	if round2(sum) != tot["revenue"].(float64) {
		t.Errorf("series sum %v != total %v", sum, tot["revenue"])
	}
	if bt := a["byType"].([]any); len(bt) == 0 {
		t.Error("byType empty after a paid booking")
	}
	if a["byStatus"].(map[string]any)["confirmed"].(float64) < 1 {
		t.Errorf("byStatus = %v", a["byStatus"])
	}

	// CSV export
	req, _ := http.NewRequest("GET", e.ts.URL+"/admin/v1/analytics/export?days=14", nil)
	req.Header.Set("Authorization", "Bearer "+ad)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")
	if resp.StatusCode != 200 || !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/csv") ||
		lines[0] != "date,paid_bookings,revenue_thb,new_users" || len(lines) != 15 {
		t.Errorf("export = %d %q lines=%d", resp.StatusCode, resp.Header.Get("Content-Type"), len(lines))
	}
	req.Header.Set("Authorization", "Bearer "+tr)
	if r2, _ := http.DefaultClient.Do(req); r2.StatusCode != 403 {
		t.Errorf("traveler export = %d, want 403", r2.StatusCode)
	}
}
