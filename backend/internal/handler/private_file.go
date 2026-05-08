package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tencentyun/cos-go-sdk-v5"
)

const defaultCOSRegion = "ap-beijing"

type privateFileSignResponse struct {
	URL       string `json:"url"`
	ExpiresIn int    `json:"expiresIn"`
}

func GetPrivateFileURL(c *gin.Context) {
	rawURL := strings.TrimSpace(c.Query("url"))
	if rawURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "url is required"})
		return
	}

	region := envWithDefault("COS_REGION", defaultCOSRegion)
	bucket, key, err := parseCOSObject(rawURL, region)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	if !allowedPrivateBucket(bucket) {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "bucket is not allowed"})
		return
	}

	secretID := strings.TrimSpace(os.Getenv("COS_SECRET_ID"))
	secretKey := strings.TrimSpace(os.Getenv("COS_SECRET_KEY"))
	if secretID == "" || secretKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "cos credentials are not configured"})
		return
	}

	expiresIn := privateFileURLExpiresIn()
	bucketURL, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket, region))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "invalid cos bucket config"})
		return
	}
	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, nil)
	signedURL, err := client.Object.GetPresignedURL(c.Request.Context(), http.MethodGet, key, secretID, secretKey, time.Duration(expiresIn)*time.Second, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "sign private file url failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": privateFileSignResponse{
			URL:       signedURL.String(),
			ExpiresIn: expiresIn,
		},
	})
}

func parseCOSObject(rawURL string, region string) (bucket string, key string, err error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid url")
	}

	if parsed.Scheme == "" && parsed.Host == "" {
		key = strings.TrimLeft(parsed.Path, "/")
		if key == "" || strings.Contains(key, "..") {
			return "", "", fmt.Errorf("invalid object key")
		}
		return strings.TrimSpace(os.Getenv("COS_PRIVATE_BUCKET")), key, nil
	}

	if parsed.Scheme != "https" {
		return "", "", fmt.Errorf("only https cos url is allowed")
	}
	host := strings.ToLower(parsed.Hostname())
	suffix := fmt.Sprintf(".cos.%s.myqcloud.com", region)
	if !strings.HasSuffix(host, suffix) {
		return "", "", fmt.Errorf("only configured cos region url is allowed")
	}
	bucket = strings.TrimSuffix(host, suffix)
	key = strings.TrimLeft(parsed.EscapedPath(), "/")
	if bucket == "" || key == "" || strings.Contains(key, "..") {
		return "", "", fmt.Errorf("invalid cos object url")
	}
	if decodedKey, decodeErr := url.PathUnescape(key); decodeErr == nil {
		key = decodedKey
	}
	return bucket, key, nil
}

func allowedPrivateBucket(bucket string) bool {
	allowed := []string{
		strings.TrimSpace(os.Getenv("COS_PRIVATE_BUCKET")),
		strings.TrimSpace(os.Getenv("COS_ADMIN_BUCKET")),
	}
	for _, item := range allowed {
		if item != "" && bucket == item {
			return true
		}
	}
	return false
}

func privateFileURLExpiresIn() int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv("COS_PRIVATE_URL_EXPIRES_SECONDS")))
	if err != nil || value <= 0 {
		return 600
	}
	if value > 3600 {
		return 3600
	}
	return value
}

func envWithDefault(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
