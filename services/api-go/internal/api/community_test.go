package api

import (
	"strings"
	"testing"
)

func list(m map[string]any) []any { v, _ := m["items"].([]any); return v }

func registerUser(e *env, email string) string {
	_, m, _ := e.do("POST", "/v1/auth/register", "", map[string]any{"email": email, "password": "longenough", "firstName": "ผู้ใช้", "lastName": "ทดสอบ"})
	return m["accessToken"].(string)
}

func TestCommunityPostsLifecycle(t *testing.T) {
	e := newEnv(t, nil)
	tok := e.login()
	other := registerUser(e, "other@example.com")

	// validation
	for name, body := range map[string]map[string]any{
		"bad type":  {"type": "trip", "text": "ข้อความยาวพอสมควรแล้ว"},
		"too short": {"type": "story", "text": "สั้น"},
		"bad group": {"type": "story", "text": "ข้อความยาวพอสมควรแล้ว", "groupId": 999},
		"bad image": {"type": "story", "text": "ข้อความยาวพอสมควรแล้ว", "imageObjectKey": "../etc/passwd"},
		"long dest": {"type": "story", "text": "ข้อความยาวพอสมควรแล้ว", "destination": strings.Repeat("x", 101)},
	} {
		if code, _, _ := e.do("POST", "/v1/community/posts", tok, body); code != 422 {
			t.Errorf("%s: status %d", name, code)
		}
	}

	code, p, _ := e.do("POST", "/v1/community/posts", tok, map[string]any{"type": "tip", "destination": "ภูเก็ต", "text": "ไปหาดในตอนเช้าคนน้อยและแดดไม่แรง", "groupId": 1})
	if code != 201 || p["type"] != "tip" || p["likeCount"] != 0.0 || p["authorId"] == "" || p["groupId"] != 1.0 {
		t.Fatal(code, p)
	}
	pid := itoa(p["id"].(float64))

	// appears first in the feed; destination filter works
	_, feed, _ := e.do("GET", "/v1/community/feed", tok, nil)
	if list(feed)[0].(map[string]any)["id"] != p["id"] || len(list(feed)) != 6 {
		t.Fatal(feed)
	}

	// comments
	if c, _, _ := e.do("POST", "/v1/community/posts/"+pid+"/comments", tok, map[string]any{"text": "  "}); c != 422 {
		t.Fatal(c)
	}
	if c, _, _ := e.do("POST", "/v1/community/posts/999/comments", tok, map[string]any{"text": "x"}); c != 404 {
		t.Fatal(c)
	}
	code, cm, _ := e.do("POST", "/v1/community/posts/"+pid+"/comments", other, map[string]any{"text": "ขอบคุณสำหรับทิป"})
	if code != 201 || cm["text"] != "ขอบคุณสำหรับทิป" || cm["authorId"] == "" {
		t.Fatal(code, cm)
	}
	code, d, _ := e.do("GET", "/v1/community/posts/"+pid, tok, nil)
	if code != 200 || len(d["comments"].([]any)) != 1 || d["commentCount"] != 1.0 {
		t.Fatal(code, d)
	}

	// like / save toggles reflect in detail
	_, lk, _ := e.do("POST", "/v1/community/posts/"+pid+"/like", tok, nil)
	_, sv, _ := e.do("POST", "/v1/community/posts/"+pid+"/save", tok, nil)
	if lk["liked"] != true || lk["likeCount"] != 1.0 || sv["saved"] != true {
		t.Fatal(lk, sv)
	}
	_, d, _ = e.do("GET", "/v1/community/posts/"+pid, tok, nil)
	if d["liked"] != true || d["saved"] != true {
		t.Fatal(d)
	}
	_, d, _ = e.do("GET", "/v1/community/posts/"+pid, other, nil)
	if d["liked"] != false || d["saved"] != false {
		t.Fatal("flags must be per viewer", d)
	}
	_, sv, _ = e.do("POST", "/v1/community/posts/"+pid+"/save", tok, nil)
	if sv["saved"] != false {
		t.Fatal(sv)
	}

	// hide is per user
	if c, _, _ := e.do("POST", "/v1/community/posts/"+pid+"/hide", other, nil); c != 204 {
		t.Fatal(c)
	}
	_, f1, _ := e.do("GET", "/v1/community/feed", other, nil)
	_, f2, _ := e.do("GET", "/v1/community/feed", tok, nil)
	if len(list(f1)) != 5 || len(list(f2)) != 6 {
		t.Fatal(len(list(f1)), len(list(f2)))
	}

	// report
	if c, _, _ := e.do("POST", "/v1/community/posts/"+pid+"/report", other, map[string]any{"reason": "bogus"}); c != 422 {
		t.Fatal(c)
	}
	if c, _, _ := e.do("POST", "/v1/community/posts/"+pid+"/report", other, map[string]any{"reason": "spam"}); c != 204 {
		t.Fatal(c)
	}
	if c, _, _ := e.do("POST", "/v1/community/posts/999/report", other, map[string]any{"reason": "spam"}); c != 404 {
		t.Fatal(c)
	}

	// delete: only the author
	if c, _, _ := e.do("DELETE", "/v1/community/posts/"+pid, other, nil); c != 403 {
		t.Fatalf("non-author delete = %d", c)
	}
	if c, _, _ := e.do("DELETE", "/v1/community/posts/"+pid, tok, nil); c != 204 {
		t.Fatal(c)
	}
	if c, _, _ := e.do("GET", "/v1/community/posts/"+pid, tok, nil); c != 404 {
		t.Fatal(c)
	}
	if c, _, _ := e.do("DELETE", "/v1/community/posts/"+pid, tok, nil); c != 404 {
		t.Fatal(c)
	}
}

