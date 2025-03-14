package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"playtime-go/models"
)

// GetPhoneNumber sends a request to WeChat API to get user's phone number
func GetPhoneNumber(code string) (models.PhoneResponse, error) {
	// Get access token first
	token, err := GetToken()
	if err != nil {
		return models.PhoneResponse{}, fmt.Errorf("failed to get access token: %v", err)
	}
	
	// Prepare request URL
	url := fmt.Sprintf("https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s", token.AccessToken)
	
	// Prepare request body
	requestBody := models.PhoneRequest{
		Code: code,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return models.PhoneResponse{}, fmt.Errorf("failed to marshal request body: %v", err)
	}
	
	// Make POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if resp.StatusCode != http.StatusOK {
		return models.PhoneResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	if err != nil {
		return models.PhoneResponse{}, fmt.Errorf("failed to send request to %s with body %s: %v", url, string(jsonBody), err)
	}
	
	defer resp.Body.Close()
	
	// Parse response
	var phoneResponse models.PhoneResponse
	if err := json.NewDecoder(resp.Body).Decode(&phoneResponse); err != nil {
		return models.PhoneResponse{}, fmt.Errorf("failed to parse response: %v", err)
	}
	
	// Check for API errors
	if phoneResponse.ErrCode != 0 {
		return phoneResponse, fmt.Errorf("WeChat API error: %d - %s", phoneResponse.ErrCode, phoneResponse.ErrMsg)
	}
	
	return phoneResponse, nil
}
