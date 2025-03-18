package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"playtime-go/models"
	"playtime-go/services"
	"playtime-go/utils"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// HandleUser handles user creation and retrieval
func HandleUser(w http.ResponseWriter, r *http.Request) {
	// Handle request based on method
	switch r.Method {
	case http.MethodPost:
		upsertUser(w, r)
	case http.MethodGet:
		getUser(w, r)
	default:
		utils.ErrorResponse(w, "Method not allowed", 400, http.StatusBadRequest)
	}
}

func upsertUser(w http.ResponseWriter, r *http.Request) {
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

	var user *models.User
	var isNewUser bool = true

	// Check if user exists with the given OpenID
	existingUser, err := services.GetUserByOpenID(request.OpenID)
	if err == nil && existingUser != nil {
		// User exists, update their information
		log.Printf("User found with OpenID %s, updating user", request.OpenID)

		// Prepare update document
		now := time.Now()
		updateData := bson.M{
			"$set": bson.M{
				"nickName":    request.NickName,
				"phoneNumber": request.PhoneNumber,
				"avatarUrl":   request.AvatarURL,
				"unionId":     request.UnionID,
				"updatedAt":   now,
			},
		}

		// Update user in the database - Fix: Use the constant from services package
		filter := bson.M{"openId": request.OpenID}
		err = services.UpdateOne("users", filter, updateData)
		if err != nil {
			utils.ErrorResponse(w, "Failed to update user: "+err.Error(), 500, http.StatusInternalServerError)
			return
		}

		// Get the updated user
		user, err = services.GetUserByOpenID(request.OpenID)
		if err != nil {
			utils.ErrorResponse(w, "Failed to retrieve updated user: "+err.Error(), 500, http.StatusInternalServerError)
			return
		}

		isNewUser = false
	} else {
		// User doesn't exist, create a new one
		log.Printf("No user found with OpenID %s, creating new user", request.OpenID)
		user, err = services.CreateUser(request)
		if err != nil {
			utils.ErrorResponse(w, "Failed to create user: "+err.Error(), 500, http.StatusInternalServerError)
			return
		}
		isNewUser = true
	}

	// Return success response with appropriate status code
	statusCode := http.StatusOK
	if isNewUser {
		statusCode = http.StatusCreated
	}
	utils.SuccessResponse(w, user, statusCode)
}

// getUser handles GET requests to retrieve user information
func getUser(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	id := query.Get("id")
	phone := query.Get("phone")

	var (
		user interface{}
		err  error
	)

	// Based on provided parameters, fetch user
	if id != "" {
		// Convert string ID to ObjectID
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			utils.ErrorResponse(w, "Invalid user ID format", 400, http.StatusBadRequest)
			return
		}
		user, _ = services.GetUserByID(objectID)
		// if err != nil {
		// 	utils.ErrorResponse(w, err.Error(), 500, http.StatusInternalServerError)
		// 	return
		// }
	} else if phone != "" {
		user, err = services.GetUserByPhone(phone)
	} else {
		// If neither ID nor phone is provided, return all users
		users, listErr := services.ListUsers()
		if listErr != nil {
			utils.ErrorResponse(w, listErr.Error(), 500, http.StatusInternalServerError)
			return
		}
		// Return response
		utils.SuccessResponse(w, users, http.StatusOK)
		return
	}

	// Handle errors
	if err != nil {
		if strings.Contains(err.Error(), "no documents") || strings.Contains(err.Error(), "no user found") {
			utils.ErrorResponse(w, "User not found", 404, http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, err.Error(), 500, http.StatusInternalServerError)
		}
		return
	}

	// Return response
	utils.SuccessResponse(w, user, http.StatusOK)
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
