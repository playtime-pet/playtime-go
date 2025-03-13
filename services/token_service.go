package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"playtime-go/config"
	"playtime-go/models"
	"sync"
	"time"
)

var (
	token      models.Token
	tokenMutex sync.RWMutex
	tokenTime  time.Time
)

// GetToken returns the cached token or fetches a new one if expired
func GetToken() (models.Token, error) {
	tokenMutex.RLock()
	// If token exists and is not expired (with 5 min buffer), return it
	if token.AccessToken != "" && time.Since(tokenTime).Seconds() < float64(token.ExpiresIn-300) {
		defer tokenMutex.RUnlock()
		return token, nil
	}
	tokenMutex.RUnlock()

	return FetchNewToken()
}

// FetchNewToken gets a new access token from WeChat API
func FetchNewToken() (models.Token, error) {
	cfg := config.GetConfig()
	
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		cfg.AppID, cfg.AppSecret)
	
	resp, err := http.Get(url)
	if err != nil {
		return models.Token{}, fmt.Errorf("failed to fetch token: %v", err)
	}
	defer resp.Body.Close()

	var newToken models.Token
	if err := json.NewDecoder(resp.Body).Decode(&newToken); err != nil {
		return models.Token{}, fmt.Errorf("failed to parse token response: %v", err)
	}

	// Check if the response contains an error
	if newToken.ErrCode != 0 {
		return newToken, fmt.Errorf("WeChat API error: %d - %s", newToken.ErrCode, newToken.ErrMsg)
	}

	// Update the cached token with a mutex lock
	tokenMutex.Lock()
	token = newToken
	tokenTime = time.Now()
	tokenMutex.Unlock()

	return newToken, nil
}