package models

// PhoneRequest represents the request body for getting phone number
type PhoneRequest struct {
	Code string `json:"code"`
}

// PhoneResponse represents the response from WeChat API
type PhoneResponse struct {
	ErrCode   int       `json:"errcode"`
	ErrMsg    string    `json:"errmsg"`
	PhoneInfo PhoneInfo `json:"phone_info,omitempty"`
}

// PhoneInfo contains the user's phone information
type PhoneInfo struct {
	PhoneNumber     string    `json:"phoneNumber"`
	PurePhoneNumber string    `json:"purePhoneNumber"`
	CountryCode     string    `json:"countryCode"`
	Watermark       Watermark `json:"watermark"`
}

// Watermark contains metadata about the phone info
type Watermark struct {
	Timestamp int64  `json:"timestamp"`
	AppID     string `json:"appid"`
}
