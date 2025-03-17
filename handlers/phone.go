package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"playtime-go/models"
	"playtime-go/services"
	"playtime-go/utils"
)

// HandlePhone handles requests to get user's phone number
func HandlePhone(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling phone request %s", r.Method)
	// Only accept POST requests
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", 405, http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ErrorResponse(w, "Failed to read request body", 400, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse request body
	var request models.PhoneRequest
	if err := json.Unmarshal(body, &request); err != nil {
		utils.ErrorResponse(w, "Invalid request format", 400, http.StatusBadRequest)
		return
	}

	// Validate request
	if request.Code == "" {
		utils.ErrorResponse(w, "Code is required", 400, http.StatusBadRequest)
		return
	}

	// Call service to get phone number
	phoneResponse, err := services.GetPhoneNumber(request.Code)
	if err != nil {
		log.Printf("Failed to get phone number: %v", err)
		utils.ErrorResponse(w, err.Error(), 500, http.StatusInternalServerError)
		return
	}

	// Return response
	utils.SuccessResponse(w, phoneResponse, http.StatusOK)
}
