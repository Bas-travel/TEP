// Package auth provides HS256 JWTs and PBKDF2 password hashing using only the standard library.
package auth

import (
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const AccessTTL = 15 * time.Minute

type Claims struct {
	Sub  int64  `json:"sub"`
	Role string `json:"role"`
	Exp  int64  `json:"exp"`
	Iat  int64  `json:"iat"`
}

var b64 = base64.RawURLEncoding

func sign(secret, data string) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(data))
	return b64.EncodeToString(m.Sum(nil))
}

func IssueJWT(secret string, userID int64, role string, now time.Time) (string, error) {
	hdr := b64.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	body, err := json.Marshal(Claims{Sub: userID, Role: role, Iat: now.Unix(), Exp: now.Add(AccessTTL).Unix()})
	if err != nil {
		return "", err
	}
	data := hdr + "." + b64.EncodeToString(body)
	return data + "." + sign(secret, data), nil
}

var ErrInvalidToken = errors.New("invalid token")

func ParseJWT(secret, tok string, now time.Time) (Claims, error) {
	var c Claims
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		return c, ErrInvalidToken
	}
	hdr, err := b64.DecodeString(parts[0])
	if err != nil || !strings.Contains(string(hdr), `"HS256"`) {
		return c, ErrInvalidToken
	}
	want := sign(secret, parts[0]+"."+parts[1])
	if subtle.ConstantTimeCompare([]byte(want), []byte(parts[2])) != 1 {
		return c, ErrInvalidToken
	}
	raw, err := b64.DecodeString(parts[1])
	if err != nil || json.Unmarshal(raw, &c) != nil {
		return c, ErrInvalidToken
	}
	if now.Unix() >= c.Exp {
		return c, fmt.Errorf("%w: expired", ErrInvalidToken)
	}
	return c, nil
}

const pbkdf2Iter = 120_000

func HashPassword(pw string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	k, err := pbkdf2.Key(sha256.New, pw, salt, pbkdf2Iter, 32)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("pbkdf2$%d$%s$%s", pbkdf2Iter, hex.EncodeToString(salt), hex.EncodeToString(k)), nil
}

func CheckPassword(pw, enc string) bool {
	p := strings.Split(enc, "$")
	if len(p) != 4 || p[0] != "pbkdf2" {
		return false
	}
	iter, err := strconv.Atoi(p[1])
	if err != nil {
		return false
	}
	salt, err1 := hex.DecodeString(p[2])
	want, err2 := hex.DecodeString(p[3])
	if err1 != nil || err2 != nil {
		return false
	}
	got, err := pbkdf2.Key(sha256.New, pw, salt, iter, len(want))
	return err == nil && subtle.ConstantTimeCompare(got, want) == 1
}

// NewOpaqueToken returns a 256-bit random URL-safe token (refresh tokens).
func NewOpaqueToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return b64.EncodeToString(b)
}
