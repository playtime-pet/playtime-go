package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

// Response represents a standardized API response
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse sends a standardized error response as JSON
func ErrorResponse(w http.ResponseWriter, message string, code int, statusCode int) {
	response := Response{
		Code:    code,
		Message: message,
		Data:    nil,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// If we can't encode the error response, fall back to a simple error
		log.Printf("Failed to encode error response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// SuccessResponse sends a standardized success response as JSON
func SuccessResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	response := Response{
		Code:    0, // 0 means success
		Message: "Success",
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// If we can't encode the success response, return an error
		log.Printf("Failed to encode success response: %v", err)
		ErrorResponse(w, "Internal Server Error", 500, http.StatusInternalServerError)
	}
}
