package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Bas-travel/tep/services/api-go/internal/config"
	"github.com/Bas-travel/tep/services/api-go/internal/httpx"
)

type env struct {
	t   *testing.T
	s   *Server
	ts  *httptest.Server
	now time.Time
}

func newEnv(t *testing.T, checks map[string]Check) *env {
	cfg := config.Config{AppEnv: "dev", APIMode: "mock", JWTSecret: config.PlaceholderJWTSecret, MockPayment: true, Version: "test"}
	s := New(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), checks)
	e := &env{t: t, s: s, now: time.Now()}
	s.now = func() time.Time { return e.now }
	e.ts = httptest.NewServer(httpx.Chain(s.Handler(), slog.New(slog.NewTextHandler(io.Discard, nil)), false))
	t.Cleanup(e.ts.Close)
	return e
}

func (e *env) do(method, path, token string, body any, hdr ...string) (int, map[string]any, http.Header) {
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, e.ts.URL+path, rd)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	for i := 0; i+1 < len(hdr); i += 2 {
		req.Header.Set(hdr[i], hdr[i+1])
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		e.t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	return resp.StatusCode, m, resp.Header
}

func (e *env) login() string {
	code, m, _ := e.do("POST", "/v1/auth/login", "", map[string]string{"email": "demo@tep.dev", "password": "Passw0rd!"})
	if code != 200 {
		e.t.Fatalf("login %d %v", code, m)
	}
	return m["accessToken"].(string)
}

func TestHealth(t *testing.T) {
	e := newEnv(t, nil)
	code, m, _ := e.do("GET", "/health/live", "", nil)
	if code != 200 || m["status"] != "ok" {
		t.Fatal(code, m)
	}
	code, m, _ = e.do("GET", "/health/ready", "", nil)
	if code != 200 || m["checks"].(map[string]any)["mysql"] != "disabled" {
		t.Fatal(code, m)
	}
	e2 := newEnv(t, map[string]Check{"mysql": func(context.Context) error { return errors.New("down") }, "redis": func(context.Context) error { return nil }})
	code, m, _ = e2.do("GET", "/health/ready", "", nil)
	c := m["checks"].(map[string]any)
	if code != 503 || m["status"] != "degraded" || c["mysql"] != "down" || c["redis"] != "ok" {
		t.Fatal(code, m)
	}
}

func TestAuthFlowAndRefreshReuse(t *testing.T) {
	e := newEnv(t, nil)
	reg := map[string]any{"email": "new@example.com", "password": "longenough", "firstName": "A", "lastName": "B"}
	code, tok, _ := e.do("POST", "/v1/auth/register", "", reg)
	if code != 201 {
		t.Fatal(code, tok)
	}
	if c, _, _ := e.do("POST", "/v1/auth/register", "", reg); c != 409 {
		t.Fatalf("dup want 409 got %d", c)
	}
	code, m, _ := e.do("POST", "/v1/auth/register", "", map[string]any{"email": "bad", "password": "x"})
	if code != 422 || m["code"] != "VALIDATION_FAILED" || len(m["errors"].([]any)) < 3 {
		t.Fatal(code, m)
	}
	if c, _, _ := e.do("GET", "/v1/me", "", nil); c != 401 {
		t.Fatalf("me without token want 401 got %d", c)
	}
	code, me, _ := e.do("GET", "/v1/me", tok["accessToken"].(string), nil)
	if code != 200 || me["email"] != "new@example.com" {
		t.Fatal(code, me)
	}
	if _, has := me["password"]; has {
		t.Fatal("profile leaks password")
	}
	// rotation
	rt1 := tok["refreshToken"].(string)
	code, t2, _ := e.do("POST", "/v1/auth/refresh", "", map[string]string{"refreshToken": rt1})
	if code != 200 || t2["refreshToken"] == rt1 {
		t.Fatal(code, t2)
	}
	// reuse of rotated token -> 401 and revokes the new one too
	if c, _, _ := e.do("POST", "/v1/auth/refresh", "", map[string]string{"refreshToken": rt1}); c != 401 {
		t.Fatalf("reuse want 401 got %d", c)
	}
	if c, _, _ := e.do("POST", "/v1/auth/refresh", "", map[string]string{"refreshToken": t2["refreshToken"].(string)}); c != 401 {
		t.Fatalf("family should be revoked, got %d", c)
	}
	// expired access token
	e.now = e.now.Add(16 * time.Minute)
	if c, _, _ := e.do("GET", "/v1/me", tok["accessToken"].(string), nil); c != 401 {
		t.Fatalf("expired want 401 got %d", c)
	}
	e.now = time.Now()
	if c, _, _ := e.do("POST", "/v1/auth/login", "", map[string]string{"email": "demo@tep.dev", "password": "wrong"}); c != 401 {
		t.Fatalf("bad login want 401 got %d", c)
	}
}

