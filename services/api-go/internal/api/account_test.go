package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Bas-travel/tep/services/api-go/internal/config"
)

func TestChangePassword(t *testing.T) {
	e := newEnv(t, nil)
	_, tk, _ := e.do("POST", "/v1/auth/register", "", map[string]any{"email": "pw@example.com", "password": "firstpass1", "firstName": "A", "lastName": "B"})
	at, rt := tk["accessToken"].(string), tk["refreshToken"].(string)
	if c, _, _ := e.do("PUT", "/v1/auth/password", "", map[string]any{"currentPassword": "firstpass1", "newPassword": "secondpass2"}); c != 401 {
		t.Fatal(c)
	}
	for name, body := range map[string]map[string]any{
		"wrong current": {"currentPassword": "nope", "newPassword": "secondpass2"},
		"short new":     {"currentPassword": "firstpass1", "newPassword": "short"},
		"same":          {"currentPassword": "firstpass1", "newPassword": "firstpass1"},
	} {
		if c, _, _ := e.do("PUT", "/v1/auth/password", at, body); c != 422 {
			t.Errorf("%s = %d", name, c)
		}
	}
	if c, _, _ := e.do("PUT", "/v1/auth/password", at, map[string]any{"currentPassword": "firstpass1", "newPassword": "secondpass2"}); c != 204 {
		t.Fatal(c)
	}
	if c, _, _ := e.do("POST", "/v1/auth/login", "", map[string]string{"email": "pw@example.com", "password": "firstpass1"}); c != 401 {
		t.Fatalf("old password still works: %d", c)
	}
	if c, _, _ := e.do("POST", "/v1/auth/login", "", map[string]string{"email": "pw@example.com", "password": "secondpass2"}); c != 200 {
		t.Fatalf("new password rejected: %d", c)
	}
	if c, _, _ := e.do("POST", "/v1/auth/refresh", "", map[string]string{"refreshToken": rt}); c != 401 {
		t.Fatalf("refresh token must be revoked after a password change, got %d", c)
	}
}

func TestDeleteAccount(t *testing.T) {
	e := newEnv(t, nil)
	tok := registerUser(e, "gone@example.com")
	_, p, _ := e.do("POST", "/v1/community/posts", tok, map[string]any{"type": "story", "text": "โพสต์ที่จะถูกลบพร้อมบัญชี"})
	if c, _, _ := e.do("DELETE", "/v1/me", "", nil); c != 401 {
		t.Fatal(c)
	}
	if c, _, _ := e.do("DELETE", "/v1/me", tok, nil); c != 204 {
		t.Fatal(c)
	}
	if c, _, _ := e.do("GET", "/v1/me", tok, nil); c != 401 {
		t.Fatalf("deleted user's token still valid: %d", c)
	}
	if c, _, _ := e.do("POST", "/v1/auth/login", "", map[string]string{"email": "gone@example.com", "password": "longenough"}); c != 401 {
		t.Fatalf("deleted user can log in: %d", c)
	}
	demo := e.login()
	if c, _, _ := e.do("GET", "/v1/community/posts/"+itoa(p["id"].(float64)), demo, nil); c != 404 {
		t.Fatalf("deleted user's post still visible: %d", c)
	}
	// the email can be registered again
	if c, _, _ := e.do("POST", "/v1/auth/register", "", map[string]any{"email": "gone@example.com", "password": "longenough", "firstName": "N", "lastName": "N"}); c != 201 {
		t.Fatalf("re-register = %d", c)
	}
}

func TestPushTokenAndPresign(t *testing.T) {
	e := newEnv(t, nil)
	tok := e.login()
	if c, _, _ := e.do("POST", "/v1/devices/push-tokens", "", map[string]any{"token": "t", "platform": "ios"}); c != 401 {
		t.Fatal(c)
	}
	for _, b := range []map[string]any{{"token": "", "platform": "ios"}, {"token": "abc", "platform": "windows"}} {
		if c, _, _ := e.do("POST", "/v1/devices/push-tokens", tok, b); c != 422 {
			t.Fatal(c, b)
		}
	}
	for i := 0; i < 2; i++ { // upsert: the same token twice is fine
		if c, _, _ := e.do("POST", "/v1/devices/push-tokens", tok, map[string]any{"token": "fcm-token-123", "platform": "android"}); c != 204 {
			t.Fatal(c)
		}
	}

	code, m, _ := e.do("POST", "/v1/media/presign", tok, map[string]any{"contentType": "image/png", "purpose": "post_image"})
	key, _ := m["objectKey"].(string)
	if code != 200 || !strings.HasPrefix(key, "post_image/") || !strings.HasSuffix(key, ".png") || m["uploadUrl"] == "" {
		t.Fatal(code, m)
	}
	for _, b := range []map[string]any{{"contentType": "text/html", "purpose": "avatar"}, {"contentType": "image/png", "purpose": "other"}} {
		if c, _, _ := e.do("POST", "/v1/media/presign", tok, b); c != 422 {
			t.Fatal(c, b)
		}
	}
	if c, _, _ := e.do("POST", "/v1/media/presign", tok, map[string]any{"contentType": "image/png", "purpose": "listing_image"}); c != 403 {
		t.Fatalf("travelers must not upload listing images: %d", c)
	}
	// a posted image key must come from presign
	if c, _, _ := e.do("POST", "/v1/community/posts", tok, map[string]any{"type": "story", "text": "มีรูปประกอบโพสต์นี้", "imageObjectKey": key}); c != 201 {
		t.Fatal(c)
	}
}

