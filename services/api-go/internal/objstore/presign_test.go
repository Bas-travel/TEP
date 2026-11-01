package objstore

import (
	"strings"
	"testing"
	"time"
)

// Vector from the AWS documentation "Authenticating Requests: Using Query Parameters (AWS Signature Version 4)":
// GET https://examplebucket.s3.amazonaws.com/test.txt, 2013-05-24T00:00:00Z, expires 86400, region us-east-1.
// The documented host-style URL is reproduced here with a path-style bucket on the same host by signing for
// host "examplebucket.s3.amazonaws.com" and an empty bucket segment, so we verify the signing algorithm itself.
func TestPresignMatchesAWSDocumentedSignature(t *testing.T) {
	s := Signer{Endpoint: "https://examplebucket.s3.amazonaws.com", Bucket: "", Region: "us-east-1",
		AccessKey: "AKIAIOSFODNN7EXAMPLE", SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"}
	got, err := s.Presign("GET", "test.txt", 86400*time.Second, time.Date(2013, 5, 24, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	const want = "aeeed9bbccd4d02ee5c0109b86d86835f995330da4c265957d157751f604d404"
	if !strings.HasSuffix(got, "X-Amz-Signature="+want) {
		t.Fatalf("signature mismatch:\n got %s\nwant …%s", got, want)
	}
}

func TestPresignPathStyleShape(t *testing.T) {
	s := Signer{Endpoint: "http://localhost:9000", Bucket: "tep-media", Region: "us-east-1", AccessKey: "k", SecretKey: "s"}
	u, err := s.Presign("PUT", "post_image/1/a b.jpg", 10*time.Minute, time.Unix(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(u, "http://localhost:9000/tep-media/post_image/1/a%20b.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256") || !strings.Contains(u, "X-Amz-Expires=600") {
		t.Fatal(u)
	}
}
