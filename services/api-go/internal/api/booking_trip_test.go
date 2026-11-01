package api

import "testing"

func TestBookingStoresDatesAndGuests(t *testing.T) {
	e := newEnv(t, nil)
	tok := e.login()
	start := e.now.AddDate(0, 0, 3).Format("2006-01-02")
	end := e.now.AddDate(0, 0, 5).Format("2006-01-02")
	code, b, _ := e.do("POST", "/v1/bookings", tok, map[string]any{
		"lines": []map[string]any{{"listingId": 1, "quantity": 1}}, "startDate": start, "endDate": end, "guests": 3})
	if code != 201 {
		t.Fatalf("create = %d %v", code, b)
	}
	if b["startDate"] != start || b["endDate"] != end || b["guests"].(float64) != 3 {
		t.Fatalf("not echoed: %v", b)
	}
	id := b["id"].(float64)
	_, got, _ := e.do("GET", "/v1/bookings/"+itoa(id), tok, nil)
	if got["startDate"] != start || got["guests"].(float64) != 3 {
		t.Fatalf("not persisted on read: %v", got)
	}
}

func TestBookingDefaultsToOneGuestWithoutDates(t *testing.T) {
	e := newEnv(t, nil)
	tok := e.login()
	code, b, _ := e.do("POST", "/v1/bookings", tok, map[string]any{"lines": []map[string]any{{"listingId": 1}}})
	if code != 201 || b["guests"].(float64) != 1 {
		t.Fatalf("got %d %v", code, b)
	}
	if _, has := b["startDate"]; has {
		t.Fatalf("startDate must be omitted when not sent: %v", b)
	}
}

func TestBookingRejectsBadDatesAndGuests(t *testing.T) {
	e := newEnv(t, nil)
	tok := e.login()
	past := e.now.AddDate(0, 0, -2).Format("2006-01-02")
	ok := e.now.AddDate(0, 0, 4).Format("2006-01-02")
	before := e.now.AddDate(0, 0, 2).Format("2006-01-02")
	cases := []struct {
		name  string
		body  map[string]any
		field string
	}{
		{"past start", map[string]any{"startDate": past}, "startDate"},
		{"garbage start", map[string]any{"startDate": "tomorrow"}, "startDate"},
		{"end before start", map[string]any{"startDate": ok, "endDate": before}, "endDate"},
		{"end without start", map[string]any{"endDate": ok}, "endDate"},
		{"too many guests", map[string]any{"guests": 21}, "guests"},
		{"negative guests", map[string]any{"guests": -1}, "guests"},
	}
	for _, c := range cases {
		c.body["lines"] = []map[string]any{{"listingId": 1}}
		code, p, _ := e.do("POST", "/v1/bookings", tok, c.body)
		if code != 422 {
			t.Errorf("%s: status %d", c.name, code)
			continue
		}
		found := false
		for _, x := range p["errors"].([]any) {
			if x.(map[string]any)["field"] == c.field {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: no error on %s: %v", c.name, c.field, p)
		}
	}
}
