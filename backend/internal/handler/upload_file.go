package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tencentyun/cos-go-sdk-v5"
)

const maxUploadImageBytes = 5 << 20

type uploadFileResponse struct {
	URL string `json:"url"`
	Key string `json:"key"`
}

func UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "file is required"})
		return
	}
	defer file.Close()

	purpose := normalizeUploadPurpose(c.PostForm("purpose"))
	if err := validateUploadImage(file, header); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	region := envWithDefault("COS_REGION", defaultCOSRegion)
	bucket := uploadPublicBucket()
	if bucket == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "cos public bucket is not configured"})
		return
	}

	secretID := strings.TrimSpace(os.Getenv("COS_SECRET_ID"))
	secretKey := strings.TrimSpace(os.Getenv("COS_SECRET_KEY"))
	if secretID == "" || secretKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "cos credentials are not configured"})
		return
	}

	key := buildUploadObjectKey(purpose, header.Filename)
	bucketURL, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket, region))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "invalid cos bucket config"})
		return
	}
	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
			Transport: http.DefaultTransport,
		},
	})

	if _, err := client.Object.Put(c.Request.Context(), key, file, nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "upload file failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": uploadFileResponse{
			URL: uploadPublicURL(bucket, region, key),
			Key: key,
		},
	})
}

func validateUploadImage(file multipart.File, header *multipart.FileHeader) error {
	if header == nil || header.Size <= 0 {
		return fmt.Errorf("empty image is not allowed")
	}
	if header.Size > maxUploadImageBytes {
		return fmt.Errorf("image size must be <= 5MB")
	}

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("read image failed")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("read image failed")
	}
	contentType := http.DetectContentType(buffer[:n])
	switch contentType {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
		return nil
	default:
		return fmt.Errorf("only jpg, png, webp or gif image is allowed")
	}
}

func normalizeUploadPurpose(raw string) string {
	switch strings.TrimSpace(raw) {
	case "product_category_nav":
		return "product-category-nav"
	default:
		return "admin-images"
	}
}

func uploadPublicBucket() string {
	for _, key := range []string{"COS_PUBLIC_BUCKET", "COS_ADMIN_PUBLIC_BUCKET", "COS_ADMIN_BUCKET"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func uploadPublicURL(bucket string, region string, key string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("COS_PUBLIC_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket, region)
	}
	return fmt.Sprintf("%s/%s", baseURL, pathEscapeKey(key))
}

func buildUploadObjectKey(purpose string, filename string) string {
	now := time.Now()
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" || len(ext) > 10 {
		ext = ".jpg"
	}
	return fmt.Sprintf("%s/%04d/%02d/%02d/%d-%s%s", purpose, now.Year(), now.Month(), now.Day(), now.UnixNano(), randomHex(4), ext)
}

func randomHex(bytesLen int) string {
	buf := make([]byte, bytesLen)
	if _, err := rand.Read(buf); err != nil {
		return "00000000"
	}
	return hex.EncodeToString(buf)
}

func pathEscapeKey(key string) string {
	parts := strings.Split(key, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}
