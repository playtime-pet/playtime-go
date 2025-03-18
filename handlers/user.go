package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"playtime-go/models"
	"playtime-go/services"
	"playtime-go/utils"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// HandleUser handles user creation, retrieval, update, and deletion
func HandleUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL if present (for specific user operations)
	urlParts := strings.Split(r.URL.Path, "/")
	var userID string

	// Check if we have a user ID in the URL
	if len(urlParts) > 2 && urlParts[1] == "user" && urlParts[2] != "" {
		userID = urlParts[2]
	}

	// Handle request based on method and whether we have a specific user ID
	switch {
	case r.Method == http.MethodPost && userID == "":
		createUser(w, r)
	case r.Method == http.MethodGet && userID == "":
		listUsers(w, r)
	case r.Method == http.MethodGet && userID != "":
		getUser(w, r, userID)
	case r.Method == http.MethodPut && userID != "":
		updateUser(w, r, userID)
	case r.Method == http.MethodDelete && userID != "":
		deleteUser(w, r, userID)
	default:
		utils.ErrorResponse(w, "Method not allowed or invalid URL", 405, http.StatusMethodNotAllowed)
	}
}

// createUser handles POST requests to create a new user
func createUser(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ErrorResponse(w, "Failed to read request body", 400, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse request body
	var request models.UserRequest
	if err := json.Unmarshal(body, &request); err != nil {
		utils.ErrorResponse(w, "Invalid request format", 400, http.StatusBadRequest)
		return
	}

	// Validate request
	if request.PhoneNumber == "" {
		utils.ErrorResponse(w, "Phone number is required", 400, http.StatusBadRequest)
		return
	}

	if request.OpenID == "" {
		utils.ErrorResponse(w, "OpenID is required", 400, http.StatusBadRequest)
		return
	}

	// Check if user already exists with the given OpenID
	existingUser, err := services.GetUserByOpenID(request.OpenID)
	if err == nil && existingUser != nil {
		utils.ErrorResponse(w, "User with this OpenID already exists", 409, http.StatusConflict)
		return
	}

	// Call service to create user
	user, err := services.CreateUser(request)
	if err != nil {
		utils.ErrorResponse(w, "Failed to create user: "+err.Error(), 500, http.StatusInternalServerError)
		return
	}

	// Return response
	utils.SuccessResponse(w, user, http.StatusCreated)
}

// listUsers handles GET requests to list users with optional filtering
func listUsers(w http.ResponseWriter, r *http.Request) {
	// Get users from service
	users, err := services.ListUsers()
	if err != nil {
		utils.ErrorResponse(w, "Failed to list users: "+err.Error(), 500, http.StatusInternalServerError)
		return
	}

	// Return response
	utils.SuccessResponse(w, users, http.StatusOK)
}

// getUser handles GET requests to retrieve a specific user
func getUser(w http.ResponseWriter, r *http.Request, userID string) {
	// Validate user ID
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.ErrorResponse(w, "Invalid user ID format", 400, http.StatusBadRequest)
		return
	}

	// Get the user
	user, err := services.GetUserByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "no user found") {
			utils.ErrorResponse(w, "User not found", 404, http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to get user: "+err.Error(), 500, http.StatusInternalServerError)
		}
		return
	}

	// Return response
	utils.SuccessResponse(w, user, http.StatusOK)
}

// updateUser handles PUT requests to update a specific user
func updateUser(w http.ResponseWriter, r *http.Request, userID string) {
	// Validate user ID
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.ErrorResponse(w, "Invalid user ID format", 400, http.StatusBadRequest)
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
	var request models.UserRequest
	if err := json.Unmarshal(body, &request); err != nil {
		utils.ErrorResponse(w, "Invalid request format", 400, http.StatusBadRequest)
		return
	}

	// Validate request
	if request.PhoneNumber == "" {
		utils.ErrorResponse(w, "Phone number is required", 400, http.StatusBadRequest)
		return
	}

	// Prepare update document
	now := time.Now()
	updateData := bson.M{
		"$set": bson.M{
			"nickName":    request.NickName,
			"phoneNumber": request.PhoneNumber,
			"avatarUrl":   request.AvatarURL,
			"openId":      request.OpenID,
			"unionId":     request.UnionID,
			"updatedAt":   now,
		},
	}

	// Update user in the database
	filter := bson.M{"_id": id}
	err = services.UpdateOne("users", filter, updateData)
	if err != nil {
		utils.ErrorResponse(w, "Failed to update user: "+err.Error(), 500, http.StatusInternalServerError)
		return
	}

	// Get the updated user
	user, err := services.GetUserByID(id)
	if err != nil {
		utils.ErrorResponse(w, "User updated but failed to retrieve: "+err.Error(), 500, http.StatusInternalServerError)
		return
	}

	// Return response
	utils.SuccessResponse(w, user, http.StatusOK)
}

// deleteUser handles DELETE requests to remove a user
func deleteUser(w http.ResponseWriter, r *http.Request, userID string) {
	// Validate user ID
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.ErrorResponse(w, "Invalid user ID format", 400, http.StatusBadRequest)
		return
	}

	// Check if user exists
	_, err = services.GetUserByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "no user found") {
			utils.ErrorResponse(w, "User not found", 404, http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to find user: "+err.Error(), 500, http.StatusInternalServerError)
		}
		return
	}

	// Delete user from the database
	filter := bson.M{"_id": id}
	err = services.DeleteOne("users", filter)
	if err != nil {
		utils.ErrorResponse(w, "Failed to delete user: "+err.Error(), 500, http.StatusInternalServerError)
		return
	}

	// Return success response
	utils.SuccessResponse(w, map[string]string{"message": "User deleted successfully"}, http.StatusOK)
}

// HandleUserByOpenID handles requests to get a user by OpenID
func HandleUserByOpenID(w http.ResponseWriter, r *http.Request) {
	// Only accept GET requests
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", 405, http.StatusMethodNotAllowed)
		return
	}

	openID := strings.TrimPrefix(r.URL.Path, "/user/openid/")
	if openID == "" {
		utils.ErrorResponse(w, "OpenID is required", 400, http.StatusBadRequest)
		return
	}

	// Get user by OpenID
	user, err := services.GetUserByOpenID(openID)
	if err != nil {
		if strings.Contains(err.Error(), "no user found") {
			utils.ErrorResponse(w, "User not found", 404, http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, err.Error(), 500, http.StatusInternalServerError)
		}
		return
	}

	// Return response
	utils.SuccessResponse(w, user, http.StatusOK)
}
