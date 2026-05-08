package handler

import "testing"

func TestParseCOSObjectAllowsConfiguredPrivateBucketURL(t *testing.T) {
	t.Setenv("COS_PRIVATE_BUCKET", "steellink-user-pri-1321917249")

	bucket, key, err := parseCOSObject("https://steellink-user-pri-1321917249.cos.ap-beijing.myqcloud.com/identity/cards/2026/05/08/file.jpeg", "ap-beijing")
	if err != nil {
		t.Fatalf("parseCOSObject returned error: %v", err)
	}
	if bucket != "steellink-user-pri-1321917249" {
		t.Fatalf("bucket=%q, want private bucket", bucket)
	}
	if key != "identity/cards/2026/05/08/file.jpeg" {
		t.Fatalf("key=%q, want identity card key", key)
	}
	if !allowedPrivateBucket(bucket) {
		t.Fatal("bucket should be allowed")
	}
}

func TestParseCOSObjectRejectsNonHTTPSURL(t *testing.T) {
	_, _, err := parseCOSObject("http://steellink-user-pri-1321917249.cos.ap-beijing.myqcloud.com/identity/cards/file.jpeg", "ap-beijing")
	if err == nil {
		t.Fatal("parseCOSObject returned nil error, want rejection")
	}
}

func TestPrivateFileURLExpiresInCapsLongValues(t *testing.T) {
	t.Setenv("COS_PRIVATE_URL_EXPIRES_SECONDS", "7200")

	if got := privateFileURLExpiresIn(); got != 3600 {
		t.Fatalf("expires=%d, want capped 3600", got)
	}
}
