package auth

import (
	"testing"
	"time"
)

const secret = "0123456789abcdef0123456789abcdef"

func TestJWTRoundTripAndExpiry(t *testing.T) {
	now := time.Now()
	tok, _ := IssueJWT(secret, 7, "traveler", now)
	c, err := ParseJWT(secret, tok, now)
	if err != nil || c.Sub != 7 || c.Role != "traveler" {
		t.Fatalf("got %+v %v", c, err)
	}
	if _, err := ParseJWT(secret, tok, now.Add(AccessTTL+time.Second)); err == nil {
		t.Fatal("expected expiry")
	}
	if _, err := ParseJWT("other-secret-other-secret-other-xx", tok, now); err == nil {
		t.Fatal("expected bad signature")
	}
	if _, err := ParseJWT(secret, tok+"x", now); err == nil {
		t.Fatal("expected tamper detection")
	}
}

func TestPassword(t *testing.T) {
	h, err := HashPassword("Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword("Passw0rd!", h) || CheckPassword("wrong", h) {
		t.Fatal("password check failed")
	}
}
