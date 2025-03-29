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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const reviewCollection = "review"

func HandleReview(w http.ResponseWriter, r *http.Request) {

	urlParts := utils.ExtractUrlParam(r.URL.Path, "/review")
	var placeID, userID, reviewID string

	if len(urlParts) > 2 {
		if urlParts[0] == "place" {
			placeID = urlParts[1]
			reviewID = urlParts[2]
		} else if urlParts[0] == "user" {
			userID = urlParts[1]
			reviewID = urlParts[2]
		} else {
			reviewID = urlParts[1]
		}
	}

	switch {
	case userID != "":
		handleUserReview(userID, w, r)
	case placeID != "":
		handlePlaceReview(placeID, w, r)
	default:
		handleSingleReview(reviewID, w, r)
	}
}

func handleSingleReview(reviewID string, w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && reviewID == "":
		createReview(w, r)
	case r.Method == http.MethodGet && reviewID == "":
		listReviews(w, r)
	case r.Method == http.MethodGet && reviewID != "":
		getReview(w, r, reviewID)
	case r.Method == http.MethodPut && reviewID != "":
		updateReview(w, r, reviewID)
	case r.Method == http.MethodDelete && reviewID != "":
		deleteReview(w, r, reviewID)
	default:
		utils.ErrorResponse(w, "Method not allowed or invalid URL", 405, http.StatusMethodNotAllowed)
	}
}

func handleUserReview(userID string, w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet:
		getAllUserReview(w, r, userID)
	case r.Method == http.MethodDelete:
		deleteAllUserReview(w, r, userID)
	default:
		utils.ErrorResponse(w, "Method not allowed or invalid URL", 405, http.StatusMethodNotAllowed)
	}
}

func handlePlaceReview(placeID string, w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet:
		getAllPlaceReview(w, r, placeID)
	case r.Method == http.MethodDelete:
		deleteAllPlaceReview(w, r, placeID)
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

func getAllUserReview(w http.ResponseWriter, r *http.Request, userID string) {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.ErrorResponse(w, "Invalid user ID format", 400, http.StatusBadRequest)
		return
	}

	reviews, err := services.GetAllUserReview(id)
	if err != nil {
		if strings.Contains(err.Error(), "no reviews found") {
			utils.ErrorResponse(w, "No reviews found for this user", 404, http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to get reviews: "+err.Error(), 500, http.StatusInternalServerError)
		}
		return
	}

	utils.SuccessResponse(w, reviews, http.StatusOK)
}

func deleteAllUserReview(w http.ResponseWriter, r *http.Request, userID string) {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.ErrorResponse(w, "Invalid user ID format", 400, http.StatusBadRequest)
		return
	}

	err = services.DeleteAllUserReview(id)
	if err != nil {
		if strings.Contains(err.Error(), "no reviews found") {
			utils.ErrorResponse(w, "No reviews found for this user", 404, http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to delete reviews: "+err.Error(), 500, http.StatusInternalServerError)
		}
		return
	}

	utils.SuccessResponse(w, map[string]string{"message": "Reviews deleted successfully"}, http.StatusOK)
}

func getAllPlaceReview(w http.ResponseWriter, r *http.Request, placeID string) {
	id, err := primitive.ObjectIDFromHex(placeID)
	if err != nil {
		utils.ErrorResponse(w, "Invalid place ID format", 400, http.StatusBadRequest)
		return
	}

	reviews, err := services.GetAllPlaceReview(id)
	if err != nil {
		if strings.Contains(err.Error(), "no reviews found") {
			utils.ErrorResponse(w, "No reviews found for this place", 404, http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to get reviews: "+err.Error(), 500, http.StatusInternalServerError)
		}
		return
	}

	utils.SuccessResponse(w, reviews, http.StatusOK)
}

func deleteAllPlaceReview(w http.ResponseWriter, r *http.Request, placeID string) {
	id, err := primitive.ObjectIDFromHex(placeID)
	if err != nil {
		utils.ErrorResponse(w, "Invalid place ID format", 400, http.StatusBadRequest)
		return
	}

	err = services.DeleteAllPlaceReview(id)
	if err != nil {
		if strings.Contains(err.Error(), "no reviews found") {
			utils.ErrorResponse(w, "No reviews found for this place", 404, http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to delete reviews: "+err.Error(), 500, http.StatusInternalServerError)
		}
		return
	}

	utils.SuccessResponse(w, map[string]string{"message": "Reviews deleted successfully"}, http.StatusOK)
}

// listReviews handles GET requests to list reviews with optional filtering
func listReviews(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()

	// Get filter parameters
	placeIDParam := query.Get("placeId")
	userIDParam := query.Get("userId")
	ratingParam := query.Get("rating")
	limitParam := query.Get("limit")

	// Prepare filter
	filter := bson.M{}

	// Add placeId filter if provided
	if placeIDParam != "" {
		placeID, err := primitive.ObjectIDFromHex(placeIDParam)
		if err != nil {
			utils.ErrorResponse(w, "Invalid place ID format", 400, http.StatusBadRequest)
			return
		}
		filter["place_id"] = placeID
	}

	// Add userId filter if provided
	if userIDParam != "" {
		userID, err := primitive.ObjectIDFromHex(userIDParam)
		if err != nil {
			utils.ErrorResponse(w, "Invalid user ID format", 400, http.StatusBadRequest)
			return
		}
		filter["user_id"] = userID
	}

	// Add rating filter if provided
	if ratingParam != "" {
		rating, err := strconv.Atoi(ratingParam)
		if err != nil || rating < 1 || rating > 5 {
			utils.ErrorResponse(w, "Invalid rating parameter, must be between 1-5", 400, http.StatusBadRequest)
			return
		}
		filter["rating"] = rating
	}

	// Parse limit if provided
	var limit int64 = 100 // Default limit
	if limitParam != "" {
		parsedLimit, err := strconv.ParseInt(limitParam, 10, 64)
		if err != nil || parsedLimit <= 0 {
			utils.ErrorResponse(w, "Invalid limit parameter", 400, http.StatusBadRequest)
			return
		}
		limit = parsedLimit
	}

	// Create find options
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "date", Value: -1}}) // Sort by date, newest first
	findOptions.SetLimit(limit)

	// Get reviews from database
	var reviews []models.Review
	err := services.FindMany(reviewCollection, filter, &reviews, findOptions)
	if err != nil {
		utils.ErrorResponse(w, "Failed to list reviews: "+err.Error(), 500, http.StatusInternalServerError)
		return
	}

	if reviews == nil {
		reviews = make([]models.Review, 0)
	}
	// Return response
	utils.SuccessResponse(w, reviews, http.StatusOK)
}
