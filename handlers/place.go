package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"playtime-go/models"
	"playtime-go/services"
	"playtime-go/utils"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// HandlePlaceReviews handles review CRUD operations
func HandlePlaceReviews(w http.ResponseWriter, r *http.Request) {
	urlParts := strings.Split(r.URL.Path, "/")
	var reviewID, placeID string

	// Extract place ID or review ID from URL
	if len(urlParts) > 3 && urlParts[1] == "place" {
		if urlParts[2] == "review" && len(urlParts) > 3 {
			reviewID = urlParts[3] // For /place/review/{id} routes
		} else if urlParts[3] == "reviews" {
			placeID = urlParts[2] // For /place/{id}/reviews route
		}
	}

	// Route to appropriate handler
	switch {
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/place/review"):
		createReview(w, r)
	case r.Method == http.MethodGet && reviewID != "":
		getReview(w, r, reviewID)
	case r.Method == http.MethodPut && reviewID != "":
		updateReview(w, r, reviewID)
	case r.Method == http.MethodDelete && reviewID != "":
		deleteReview(w, r, reviewID)
	case r.Method == http.MethodGet && placeID != "":
		getPlaceReviews(w, r, placeID)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/user/") && strings.HasSuffix(r.URL.Path, "/reviews"):
		userID := urlParts[2] // Extract from /user/{id}/reviews
		getUserReviews(w, r, userID)
	default:
		utils.ErrorResponse(w, "Method not allowed or invalid URL", 405, http.StatusMethodNotAllowed)
	}
}

// createReview handles POST /place/review
func createReview(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ErrorResponse(w, "Failed to read request body", 400, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var request models.Review
	if err := json.Unmarshal(body, &request); err != nil {
		utils.ErrorResponse(w, "Invalid request format", 400, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if request.PlaceID == "" {
		utils.ErrorResponse(w, "Place ID is required", 400, http.StatusBadRequest)
		return
	}
	if request.UserID == "" {
		utils.ErrorResponse(w, "User ID is required", 400, http.StatusBadRequest)
		return
	}
	if request.Content == "" {
		utils.ErrorResponse(w, "Content is required", 400, http.StatusBadRequest)
		return
	}
	if request.Rating < 1 || request.Rating > 5 {
		utils.ErrorResponse(w, "Rating must be between 1 and 5", 400, http.StatusBadRequest)
		return
	}

	review, err := services.CreateReview(request)
	if err != nil {
		utils.ErrorResponse(w, "Failed to create review: "+err.Error(), 500, http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, review, http.StatusCreated)
}

// getReview handles GET /place/review/{id}
func getReview(w http.ResponseWriter, r *http.Request, reviewID string) {
	id, err := primitive.ObjectIDFromHex(reviewID)
	if err != nil {
		utils.ErrorResponse(w, "Invalid review ID format", 400, http.StatusBadRequest)
		return
	}

	review, err := services.GetReview(id)
	if err != nil {
		if strings.Contains(err.Error(), "no review found") {
			utils.ErrorResponse(w, "Review not found", 404, http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to get review: "+err.Error(), 500, http.StatusInternalServerError)
		}
		return
	}

	utils.SuccessResponse(w, review, http.StatusOK)
}

// updateReview handles PUT /place/review/{id}
func updateReview(w http.ResponseWriter, r *http.Request, reviewID string) {
	id, err := primitive.ObjectIDFromHex(reviewID)
	if err != nil {
		utils.ErrorResponse(w, "Invalid review ID format", 400, http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ErrorResponse(w, "Failed to read request body", 400, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var request models.Review
	if err := json.Unmarshal(body, &request); err != nil {
		utils.ErrorResponse(w, "Invalid request format", 400, http.StatusBadRequest)
		return
	}

	// Validate content and rating if provided
	if request.Content == "" {
		utils.ErrorResponse(w, "Content is required", 400, http.StatusBadRequest)
		return
	}
	if request.Rating < 1 || request.Rating > 5 {
		utils.ErrorResponse(w, "Rating must be between 1 and 5", 400, http.StatusBadRequest)
		return
	}

	review, err := services.UpdateReview(id, request)
	if err != nil {
		if strings.Contains(err.Error(), "no review found") {
			utils.ErrorResponse(w, "Review not found", 404, http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to update review: "+err.Error(), 500, http.StatusInternalServerError)
		}
		return
	}

	utils.SuccessResponse(w, review, http.StatusOK)
}

// deleteReview handles DELETE /place/review/{id}
func deleteReview(w http.ResponseWriter, r *http.Request, reviewID string) {
	id, err := primitive.ObjectIDFromHex(reviewID)
	if err != nil {
		utils.ErrorResponse(w, "Invalid review ID format", 400, http.StatusBadRequest)
		return
	}

	err = services.DeleteReview(id)
	if err != nil {
		if strings.Contains(err.Error(), "no review found") {
			utils.ErrorResponse(w, "Review not found", 404, http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to delete review: "+err.Error(), 500, http.StatusInternalServerError)
		}
		return
	}

	utils.SuccessResponse(w, map[string]string{"message": "Review deleted successfully"}, http.StatusOK)
}

// getPlaceReviews handles GET /place/{id}/reviews
func getPlaceReviews(w http.ResponseWriter, r *http.Request, placeID string) {
	// Get optional limit query parameter
	limit := int64(0) // 0 means no limit (service will use default)
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.ParseInt(limitStr, 10, 64)
		if err == nil {
			limit = parsedLimit
		}
	}

	reviews, err := services.GetReviewsByPlace(placeID, limit)
	if err != nil {
		utils.ErrorResponse(w, "Failed to get reviews: "+err.Error(), 500, http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, reviews, http.StatusOK)
}

// getUserReviews handles GET /user/{id}/reviews
func getUserReviews(w http.ResponseWriter, r *http.Request, userID string) {
	reviews, err := services.GetReviewsByUserID(r.Context(), userID)
	if err != nil {
		utils.ErrorResponse(w, "Failed to get user reviews: "+err.Error(), 500, http.StatusInternalServerError)
		return
	}

	utils.SuccessResponse(w, reviews, http.StatusOK)
}
