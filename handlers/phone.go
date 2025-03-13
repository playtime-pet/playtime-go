package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"playtime-go/models"
	"playtime-go/services"
)

// HandlePhone handles requests to get user's phone number
func HandlePhone(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	
	// Parse request body
	var request models.PhoneRequest
	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}
	
	// Validate request
	if request.Code == "" {
		http.Error(w, "Code is required", http.StatusBadRequest)
		return
	}
	
	// Call service to get phone number
	phoneResponse, err := services.GetPhoneNumber(request.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(phoneResponse)
}
