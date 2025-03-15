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

// CreateUser creates a new user from a request
func CreateUser(request models.UserRequest) (*models.User, error) {
	// Check if user with this phone number already exists
	existingUser, err := GetUserByPhone(request.PhoneNumber)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("error checking existing user: %v", err)
	}

	// If user exists, return the existing user
	if existingUser != nil {
		return existingUser, nil
	}

	// Create new user
	now := time.Now()
	user := models.User{
		NickName:    request.NickName,
		PhoneNumber: request.PhoneNumber,
		AvatarURL:   request.AvatarURL,
		OpenID:      request.OpenID,
		UnionID:     request.UnionID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Insert user into database
	id, err := InsertOne(userCollection, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	user.ID = id
	return &user, nil
}

// GetUserByPhone retrieves a user by phone number
func GetUserByPhone(phoneNumber string) (*models.User, error) {
	filter := bson.M{"phoneNumber": phoneNumber}
	var user models.User

	err := FindOne(userCollection, filter, &user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get user by phone: %v", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user by their ObjectID
func GetUserByID(id primitive.ObjectID) (*models.User, error) {
	filter := bson.M{"_id": id}
	var user models.User

	err := FindOne(userCollection, filter, &user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no user found with ID: %s", id.Hex())
		}
		return nil, fmt.Errorf("failed to get user by ID: %v", err)
	}

	return &user, nil
}

// GetUserByOpenID retrieves a user by their OpenID
func GetUserByOpenID(openID string) (*models.User, error) {
	filter := bson.M{"openId": openID}
	var user models.User

	err := FindOne(userCollection, filter, &user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no user found with OpenID: %s", openID)
		}
		return nil, fmt.Errorf("failed to get user by OpenID: %v", err)
	}

	return &user, nil
}

// ListUsers retrieves all users with optional pagination
func ListUsers() ([]models.User, error) {
	filter := bson.M{}
	var users []models.User

	// Set options for sorting by creation time (descending) and limit
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}})
	findOptions.SetLimit(100) // Limiting to 100 users for safety

	err := FindMany(userCollection, filter, &users, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %v", err)
	}

	return users, nil
}
