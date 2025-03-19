package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"playtime-go/config"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

type LoginSession struct {
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
	SessionKey string `json:"session_key"`
	OpenID     string `json:"openid"`
	UnionID    string `json:"unionid"`
}

// get user login session info
func GetLoginSession(code string) (LoginSession, error) {

	cfg := config.GetConfig()

	// Prepare request URL
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", cfg.AppID, cfg.AppSecret, code)

	// Make POST request
	resp, err := http.Get(url)

	if err != nil {
		return LoginSession{}, fmt.Errorf("failed to fetch token: %v", err)
	}
	defer resp.Body.Close()

	var session LoginSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return LoginSession{}, fmt.Errorf("failed to parse token response: %v", err)
	}

	// Check if the response contains an error
	if session.ErrCode != 0 {
		return session, fmt.Errorf("WeChat API error: %d - %s", session.ErrCode, session.ErrMsg)
	}

	return session, nil
}

// UploadResponse represents the response for file upload
type UploadResponse struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
}

// UploadFileToCOS uploads a file to Tencent Cloud COS and returns the public URL
func UploadFileToCOS(fileReader io.Reader, originalFilename string, contentType string) (*UploadResponse, error) {
	// Get configuration from config
	cfg := config.GetConfig()
	
	if cfg.COSSecretID == "" || cfg.COSSecretKey == "" || cfg.COSBucketURL == "" {
		return nil, fmt.Errorf("missing COS configuration")
	}

	// Parse bucket URL
	u, err := url.Parse(cfg.COSBucketURL)
	if err != nil {
		return nil, fmt.Errorf("invalid COS bucket URL: %v", err)
	}

	// Initialize COS client
	b := &cos.BaseURL{BucketURL: u}
	cosClient := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.COSSecretID,
			SecretKey: cfg.COSSecretKey,
		},
	})

	// Generate unique filename
	fileExt := filepath.Ext(originalFilename)
	if fileExt == "" {
		// Default to .jpg if no extension provided
		if strings.HasPrefix(contentType, "image/jpeg") {
			fileExt = ".jpg"
		} else if strings.HasPrefix(contentType, "image/png") {
			fileExt = ".png"
		} else {
			fileExt = ".bin"
		}
	}
	
	fileName := fmt.Sprintf("avatar/%d%s", time.Now().UnixNano(), fileExt)

	// Set upload options
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	}

	// Upload the file
	_, err = cosClient.Object.Put(context.Background(), fileName, fileReader, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to COS: %v", err)
	}

	// Generate public URL
	publicURL := fmt.Sprintf("https://%s/%s", u.Host, fileName)

	return &UploadResponse{
		URL:      publicURL,
		Filename: fileName,
	}, nil
}
