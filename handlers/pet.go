package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"playtime-go/models"
	"playtime-go/services"
	"playtime-go/utils"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// HandlePet handles pet creation, retrieval, update, and deletion
func HandlePet(w http.ResponseWriter, r *http.Request) {
	// Extract pet ID from URL if present (for specific pet operations)
	urlParts := strings.Split(r.URL.Path, "/")
	var petID string

	// Check if we have a pet ID in the URL
	if len(urlParts) > 2 && urlParts[1] == "pet" && urlParts[2] != "" {
		petID = urlParts[2]
	}

	// Handle request based on method and whether we have a specific pet ID
	switch {
	case r.Method == http.MethodPost && petID == "":
		createPet(w, r)
	case r.Method == http.MethodGet && petID == "":
		listPets(w, r)
	case r.Method == http.MethodGet && petID != "":
		getPet(w, r, petID)
	case r.Method == http.MethodPut && petID != "":
		updatePet(w, r, petID)
	case r.Method == http.MethodDelete && petID != "":
		deletePet(w, r, petID)
	default:
		utils.ErrorResponse(w, "Method not allowed or invalid URL", 405, http.StatusMethodNotAllowed)
	}
}

// createPet handles POST requests to create a new pet
func createPet(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ErrorResponse(w, "Failed to read request body", 400, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse request body
	var request models.PetRequest
	if err := json.Unmarshal(body, &request); err != nil {
		utils.ErrorResponse(w, "Invalid request format", 400, http.StatusBadRequest)
		return
	}

	// Validate request
	if request.Name == "" {
		utils.ErrorResponse(w, "Pet name is required", 400, http.StatusBadRequest)
		return
	}

	// Optional: Add validation for age if needed
	if request.Age <= 0 {
		utils.ErrorResponse(w, "Valid pet age is required", 400, http.StatusBadRequest)
		return
	}

	// Call service to create pet
	pet, err := services.CreatePet(request)
	if err != nil {
		utils.ErrorResponse(w, "Failed to create pet: "+err.Error(), 500, http.StatusInternalServerError)
		return
	}

	// Return response
	utils.SuccessResponse(w, pet, http.StatusCreated)
}

// listPets handles GET requests to list pets with optional filtering
func listPets(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for filtering
	query := r.URL.Query()
	ownerIDStr := query.Get("ownerId")

	var ownerID *primitive.ObjectID

	// Process owner ID filter if present
	if ownerIDStr != "" {
		id, err := primitive.ObjectIDFromHex(ownerIDStr)
		if err != nil {
			utils.ErrorResponse(w, "Invalid owner ID format", 400, http.StatusBadRequest)
			return
		}
		ownerID = &id
	}

	// Get pets from service
	pets, err := services.ListPets(ownerID, 100)
	if err != nil {
		utils.ErrorResponse(w, "Failed to list pets: "+err.Error(), 500, http.StatusInternalServerError)
		return
	}

	// Return response
	utils.SuccessResponse(w, pets, http.StatusOK)
}

// getPet handles GET requests to retrieve a specific pet by ID
func getPet(w http.ResponseWriter, r *http.Request, petID string) {
	// Validate pet ID
	id, err := primitive.ObjectIDFromHex(petID)
	if err != nil {
		utils.ErrorResponse(w, "Invalid pet ID format", 400, http.StatusBadRequest)
		return
	}

	// Get the pet
	pet, err := services.GetPetByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "no pet found") {
			utils.ErrorResponse(w, "Pet not found", 404, http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to get pet: "+err.Error(), 500, http.StatusInternalServerError)
		}
		return
	}

	// Return response
	utils.SuccessResponse(w, pet, http.StatusOK)
}

// updatePet handles PUT requests to update a specific pet
func updatePet(w http.ResponseWriter, r *http.Request, petID string) {
	// Validate pet ID
	id, err := primitive.ObjectIDFromHex(petID)
	if err != nil {
		utils.ErrorResponse(w, "Invalid pet ID format", 400, http.StatusBadRequest)
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
	var request models.PetRequest
	if err := json.Unmarshal(body, &request); err != nil {
		utils.ErrorResponse(w, "Invalid request format", 400, http.StatusBadRequest)
		return
	}

	// Validate request
	if request.Name == "" {
		utils.ErrorResponse(w, "Pet name is required", 400, http.StatusBadRequest)
		return
	}

	// Optional: Add validation for age if needed
	if request.Age <= 0 {
		utils.ErrorResponse(w, "Valid pet age is required", 400, http.StatusBadRequest)
		return
	}

	// Update the pet
	pet, err := services.UpdatePet(id, request)
	if err != nil {
		if strings.Contains(err.Error(), "no pet found") {
			utils.ErrorResponse(w, "Pet not found", 404, http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to update pet: "+err.Error(), 500, http.StatusInternalServerError)
		}
		return
	}

	// Return response
	utils.SuccessResponse(w, pet, http.StatusOK)
}

// deletePet handles DELETE requests to remove a pet
func deletePet(w http.ResponseWriter, r *http.Request, petID string) {
	// Validate pet ID
	id, err := primitive.ObjectIDFromHex(petID)
	if err != nil {
		utils.ErrorResponse(w, "Invalid pet ID format", 400, http.StatusBadRequest)
		return
	}

	// Delete the pet
	err = services.DeletePet(id)
	if err != nil {
		if strings.Contains(err.Error(), "no pet found") {
			utils.ErrorResponse(w, "Pet not found", 404, http.StatusNotFound)
		} else {
			utils.ErrorResponse(w, "Failed to delete pet: "+err.Error(), 500, http.StatusInternalServerError)
		}
		return
	}

	// Return success response
	utils.SuccessResponse(w, map[string]string{"message": "Pet deleted successfully"}, http.StatusOK)
}
