package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"playtime-go/config"
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