func TestCommunityCopyTrip(t *testing.T) {
	e := newEnv(t, nil)
	tok := e.login()
	code, tr, _ := e.do("POST", "/v1/community/posts/4/copy-trip", tok, nil) // seeded "trip" post
	if code != 201 || tr["source"] != "community" || tr["destination"] != "เกาะสมุย" || tr["days"] != 3.0 || tr["budget"] != 12000.0 {
		t.Fatal(code, tr)
	}
	if _, l, _ := e.do("GET", "/v1/trips/saved", tok, nil); len(list(l)) != 1 {
		t.Fatal(l)
	}
	if c, _, _ := e.do("POST", "/v1/community/posts/1/copy-trip", tok, nil); c != 422 {
		t.Fatalf("post without a shared trip = %d", c)
	}
	if c, _, _ := e.do("POST", "/v1/community/posts/999/copy-trip", tok, nil); c != 404 {
		t.Fatal(c)
	}
}

func TestCommunityGroupsAndMembers(t *testing.T) {
	e := newEnv(t, nil)
	tok := e.login()
	_, g, _ := e.do("GET", "/v1/community/groups", tok, nil)
	if len(list(g)) != 3 {
		t.Fatal(g)
	}
	first := list(g)[0].(map[string]any)
	if first["memberCount"] != 3.0 || first["gradientCss"] == "" {
		t.Fatal(first)
	}
	code, d, _ := e.do("GET", "/v1/community/groups/"+itoa(first["id"].(float64)), tok, nil)
	if code != 200 || d["joined"] != true || len(d["rules"].([]any)) == 0 {
		t.Fatal(code, d)
	}
	_, j, _ := e.do("POST", "/v1/community/groups/"+itoa(first["id"].(float64))+"/join", tok, nil)
	_, d, _ = e.do("GET", "/v1/community/groups/"+itoa(first["id"].(float64)), tok, nil)
	if j["joined"] != false || d["joined"] != false || d["memberCount"] != 2.0 {
		t.Fatal(j, d)
	}
	if c, _, _ := e.do("GET", "/v1/community/groups/999", tok, nil); c != 404 {
		t.Fatal(c)
	}
	if c, _, _ := e.do("POST", "/v1/community/groups/999/join", tok, nil); c != 404 {
		t.Fatal(c)
	}

	code, m, _ := e.do("GET", "/v1/community/members/m_nok", tok, nil)
	if code != 200 || m["id"] != "m_nok" || m["followerCount"] != 1.0 || m["following"] != true || m["name"] == "" {
		t.Fatal(code, m)
	}
	_, f, _ := e.do("POST", "/v1/community/members/m_nok/follow", tok, nil)
	_, m, _ = e.do("GET", "/v1/community/members/m_nok", tok, nil)
	if f["following"] != false || m["followerCount"] != 0.0 {
		t.Fatal(f, m)
	}
	if c, _, _ := e.do("GET", "/v1/community/members/nobody", tok, nil); c != 404 {
		t.Fatal(c)
	}
	if c, _, _ := e.do("POST", "/v1/community/members/nobody/follow", tok, nil); c != 404 {
		t.Fatal(c)
	}
	_, me, _ := e.do("GET", "/v1/me", tok, nil)
	self := "m_" + itoa(me["id"].(float64))
	if c, _, _ := e.do("POST", "/v1/community/members/"+self+"/follow", tok, nil); c != 422 {
		t.Fatalf("self follow = %d", c)
	}
}
