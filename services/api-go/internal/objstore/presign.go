// Package objstore signs S3-compatible (AWS Signature V4, query-string) upload URLs using only the standard library.
package objstore

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

type Signer struct {
	Endpoint  string // e.g. http://localhost:9000 or https://s3.ap-southeast-1.amazonaws.com (path-style)
	Bucket    string
	Region    string
	AccessKey string
	SecretKey string
}

func hmacSHA256(key []byte, data string) []byte {
	m := hmac.New(sha256.New, key)
	m.Write([]byte(data))
	return m.Sum(nil)
}

// escape percent-encodes per SigV4 (unreserved characters untouched; "/" kept when keepSlash).
func escape(s string, keepSlash bool) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9', c == '-', c == '_', c == '.', c == '~':
			b.WriteByte(c)
		case c == '/' && keepSlash:
			b.WriteByte(c)
		default:
			fmt.Fprintf(&b, "%%%02X", c)
		}
	}
	return b.String()
}

// Presign returns a URL that allows `method` on bucket/key for `ttl`, without sending any credentials.
func (s Signer) Presign(method, key string, ttl time.Duration, now time.Time) (string, error) {
	u, err := url.Parse(s.Endpoint)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("invalid S3 endpoint %q", s.Endpoint)
	}
	now = now.UTC()
	date, stamp := now.Format("20060102"), now.Format("20060102T150405Z")
	scope := date + "/" + s.Region + "/s3/aws4_request"
	path := "/" + key
	if s.Bucket != "" { // empty bucket = virtual-hosted style (bucket is part of the endpoint host)
		path = "/" + s.Bucket + path
	}
	q := map[string]string{
		"X-Amz-Algorithm":     "AWS4-HMAC-SHA256",
		"X-Amz-Credential":    s.AccessKey + "/" + scope,
		"X-Amz-Date":          stamp,
		"X-Amz-Expires":       fmt.Sprint(int(ttl.Seconds())),
		"X-Amz-SignedHeaders": "host",
	}
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, escape(k, false)+"="+escape(q[k], false))
	}
	canonQuery := strings.Join(parts, "&")
	canonical := strings.Join([]string{method, escape(path, true), canonQuery, "host:" + u.Host + "\n", "host", "UNSIGNED-PAYLOAD"}, "\n")
	sum := sha256.Sum256([]byte(canonical))
	toSign := "AWS4-HMAC-SHA256\n" + stamp + "\n" + scope + "\n" + hex.EncodeToString(sum[:])
	k := hmacSHA256([]byte("AWS4"+s.SecretKey), date)
	k = hmacSHA256(k, s.Region)
	k = hmacSHA256(k, "s3")
	k = hmacSHA256(k, "aws4_request")
	sig := hex.EncodeToString(hmacSHA256(k, toSign))
	return u.Scheme + "://" + u.Host + escape(path, true) + "?" + canonQuery + "&X-Amz-Signature=" + sig, nil
}
