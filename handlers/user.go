package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"playtime-go/models"
	"playtime-go/services"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// HandleUser handles user creation and retrieval
func HandleUser(w http.ResponseWriter, r *http.Request) {
	// Handle request based on method
	switch r.Method {
	case http.MethodPost:
		createUser(w, r)
	case http.MethodGet:
		getUser(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func createUser(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse request body
	var request models.UserRequest
	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.PhoneNumber == "" {
		http.Error(w, "Phone number is required", http.StatusBadRequest)
		return
	}

	// Call service to create user
	user, err := services.CreateUser(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
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
			http.Error(w, "Invalid user ID format", http.StatusBadRequest)
			return
		}
		user, err = services.GetUserByID(objectID)
	} else if phone != "" {
		user, err = services.GetUserByPhone(phone)
	} else {
		// If neither ID nor phone is provided, return all users
		users, listErr := services.ListUsers()
		if listErr != nil {
			http.Error(w, listErr.Error(), http.StatusInternalServerError)
			return
		}
		// Return response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
		return
	}

	// Handle errors
	if err != nil {
		if strings.Contains(err.Error(), "no documents") || strings.Contains(err.Error(), "no user found") {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// HandleUserByOpenID handles requests to get a user by OpenID
func HandleUserByOpenID(w http.ResponseWriter, r *http.Request) {
	// Only accept GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	openID := strings.TrimPrefix(r.URL.Path, "/user/openid/")
	if openID == "" {
		http.Error(w, "OpenID is required", http.StatusBadRequest)
		return
	}

	// Get user by OpenID
	user, err := services.GetUserByOpenID(openID)
	if err != nil {
		if strings.Contains(err.Error(), "no user found") {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