func TestCatalog(t *testing.T) {
	e := newEnv(t, nil)
	_, m, _ := e.do("GET", "/v1/catalog/destinations", "", nil)
	if len(m["items"].([]any)) != 6 {
		t.Fatal(m)
	}
	_, m, _ = e.do("GET", "/v1/catalog/listings?type=hotel", "", nil)
	if n := len(m["items"].([]any)); n != 5 {
		t.Fatalf("hotels %d", n)
	}
	_, m, _ = e.do("GET", "/v1/catalog/listings?q=ภูเก็ต&limit=2", "", nil)
	if len(m["items"].([]any)) != 2 || m["nextCursor"] != "2" {
		t.Fatal(m)
	}
	_, m, _ = e.do("GET", "/v1/catalog/listings?q=ภูเก็ต&limit=2&cursor=2", "", nil)
	if len(m["items"].([]any)) == 0 {
		t.Fatal(m)
	}
	if c, _, _ := e.do("GET", "/v1/catalog/listings?limit=500", "", nil); c != 422 {
		t.Fatal(c)
	}
	if c, _, _ := e.do("GET", "/v1/catalog/listings/999", "", nil); c != 404 {
		t.Fatal(c)
	}
	code, m, _ := e.do("GET", "/v1/catalog/listings/1", "", nil)
	if code != 200 || m["name"] != "Andaman Sunset Resort" {
		t.Fatal(code, m)
	}
}

