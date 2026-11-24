package api

import (
	"encoding/json"
	"github.com/Bas-travel/tep/services/api-go/internal/config"
	"io"
	"net/http"
	"testing"
)

// The demo dataset must give every module and every menu of the four roles 20-30+ rows to work with.
func TestDemoDatasetCoversEveryMenu(t *testing.T) {
	e := newEnvCfg(t, nil, func(c *config.Config) { c.DemoDataset = true })
	admin, staff, partner := e.loginAs("admin@tep.dev"), e.loginAs("staff@tep.dev"), e.loginAs("partner@tep.dev")
	traveler := e.login()

	// rows decodes any list response (bare array or {items|transactions|...}) and returns its row count and total.
	rows := func(tok, path string) (n, total int) {
		t.Helper()
		req, _ := http.NewRequest("GET", e.ts.URL+path, nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		raw, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != 200 {
			t.Fatalf("%s = %d %s", path, resp.StatusCode, raw)
		}
		var v any
		if err := json.Unmarshal(raw, &v); err != nil {
			t.Fatal(err)
		}
		switch x := v.(type) {
		case []any:
			return len(x), len(x)
		case map[string]any:
			for _, k := range []string{"items", "transactions", "posts", "groups"} {
				if a, ok := x[k].([]any); ok {
					n = len(a)
				}
			}
			if f, ok := x["total"].(float64); ok {
				total = int(f)
			} else {
				total = n
			}
		}
		return
	}

	// Portal menus report a total.
	for _, c := range []struct {
		name, tok, path string
		min             int
	}{
		{"admin users", admin, "/admin/v1/users?limit=1", 30},
		{"admin partners", admin, "/admin/v1/partners?limit=1", 8},
		{"admin listings", admin, "/admin/v1/listings?limit=1", 30},
		{"admin audit logs", admin, "/admin/v1/audit-logs?limit=1", 20},
		{"staff bookings", staff, "/staff/v1/bookings?limit=1", 26},
		{"staff checked", staff, "/staff/v1/bookings?limit=1&checked=true", 20},
		{"staff tickets", staff, "/staff/v1/tickets?limit=1", 26},
		{"partner listings", partner, "/partner/v1/listings?limit=1", 7},
		{"partner bookings", partner, "/partner/v1/bookings?limit=1", 1},
	} {
		if _, total := rows(c.tok, c.path); total < c.min {
			t.Errorf("%s: total %d < %d", c.name, total, c.min)
		}
	}

	// Traveler menus.
	for _, c := range []struct {
		name, path string
		min        int
	}{
		{"bookings", "/v1/bookings", 15},
		{"trips", "/v1/trips/saved", 12},
		{"coupons", "/v1/coupons", 15},
		{"notifications", "/v1/notifications", 24},
		{"faqs", "/v1/support/faqs", 20},
		{"my reviews", "/v1/reviews", 12},
		{"my tickets", "/v1/support/tickets", 8},
		{"wallet", "/v1/wallet", 20},
		{"listings", "/v1/catalog/listings?limit=50", 25},
		{"groups", "/v1/community/groups", 10},
		{"feed", "/v1/community/feed", 20},
	} {
		if n, _ := rows(traveler, c.path); n < c.min {
			t.Errorf("%s: %d rows < %d", c.name, n, c.min)
		}
	}

	// Every role sees the data for its own menus; a second seed run is a no-op (idempotent).
	_, before := rows(admin, "/admin/v1/users?limit=1")
	if err := e.s.seedDemo(t.Context()); err != nil {
		t.Fatal(err)
	}
	if _, after := rows(admin, "/admin/v1/users?limit=1"); after != before {
		t.Fatalf("seed not idempotent: users %d -> %d", before, after)
	}
	// the seeded money trail must reconcile and the payments console must have 20+ rows
	if _, rec, _ := e.do("GET", "/admin/v1/payments/reconciliation", admin, nil); rec["balanced"] != true {
		t.Errorf("demo data does not reconcile: %v", rec)
	}
	if _, an, _ := e.do("GET", "/admin/v1/analytics?days=365", admin, nil); an["totals"].(map[string]any)["bookings"].(float64) < 10 {
		t.Errorf("analytics shows too few paid bookings: %v", an["totals"])
	}
	if n := e.itemsTotal(admin, "/admin/v1/payments"); n < 20 {
		t.Errorf("payments console has %d rows, want >= 20", n)
	}
}
