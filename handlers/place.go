package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"playtime-go/models"
	"playtime-go/services"
	"playtime-go/utils"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// HandleMap handles location operations
func HandlePlace(w http.ResponseWriter, r *http.Request) {
	// Extract path for more specific handlers
	urlParts := utils.ExtractUrlParam(r.URL.Path, "/place")
	var placeID string
	if len(urlParts) > 0 {
		// might be search
		placeID = urlParts[0]
	}

	// Route to the appropriate handler based on the path and method
	switch {
	case placeID == "" && r.Method == http.MethodPost:
		createPlace(w, r)
	case placeID == "" && r.Method == http.MethodGet:
		listPlaces(w, r)
	case placeID == "search" && r.Method == http.MethodGet:
		searchPlaces(w, r)
	case placeID != "" && r.Method == http.MethodGet:
		getPlace(placeID, w, r)
	case placeID != "" && r.Method == http.MethodPut:
		updatePlace(placeID, w, r)
	case placeID != "" && r.Method == http.MethodDelete:
		deletePlace(placeID, w, r)
	default:
		utils.ErrorResponse(w, "Method not allowed or invalid URL", 405, http.StatusMethodNotAllowed)
	}
}

// createPlace handles POST requests to create a new location
func createPlace(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ErrorResponse(w, "Failed to read request body", 400, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse request body
	var request models.LocationRequest
	if err := json.Unmarshal(body, &request); err != nil {
		utils.ErrorResponse(w, "Invalid request format", 400, http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateLocationRequest(request); err != nil {
		utils.ErrorResponse(w, err.Error(), 400, http.StatusBadRequest)
		return
	}

	// Call service to create location
	location, err := services.CreateLocation(request)
	if err != nil {
		utils.ErrorResponse(w, "Failed to create location: "+err.Error(), 500, http.StatusInternalServerError)
		return
	}

	// Return response
	utils.SuccessResponse(w, location, http.StatusCreated)
}

// getPlace handles GET requests to retrieve a specific location
func getPlace(id string, w http.ResponseWriter, r *http.Request) {

	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(w, "Invalid location ID format", 400, http.StatusBadRequest)
		return
	}

	// Get the location
	location, err := services.GetLocationByID(objectID)
	if err != nil {
		if strings.Contains(err.Error(), "no location found") {
			utils.ErrorResponse(w, "Location not found", 404, http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to get location: "+err.Error(), 500, http.StatusInternalServerError)
		}
		return
	}

	// Return response
	utils.SuccessResponse(w, location, http.StatusOK)
}

// listPlaces handles GET requests to list all locations
func listPlaces(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	category := query.Get("category")
	limitStr := query.Get("limit")

	// Parse limit if provided
	var limit int64 = 100 // default limit
	if limitStr != "" {
		parsedLimit, err := strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			utils.ErrorResponse(w, "Invalid limit parameter", 400, http.StatusBadRequest)
			return
		}
		limit = parsedLimit
	}

	// Get locations
	locations, err := services.ListLocations(category, limit)
	if err != nil {
		utils.ErrorResponse(w, "Failed to list locations: "+err.Error(), 500, http.StatusInternalServerError)
		return
	}

	// Return response
	utils.SuccessResponse(w, locations, http.StatusOK)
}

// updatePlace handles PUT requests to update a location
func updatePlace(id string, w http.ResponseWriter, r *http.Request) {

	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(w, "Invalid location ID format", 400, http.StatusBadRequest)
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
	var request models.LocationRequest
	if err := json.Unmarshal(body, &request); err != nil {
		utils.ErrorResponse(w, "Invalid request format", 400, http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateLocationRequest(request); err != nil {
		utils.ErrorResponse(w, err.Error(), 400, http.StatusBadRequest)
		return
	}

	// Call service to update location
	location, err := services.UpdateLocation(objectID, request)
	if err != nil {
		if strings.Contains(err.Error(), "no location found") {
			utils.ErrorResponse(w, "Location not found", 404, http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to update location: "+err.Error(), 500, http.StatusInternalServerError)
		}
		return
	}

	// Return response
	utils.SuccessResponse(w, location, http.StatusOK)
}

// deletePlace handles DELETE requests to remove a location
func deletePlace(id string, w http.ResponseWriter, r *http.Request) {

	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.ErrorResponse(w, "Invalid location ID format", 400, http.StatusBadRequest)
		return
	}

	// Call service to delete location
	err = services.DeleteLocation(objectID)
	if err != nil {
		if strings.Contains(err.Error(), "no location found") {
			utils.ErrorResponse(w, "Location not found", 404, http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to delete location: "+err.Error(), 500, http.StatusInternalServerError)
		}
		return
	}

	// Return success response
	utils.SuccessResponse(w, map[string]string{"message": "Location deleted successfully"}, http.StatusOK)
}

// searchPlaces handles GET requests to search for nearby locations
func searchPlaces(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()

	// Extract required parameters
	latStr := query.Get("latitude")
	lngStr := query.Get("longitude")
	keyword := query.Get("keyword")

	// Extract optional parameters
	radiusStr := query.Get("radius")
	limitStr := query.Get("limit")
	category := query.Get("category")

	// Validate and parse latitude
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil || lat < -90 || lat > 90 {
		utils.ErrorResponse(w, "Invalid latitude parameter", 400, http.StatusBadRequest)
		return
	}

	// Validate and parse longitude
	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil || lng < -180 || lng > 180 {
		utils.ErrorResponse(w, "Invalid longitude parameter", 400, http.StatusBadRequest)
		return
	}

	// Parse optional radius
	var radius float64 = 1000 // default 1km
	if radiusStr != "" {
		parsedRadius, err := strconv.ParseFloat(radiusStr, 64)
		if err != nil || parsedRadius <= 0 {
			utils.ErrorResponse(w, "Invalid radius parameter", 400, http.StatusBadRequest)
			return
		}
		radius = parsedRadius
	}

	// Parse optional limit
	var limit int64 = 10 // default 10 results
	if limitStr != "" {
		parsedLimit, err := strconv.ParseInt(limitStr, 10, 64)
		if err != nil || parsedLimit <= 0 {
			utils.ErrorResponse(w, "Invalid limit parameter", 400, http.StatusBadRequest)
			return
		}
		limit = parsedLimit
	}

	// Prepare search request
	searchRequest := models.SearchRequest{
		Latitude:  lat,
		Longitude: lng,
		Keyword:   keyword,
		Radius:    radius,
		Limit:     limit,
		Category:  category,
	}

	// Perform search
	results, err := services.SearchNearbyLocations(searchRequest)
	if err != nil {
		utils.ErrorResponse(w, "Failed to search locations: "+err.Error(), 500, http.StatusInternalServerError)
		return
	}

	// Return response
	utils.SuccessResponse(w, results, http.StatusOK)
}

// Helper function to validate location request
func validateLocationRequest(request models.LocationRequest) error {
	if request.Name == "" {
		return fmt.Errorf("name is required")
	}
	if request.Latitude < -90 || request.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if request.Longitude < -180 || request.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}
	return nil
}
