package handlers

import (
	"log"
	"net/http"
	"playtime-go/services"
	"playtime-go/utils"
)

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
