package handler

import (
	"mime/multipart"
	"os"
	"strings"
	"testing"
)

type testMultipartFile struct {
	*strings.Reader
}

func (f testMultipartFile) Close() error {
	return nil
}

func TestValidateUploadImageAcceptsPNG(t *testing.T) {
	file := multipart.File(testMultipartFile{strings.NewReader("\x89PNG\r\n\x1a\n\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")})
	header := &multipart.FileHeader{Filename: "nav.png", Size: 20}

	if err := validateUploadImage(file, header); err != nil {
		t.Fatalf("validateUploadImage returned error: %v", err)
	}
}

func TestValidateUploadImageRejectsText(t *testing.T) {
	file := multipart.File(testMultipartFile{strings.NewReader("not an image")})
	header := &multipart.FileHeader{Filename: "nav.txt", Size: 12}

	if err := validateUploadImage(file, header); err == nil {
		t.Fatal("validateUploadImage returned nil error, want rejection")
	}
}

func TestUploadPublicBucketPrefersPublicBucket(t *testing.T) {
	t.Setenv("COS_PUBLIC_BUCKET", "public-bucket")
	t.Setenv("COS_ADMIN_BUCKET", "admin-bucket")

	if got := uploadPublicBucket(); got != "public-bucket" {
		t.Fatalf("bucket=%q, want public-bucket", got)
	}
}

func TestBuildUploadObjectKeyUsesPurposeAndDate(t *testing.T) {
	key := buildUploadObjectKey("product-category-nav", "nav.PNG")

	if !strings.HasPrefix(key, "product-category-nav/") {
		t.Fatalf("key=%q, want product-category-nav prefix", key)
	}
	if !strings.HasSuffix(key, ".png") {
		t.Fatalf("key=%q, want lower-case extension", key)
	}
}

func TestPathEscapeKeyKeepsSlashes(t *testing.T) {
	got := pathEscapeKey("product-category-nav/2026/05/09/a b.png")
	want := "product-category-nav/2026/05/09/a%20b.png"
	if got != want {
		t.Fatalf("pathEscapeKey=%q, want %q", got, want)
	}
}

func TestUploadPublicURLUsesConfiguredBaseURL(t *testing.T) {
	t.Setenv("COS_PUBLIC_BASE_URL", "https://cdn.example.com/")

	got := uploadPublicURL("bucket", "ap-beijing", "a b/nav.png")
	want := "https://cdn.example.com/a%20b/nav.png"
	if got != want {
		t.Fatalf("uploadPublicURL=%q, want %q", got, want)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
