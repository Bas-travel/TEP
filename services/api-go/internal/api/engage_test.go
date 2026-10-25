package api

import "testing"

func TestEngage(t *testing.T) {
	e := newEnv(t, nil)
	tok := e.login()
	items := func(m map[string]any) []any { v, _ := m["items"].([]any); return v }

	// 401 without token on every route; 404/422/happy paths below.
	for _, rt := range [][2]string{{"GET", "/v1/me/settings"}, {"PATCH", "/v1/me/settings"}, {"GET", "/v1/coupons"}, {"POST", "/v1/coupons/X/apply"},
		{"GET", "/v1/notifications"}, {"POST", "/v1/notifications/read-all"}, {"POST", "/v1/notifications/1/read"}, {"GET", "/v1/trips/saved"},
		{"POST", "/v1/trips/saved"}, {"GET", "/v1/trips/saved/1"}, {"GET", "/v1/reviews"}, {"POST", "/v1/reviews"}, {"GET", "/v1/support/tickets"},
		{"POST", "/v1/support/tickets"}, {"GET", "/v1/community/feed"}, {"POST", "/v1/community/posts/1/like"}, {"POST", "/v1/ai/itineraries"},
		{"GET", "/v1/ai/itineraries/jobs/x"}} {
		if code, _, _ := e.do(rt[0], rt[1], "", map[string]any{}); code != 401 {
			t.Errorf("%s %s without token = %d", rt[0], rt[1], code)
		}
	}

	t.Run("settings", func(t *testing.T) {
		code, m, _ := e.do("GET", "/v1/me/settings", tok, nil)
		if code != 200 || m["language"] != "th" || m["theme"] != "light" {
			t.Fatal(code, m)
		}
		if code, _, _ = e.do("PATCH", "/v1/me/settings", tok, map[string]any{"theme": "neon"}); code != 422 {
			t.Fatal(code)
		}
		code, m, _ = e.do("PATCH", "/v1/me/settings", tok, map[string]any{"theme": "dark", "notifications": map[string]any{"promo": false}})
		n := m["notifications"].(map[string]any)
		if code != 200 || m["theme"] != "dark" || n["promo"] != false || n["push"] != true {
			t.Fatal(code, m)
		}
	})

	t.Run("coupons", func(t *testing.T) {
		code, m, _ := e.do("GET", "/v1/coupons", tok, nil)
		if code != 200 || len(items(m)) != 3 {
			t.Fatal(code, m)
		}
		code, m, _ = e.do("POST", "/v1/coupons/WELCOME10/apply", tok, map[string]any{"bookingDraftTotal": 3000})
		if code != 200 || m["valid"] != true || m["discount"] != 300.0 {
			t.Fatal(code, m)
		}
		if code, _, _ = e.do("POST", "/v1/coupons/WELCOME10/apply", tok, map[string]any{"bookingDraftTotal": 10}); code != 422 {
			t.Fatal(code)
		}
		if code, _, _ = e.do("POST", "/v1/coupons/OLD50/apply", tok, map[string]any{"bookingDraftTotal": 5000}); code != 422 {
			t.Fatal(code)
		}
		if code, _, _ = e.do("POST", "/v1/coupons/NOPE/apply", tok, map[string]any{"bookingDraftTotal": 5000}); code != 404 {
			t.Fatal(code)
		}
	})

	t.Run("notifications", func(t *testing.T) {
		code, m, _ := e.do("GET", "/v1/notifications", tok, nil)
		if code != 200 || len(items(m)) != 4 {
			t.Fatal(code, m)
		}
		id := items(m)[0].(map[string]any)["id"].(float64)
		if code, _, _ = e.do("POST", "/v1/notifications/"+itoa(id)+"/read", tok, nil); code != 204 {
			t.Fatal(code)
		}
		if code, _, _ = e.do("POST", "/v1/notifications/9999/read", tok, nil); code != 404 {
			t.Fatal(code)
		}
		if code, _, _ = e.do("POST", "/v1/notifications/read-all", tok, nil); code != 204 {
			t.Fatal(code)
		}
		_, m, _ = e.do("GET", "/v1/notifications", tok, nil)
		for _, it := range items(m) {
			if it.(map[string]any)["read"] != true {
				t.Fatal("not all read", m)
			}
		}
	})

	t.Run("ai+trips", func(t *testing.T) {
		if code, _, _ := e.do("POST", "/v1/ai/itineraries", tok, map[string]any{"destination": "ภูเก็ต", "days": 20, "budget": 5000}); code != 422 {
			t.Fatal(code)
		}
		code, m, _ := e.do("POST", "/v1/ai/itineraries", tok, map[string]any{"destination": "ภูเก็ต", "days": 3, "budget": 10000, "travelers": 2})
		if code != 202 || m["status"] != "completed" || m["degraded"] != true {
			t.Fatal(code, m)
		}
		res := m["result"].(map[string]any)
		if len(res["days"].([]any)) != 3 || len(res["budgetBreakdown"].([]any)) != 4 {
			t.Fatal(res)
		}
		if code, _, _ = e.do("GET", "/v1/ai/itineraries/jobs/"+m["jobId"].(string), tok, nil); code != 200 {
			t.Fatal(code)
		}
		if code, _, _ = e.do("GET", "/v1/ai/itineraries/jobs/nope", tok, nil); code != 404 {
			t.Fatal(code)
		}
		if code, _, _ = e.do("POST", "/v1/trips/saved", tok, map[string]any{"destination": "", "days": 3}); code != 422 {
			t.Fatal(code)
		}
		code, tm, _ := e.do("POST", "/v1/trips/saved", tok, map[string]any{"destination": "ภูเก็ต", "days": 3, "budget": 10000, "itinerary": res, "source": "ai"})
		if code != 201 || tm["source"] != "ai" {
			t.Fatal(code, tm)
		}
		if code, got, _ := e.do("GET", "/v1/trips/saved/"+itoa(tm["id"].(float64)), tok, nil); code != 200 || got["destination"] != "ภูเก็ต" {
			t.Fatal(code, got)
		}
		if code, _, _ = e.do("GET", "/v1/trips/saved/9999", tok, nil); code != 404 {
			t.Fatal(code)
		}
		if _, l, _ := e.do("GET", "/v1/trips/saved", tok, nil); len(items(l)) != 1 {
			t.Fatal(l)
		}
	})

	t.Run("reviews", func(t *testing.T) {
		for _, bad := range []map[string]any{{"listingId": 1, "rating": 6, "comment": "x"}, {"listingId": 999, "rating": 5, "comment": "x"}, {"listingId": 1, "rating": 5}} {
			if code, _, _ := e.do("POST", "/v1/reviews", tok, bad); code != 422 {
				t.Fatal(code, bad)
			}
		}
		code, m, _ := e.do("POST", "/v1/reviews", tok, map[string]any{"listingId": 1, "rating": 5, "comment": "ดีมาก"})
		if code != 201 || m["itemName"] != "Andaman Sunset Resort" || m["itemType"] != "hotel" {
			t.Fatal(code, m)
		}
		if _, l, _ := e.do("GET", "/v1/reviews", tok, nil); len(items(l)) != 1 {
			t.Fatal(l)
		}
	})

	t.Run("support", func(t *testing.T) {
		if code, _, _ := e.do("POST", "/v1/support/tickets", tok, map[string]any{"subject": "", "message": "x"}); code != 422 {
			t.Fatal(code)
		}
		if code, _, _ := e.do("POST", "/v1/support/tickets", tok, map[string]any{"subject": "ช่วยด้วย", "message": "ขอเลื่อนวัน"}); code != 201 {
			t.Fatal(code)
		}
		if _, l, _ := e.do("GET", "/v1/support/tickets", tok, nil); len(items(l)) != 1 {
			t.Fatal(l)
		}
	})

	t.Run("community", func(t *testing.T) {
		_, m, _ := e.do("GET", "/v1/community/feed", tok, nil)
		if len(items(m)) != 5 {
			t.Fatal(m)
		}
		_, m, _ = e.do("GET", "/v1/community/feed?destination=ภูเก็ต", tok, nil)
		if len(items(m)) != 1 {
			t.Fatal(m)
		}
		_, m, _ = e.do("GET", "/v1/community/feed?mode=following", tok, nil)
		if len(items(m)) != 3 {
			t.Fatal(m)
		}
		if code, _, _ := e.do("GET", "/v1/community/feed?mode=x", tok, nil); code != 422 {
			t.Fatal(code)
		}
		code, m, _ := e.do("POST", "/v1/community/posts/1/like", tok, nil)
		if code != 200 || m["liked"] != true || m["likeCount"] != 43.0 {
			t.Fatal(code, m)
		}
		_, m, _ = e.do("POST", "/v1/community/posts/1/like", tok, nil)
		if m["liked"] != false || m["likeCount"] != 42.0 {
			t.Fatal(m)
		}
		if code, _, _ := e.do("POST", "/v1/community/posts/999/like", tok, nil); code != 404 {
			t.Fatal(code)
		}
	})
}