func TestPresignRealStorageSignsURL(t *testing.T) {
	e := newEnvCfg(t, nil, func(c *config.Config) {
		c.S3Endpoint, c.S3Bucket, c.S3Region, c.S3AccessKey, c.S3SecretKey = "http://localhost:9000", "tep-media", "us-east-1", "ak", "sk"
	})
	_, m, _ := e.do("POST", "/v1/media/presign", e.login(), map[string]any{"contentType": "image/jpeg", "purpose": "avatar"})
	u, _ := m["uploadUrl"].(string)
	if !strings.HasPrefix(u, "http://localhost:9000/tep-media/avatar/") || !strings.Contains(u, "X-Amz-Signature=") || strings.Contains(u, "sk") && strings.Contains(u, "SecretKey") {
		t.Fatal(u)
	}
}

func sign(secret string, body []byte) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write(body)
	return hex.EncodeToString(m.Sum(nil))
}

func (e *env) webhook(provider, secret string, ev map[string]any, badSig bool) int {
	raw, _ := json.Marshal(ev)
	sig := sign(secret, raw)
	if badSig {
		sig = sign("wrong-secret", raw)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/webhooks/payments/"+provider, strings.NewReader(string(raw)))
	req.Header.Set("X-Signature", sig)
	rec := httptest.NewRecorder()
	e.s.Handler().ServeHTTP(rec, req)
	return rec.Code
}

func TestPaymentWebhook(t *testing.T) {
	const secret = "webhook-test-secret-0123456789abcdef"
	e := newEnvCfg(t, nil, func(c *config.Config) { c.WebhookSecret, c.MockPayment = secret, false })
	tk := e.login()
	_, b, _ := e.do("POST", "/v1/bookings", tk, map[string]any{"lines": []any{map[string]any{"listingId": 6, "quantity": 1}}})
	_, pi, _ := e.do("POST", "/v1/payments/intents", tk, map[string]any{"bookingId": b["id"], "method": "promptpay"})
	id := pi["id"].(string)
	status := func() string {
		_, m, _ := e.do("GET", "/v1/bookings/"+itoa(b["id"].(float64)), tk, nil)
		return m["status"].(string)
	}

	// MOCK_PAYMENT=false: nothing settles by itself
	e.now = e.now.Add(time.Minute)
	if _, m, _ := e.do("GET", "/v1/payments/intents/"+id, tk, nil); m["status"] != "pending" || status() != "pending_payment" {
		t.Fatal("must stay pending without a webhook", m)
	}
	ev := map[string]any{"eventId": "evt_1", "intentId": id, "status": "succeeded"}
	if c := e.webhook("promptpay", secret, ev, true); c != 401 {
		t.Fatalf("bad signature = %d", c)
	}
	if c := e.webhook("unknown-psp", secret, ev, false); c != 404 {
		t.Fatalf("unknown provider = %d", c)
	}
	if c := e.webhook("promptpay", secret, map[string]any{"eventId": "evt_x", "intentId": "pi_nope", "status": "succeeded"}, false); c != 404 {
		t.Fatalf("unknown intent = %d", c)
	}
	if c := e.webhook("promptpay", secret, map[string]any{"eventId": "evt_y", "status": "weird"}, false); c != 422 {
		t.Fatalf("bad body = %d", c)
	}
	if status() != "pending_payment" {
		t.Fatal("rejected webhooks must not change state")
	}
	if c := e.webhook("promptpay", secret, ev, false); c != 200 {
		t.Fatalf("valid webhook = %d", c)
	}
	if status() != "confirmed" {
		t.Fatal("booking not confirmed by webhook")
	}
	// replay of the same event is acknowledged and changes nothing
	if c := e.webhook("promptpay", secret, ev, false); c != 200 {
		t.Fatal(c)
	}
	// top-up credited exactly once
	_, tp, _ := e.do("POST", "/v1/wallet/top-ups", tk, map[string]any{"amount": 250})
	tev := map[string]any{"eventId": "evt_2", "intentId": tp["id"], "status": "succeeded"}
	e.webhook("promptpay", secret, tev, false)
	e.webhook("promptpay", secret, tev, false)
	if _, w, _ := e.do("GET", "/v1/wallet", tk, nil); w["balance"] != 1750.0 {
		t.Fatalf("balance %v, want 1750 (credited once)", w["balance"])
	}
	// failed payment leaves the booking payable
	_, b2, _ := e.do("POST", "/v1/bookings", tk, map[string]any{"lines": []any{map[string]any{"listingId": 6}}})
	_, pi2, _ := e.do("POST", "/v1/payments/intents", tk, map[string]any{"bookingId": b2["id"], "method": "card"})
	if c := e.webhook("card", secret, map[string]any{"eventId": "evt_3", "intentId": pi2["id"], "status": "failed"}, false); c != 200 {
		t.Fatal(c)
	}
	if _, m, _ := e.do("GET", "/v1/payments/intents/"+pi2["id"].(string), tk, nil); m["status"] != "failed" {
		t.Fatal(m)
	}
	if _, m, _ := e.do("GET", "/v1/bookings/"+itoa(b2["id"].(float64)), tk, nil); m["status"] != "pending_payment" {
		t.Fatal(m)
	}
}

func TestWebhookDisabledWithoutSecret(t *testing.T) {
	e := newEnv(t, nil)
	if c := e.webhook("promptpay", "x", map[string]any{"eventId": "e", "intentId": "i", "status": "succeeded"}, false); c != 404 {
		t.Fatal(c)
	}
}
