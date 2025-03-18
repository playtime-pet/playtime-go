package services

import (
	"fmt"
	"playtime-go/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const petCollection = "pets"

// CreatePet creates a new pet in the database
func CreatePet(request models.PetRequest) (*models.Pet, error) {
	// Create new pet
	now := time.Now()
	pet := models.Pet{
		Name:      request.Name,
		Gender:    request.Gender,
		Size:      request.Size,
		Breed:     request.Breed,
		Avatar:    request.Avatar,
		Desc:      request.Desc,
		Age:       request.Age,
		OwnerID:   request.OwnerID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Insert pet into database
	id, err := InsertOne(petCollection, pet)
	if err != nil {
		return nil, fmt.Errorf("failed to create pet: %v", err)
	}

	pet.ID = id
	return &pet, nil
}

// GetPetByID retrieves a pet by ID
func GetPetByID(id primitive.ObjectID) (*models.Pet, error) {
	filter := bson.M{"_id": id}
	var pet models.Pet

	err := FindOne(petCollection, filter, &pet)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no pet found with ID: %s", id.Hex())
		}
		return nil, fmt.Errorf("failed to get pet by ID: %v", err)
	}

	return &pet, nil
}

// UpdatePet updates an existing pet
func UpdatePet(id primitive.ObjectID, request models.PetRequest) (*models.Pet, error) {
	// Check if pet exists
	_, err := GetPetByID(id)
	if err != nil {
		return nil, err
	}

	// Prepare update document
	now := time.Now()
	updateData := bson.M{
		"$set": bson.M{
			"name":      request.Name,
			"gender":    request.Gender,
			"size":      request.Size,
			"breed":     request.Breed,
			"avatar":    request.Avatar,
			"desc":      request.Desc,
			"age":       request.Age,
			"updatedAt": now,
		},
	}

	// Update pet in the database
	filter := bson.M{"_id": id}
	err = UpdateOne(petCollection, filter, updateData)
	if err != nil {
		return nil, fmt.Errorf("failed to update pet: %v", err)
	}

	// Get the updated pet
	return GetPetByID(id)
}

// DeletePet deletes a pet by ID
func DeletePet(id primitive.ObjectID) error {
	// Check if pet exists
	_, err := GetPetByID(id)
	if err != nil {
		return err
	}

	// Delete pet from the database
	filter := bson.M{"_id": id}
	err = DeleteOne(petCollection, filter)
	if err != nil {
		return fmt.Errorf("failed to delete pet: %v", err)
	}

	return nil
}

// ListPets retrieves all pets with optional pagination and filtering
func ListPets(ownerID *primitive.ObjectID, limit int64) ([]models.Pet, error) {
	// Prepare filter
	filter := bson.M{}
	if ownerID != nil {
		filter["ownerId"] = *ownerID
	}

	var pets []models.Pet

	// Set options for sorting by creation time (descending) and limit
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}})

	// Apply limit if specified
	if limit > 0 {
		findOptions.SetLimit(limit)
	} else {
		findOptions.SetLimit(100) // Default limit
	}

	err := FindMany(petCollection, filter, &pets, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list pets: %v", err)
	}

	return pets, nil
}
