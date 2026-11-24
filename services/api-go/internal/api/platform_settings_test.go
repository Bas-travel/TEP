package api

import "testing"

func TestPlatformSettings(t *testing.T) {
	e := newEnv(t, nil)
	ad, tr := e.loginAs("admin@tep.dev"), e.login()
	if code, _, _ := e.do("GET", "/admin/v1/settings", tr, nil); code != 403 {
		t.Errorf("traveler list = %d, want 403", code)
	}
	if code, _, _ := e.do("PUT", "/admin/v1/settings/service_fee_percent", tr, map[string]any{"value": "5"}); code != 403 {
		t.Errorf("traveler put = %d, want 403", code)
	}
	code, list, _ := e.do("GET", "/admin/v1/settings", ad, nil)
	if code != 200 || len(pitems(list)) != 3 {
		t.Fatalf("list = %d %v", code, list)
	}
	if _, cfg, _ := e.do("GET", "/v1/config", "", nil); cfg["serviceFeePercent"].(float64) != 3 || cfg["maintenanceBanner"] != "" {
		t.Fatalf("default public config = %v", cfg)
	}
	long := ""
	for i := 0; i < 201; i++ {
		long += "ก"
	}
	for name, c := range map[string]struct{ key, val string }{
		"fee too high": {"service_fee_percent", "21"}, "fee negative": {"service_fee_percent", "-1"}, "fee text": {"service_fee_percent", "abc"},
		"bad email": {"support_email", "not-an-email"}, "banner too long": {"maintenance_banner", long},
	} {
		if code, _, _ := e.do("PUT", "/admin/v1/settings/"+c.key, ad, map[string]any{"value": c.val}); code != 422 {
			t.Errorf("%s: want 422, got %d", name, code)
		}
	}
	if code, _, _ := e.do("PUT", "/admin/v1/settings/unknown_key", ad, map[string]any{"value": "x"}); code != 404 {
		t.Errorf("unknown key = %d, want 404", code)
	}

	// the fee drives new bookings: 5% of 3200 = 160
	if code, m, _ := e.do("PUT", "/admin/v1/settings/service_fee_percent", ad, map[string]any{"value": "5"}); code != 200 || m["value"] != "5" {
		t.Fatalf("set fee = %d %v", code, m)
	}
	_, b, _ := e.do("POST", "/v1/bookings", tr, map[string]any{"lines": []any{map[string]any{"listingId": 1, "quantity": 1}}})
	if b["amount"].(float64) != 3360 {
		t.Errorf("booking amount with 5%% fee = %v, want 3360", b["amount"])
	}
	e.do("PUT", "/admin/v1/settings/service_fee_percent", ad, map[string]any{"value": "0"})
	_, b0, _ := e.do("POST", "/v1/bookings", tr, map[string]any{"lines": []any{map[string]any{"listingId": 1, "quantity": 1}}})
	if b0["amount"].(float64) != 3200 {
		t.Errorf("booking amount with 0%% fee = %v, want 3200", b0["amount"])
	}

	if code, _, _ := e.do("PUT", "/admin/v1/settings/maintenance_banner", ad, map[string]any{"value": "  ปิดปรับปรุงคืนวันอาทิตย์  "}); code != 200 {
		t.Errorf("set banner = %d", code)
	}
	if code, _, _ := e.do("PUT", "/admin/v1/settings/support_email", ad, map[string]any{"value": "help@sabai.example"}); code != 200 {
		t.Errorf("set email = %d", code)
	}
	_, cfg, _ := e.do("GET", "/v1/config", "", nil)
	if cfg["maintenanceBanner"] != "ปิดปรับปรุงคืนวันอาทิตย์" || cfg["supportEmail"] != "help@sabai.example" || cfg["serviceFeePercent"].(float64) != 0 {
		t.Errorf("public config = %v", cfg)
	}
	_, list, _ = e.do("GET", "/admin/v1/settings", ad, nil)
	for _, it := range pitems(list) {
		m := it.(map[string]any)
		if m["key"] == "service_fee_percent" && (m["value"] != "0" || m["default"] != "3" || m["updatedAt"] == nil) {
			t.Errorf("fee row = %v", m)
		}
	}
	e.requireAudit(ad, "admin.setting.update")
}