func TestBookingPaymentWalletLifecycle(t *testing.T) {
	e := newEnv(t, nil)
	tk := e.login()
	if c, _, _ := e.do("POST", "/v1/bookings", tk, map[string]any{"lines": []any{}}); c != 422 {
		t.Fatalf("empty lines want 422 got %d", c)
	}
	// listing 1: 3200 → fee 96 → total 3296, exceeds demo wallet (1500) => 402
	idem := []string{"Idempotency-Key", "k1"}
	code, b, _ := e.do("POST", "/v1/bookings", tk, map[string]any{"lines": []any{map[string]any{"listingId": 1, "quantity": 1}}}, idem...)
	if code != 201 || b["status"] != "pending_payment" || b["amount"].(float64) != 3296 {
		t.Fatal(code, b)
	}
	code, b2, h := e.do("POST", "/v1/bookings", tk, map[string]any{"lines": []any{map[string]any{"listingId": 1, "quantity": 1}}}, idem...)
	if code != 201 || b2["id"] != b["id"] || h.Get("Idempotent-Replay") != "true" {
		t.Fatal("idempotent replay failed", code, b2)
	}
	bid := b["id"].(float64)
	code, m, _ := e.do("POST", "/v1/payments/intents", tk, map[string]any{"bookingId": bid, "method": "wallet"})
	if code != 402 || m["code"] != "WALLET_INSUFFICIENT" {
		t.Fatal(code, m)
	}
	// PromptPay: pending, then succeeds after settle delay
	code, pi, _ := e.do("POST", "/v1/payments/intents", tk, map[string]any{"bookingId": bid, "method": "promptpay"})
	if code != 201 || pi["status"] != "pending" || pi["promptPayQrUrl"] == nil {
		t.Fatal(code, pi)
	}
	id := pi["id"].(string)
	_, m, _ = e.do("GET", "/v1/payments/intents/"+id, tk, nil)
	if m["status"] != "pending" {
		t.Fatal(m)
	}
	e.now = e.now.Add(4 * time.Second)
	_, m, _ = e.do("GET", "/v1/payments/intents/"+id, tk, nil)
	if m["status"] != "succeeded" {
		t.Fatal(m)
	}
	_, m, _ = e.do("GET", "/v1/bookings/"+itoa(bid), tk, nil)
	if m["status"] != "confirmed" || m["method"] != "PromptPay" {
		t.Fatal(m)
	}
	// cancel -> refund to wallet
	if c, _, _ := e.do("POST", "/v1/bookings/"+itoa(bid)+"/cancel", tk, map[string]string{"reason": ""}); c != 422 {
		t.Fatal(c)
	}
	code, m, _ = e.do("POST", "/v1/bookings/"+itoa(bid)+"/cancel", tk, map[string]string{"reason": "เปลี่ยนแผน"})
	if code != 200 || m["status"] != "cancelled" || m["refundAmount"].(float64) != 3296 {
		t.Fatal(code, m)
	}
	if c, _, _ := e.do("POST", "/v1/bookings/"+itoa(bid)+"/cancel", tk, map[string]string{"reason": "again"}); c != 422 {
		t.Fatalf("double cancel want 422 got %d", c)
	}
	_, w, _ := e.do("GET", "/v1/wallet", tk, nil)
	if w["balance"].(float64) != 1500+3296 {
		t.Fatal(w)
	}
	// wallet payment now succeeds
	_, b3, _ := e.do("POST", "/v1/bookings", tk, map[string]any{"lines": []any{map[string]any{"listingId": 6, "quantity": 2}}})
	code, pi, _ = e.do("POST", "/v1/payments/intents", tk, map[string]any{"bookingId": b3["id"], "method": "wallet"})
	if code != 201 || pi["status"] != "succeeded" {
		t.Fatal(code, pi)
	}
	_, w, _ = e.do("GET", "/v1/wallet", tk, nil)
	if w["balance"].(float64) != 1500+3296-247.2 {
		t.Fatal(w["balance"])
	}
	// bookings list + filter, ownership
	_, l, _ := e.do("GET", "/v1/bookings?status=confirmed", tk, nil)
	if len(l["items"].([]any)) != 1 {
		t.Fatal(l)
	}
	_, other, _ := e.do("POST", "/v1/auth/register", "", map[string]any{"email": "o@example.com", "password": "longenough", "firstName": "O", "lastName": "T"})
	if c, _, _ := e.do("GET", "/v1/bookings/"+itoa(bid), other["accessToken"].(string), nil); c != 404 {
		t.Fatalf("other user's booking want 404 got %d", c)
	}
}

func TestTopUp(t *testing.T) {
	e := newEnv(t, nil)
	tk := e.login()
	if c, _, _ := e.do("POST", "/v1/wallet/top-ups", tk, map[string]any{"amount": 0}); c != 422 {
		t.Fatal(c)
	}
	code, pi, _ := e.do("POST", "/v1/wallet/top-ups", tk, map[string]any{"amount": 500, "method": "promptpay"})
	if code != 202 {
		t.Fatal(code, pi)
	}
	e.now = e.now.Add(4 * time.Second)
	_, w, _ := e.do("GET", "/v1/wallet", tk, nil)
	if w["balance"].(float64) != 2000 {
		t.Fatal(w)
	}
}

func TestUnknownRouteIsProblemJSON(t *testing.T) {
	e := newEnv(t, nil)
	code, m, h := e.do("GET", "/v1/nope", "", nil)
	if code != 404 || m["code"] != "NOT_FOUND" || h.Get("Content-Type") == "" || h.Get("X-Request-Id") == "" {
		t.Fatal(code, m, h)
	}
}

func itoa(f float64) string { b, _ := json.Marshal(int64(f)); return string(b) }
