package services

import (
	"context"
	"fmt"
	"playtime-go/db"
	"playtime-go/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	userCollection = "users"
	timeout        = 10 * time.Second
)

// CreateUser creates a new user in the database
func CreateUser(request models.UserRequest) (*models.User, error) {
	collection := db.GetCollection(userCollection)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

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
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Insert user into database
	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// Get the inserted ID and set it on the user
	if objectID, ok := result.InsertedID.(interface{}); ok {
		user.ID = objectID.(primitive.ObjectID)
	}

	return &user, nil
}

// GetUserByPhone retrieves a user by phone number
func GetUserByPhone(phoneNumber string) (*models.User, error) {
	collection := db.GetCollection(userCollection)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"phoneNumber": phoneNumber}).Decode(&user)
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
	collection := db.GetCollection(userCollection)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no user found with ID: %s", id.Hex())
		}
		return nil, fmt.Errorf("failed to get user by ID: %v", err)
	}

	return &user, nil
}

// ListUsers retrieves all users with optional pagination
func ListUsers() ([]models.User, error) {
	collection := db.GetCollection(userCollection)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Options for sorting by creation time (descending)
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}})
	findOptions.SetLimit(100) // Limiting to 100 users for safety

	cursor, err := collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %v", err)
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %v", err)
	}

	return users, nil
}
