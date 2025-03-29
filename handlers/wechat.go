package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"playtime-go/config"
	"playtime-go/models"
	"playtime-go/services"
	"playtime-go/utils"
)

func HandleWechat(w http.ResponseWriter, r *http.Request) {
	// Extract path for more specific handlers
	path := r.URL.Path
	path = path[len("/wechat/"):]

	// Route to the appropriate handler based on the path and method
	switch {
	case path == "phone":
		HandlePhone(w, r)
	case path == "auth":
		handleWechatAuth(w, r)
	case path == "login" && r.Method == http.MethodGet:
		HandleLogin(w, r)
	case path == "upload" && r.Method == http.MethodPost:
		HandleUpload(w, r)
	case path == "map/reverseGeocode" && r.Method == http.MethodGet:
		HandleReverseGeocode(w, r)
	default:
		utils.ErrorResponse(w, "Method not allowed or invalid URL", 405, http.StatusMethodNotAllowed)
	}
}

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

func handleWechatAuth(w http.ResponseWriter, r *http.Request) {
	// Extract path for more specific handlers
	cfg := config.GetConfig()
	if cfg.MiniMapKey == "" {
		utils.ErrorResponse(w, "MiniMap API key is not set", 500, http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, map[string]string{"key": cfg.MiniMapKey}, http.StatusOK)
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", 400, http.StatusBadRequest)
		return
	}

	query := r.URL.Query()
	code := query.Get("code")

	if code == "" {
		utils.ErrorResponse(w, "Code is required", 401, http.StatusBadRequest)
		return
	}

	session, err := services.GetLoginSession(code)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), 500, http.StatusInternalServerError)
		return
	}

	// Return response
	utils.SuccessResponse(w, session, http.StatusOK)
}

// HandleUpload handles file uploads to Tencent Cloud COS
func HandleUpload(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", 405, http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form with 10 MB max memory
	const maxMemory = 10 * 1024 * 1024 // 10 MB
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		utils.ErrorResponse(w, "Failed to parse form: "+err.Error(), 400, http.StatusBadRequest)
		return
	}

	// Get the file from form data
	file, header, err := r.FormFile("file")
	if err != nil {
		utils.ErrorResponse(w, "No file provided or invalid file field", 400, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check content type
	contentType := header.Header.Get("Content-Type")
	if !isAllowedImageType(contentType) {
		utils.ErrorResponse(w, "Unsupported file type: only images are allowed", 400, http.StatusBadRequest)
		return
	}

	log.Printf("Received file upload: %s, size: %d bytes, type: %s", header.Filename, header.Size, contentType)

	// Upload file to COS
	response, err := services.UploadFileToCOS(file, header.Filename, contentType)
	if err != nil {
		log.Printf("Failed to upload file to COS: %v", err)
		utils.ErrorResponse(w, "Failed to upload file: "+err.Error(), 500, http.StatusInternalServerError)
		return
	}

	// Return success response with file URL
	utils.SuccessResponse(w, response, http.StatusOK)
}

// isAllowedImageType checks if the content type is an allowed image type
func isAllowedImageType(contentType string) bool {
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	return allowedTypes[contentType]
}

func HandleReverseGeocode(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters
	query := r.URL.Query()
	lat := query.Get("lat")
	lng := query.Get("lng")

	// Validate latitude and longitude
	if lat == "" || lng == "" {
		utils.ErrorResponse(w, "Latitude and longitude are required", 400, http.StatusBadRequest)
		return
	}

	// Call service to reverse geocode
	location, err := services.ReverseGeocode(lat, lng)
	if err != nil {
		utils.ErrorResponse(w, "Failed to reverse geocode: "+err.Error(), 500, http.StatusInternalServerError)
		return
	}

	// Return response
	utils.SuccessResponse(w, location, http.StatusOK)
}
